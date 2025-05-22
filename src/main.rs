use dotenv::dotenv;
use futures_util::StreamExt;
use rdkafka::config::ClientConfig;
use rdkafka::consumer::{Consumer, StreamConsumer};
use rdkafka::message::Message as KafkaMessageRd; // Renamed to avoid conflict
use rdkafka::producer::{FutureProducer, FutureRecord};
use serde::{Deserialize, Serialize};
use std::env;
use std::error::Error;
use std::sync::Arc;
use teloxide::dispatching::UpdateFilterExt;
use teloxide::types::{CallbackQuery, InlineKeyboardButton, InlineKeyboardMarkup, Update};
use teloxide::{dptree, prelude::*};
use tracing_subscriber::{EnvFilter, fmt};

#[derive(Serialize, Deserialize, Debug)]
struct ButtonInfo {
    text: String,
    callback_data: String,
}

#[derive(Serialize, Deserialize, Debug)]
struct OutgoingKafkaMessage {
    chat_id: i64,
    text: String,
    buttons: Option<Vec<Vec<ButtonInfo>>>,
}

#[derive(Serialize, Deserialize, Debug)]
pub struct IncomingCallbackMessage {
    chat_id: i64,
    user_id: u64,
    message_id: i32,
    callback_data: String,
    callback_query_id: String,
}

async fn message_handler(
    bot: Bot,
    msg: Message,
    producer: Arc<FutureProducer>,
    kafka_in_topic: Arc<String>,
) -> Result<(), Box<dyn Error + Send + Sync>> {
    tracing::debug!(message_id = %msg.id.0, chat_id = %msg.chat.id.0, user_id = ?msg.from().map(|u| u.id.0), "Received Telegram message");

    let json = match serde_json::to_string(&msg) {
        Ok(json_string) => json_string,
        Err(e) => {
            tracing::error!(message_id = %msg.id.0, chat_id = %msg.chat.id.0, error = %e, "Failed to serialize Telegram message to JSON");
            return Err(Box::new(e));
        }
    };

    tracing::info!(topic = %kafka_in_topic, key = "message", message_id = %msg.id.0, chat_id = %msg.chat.id.0, "Sending Telegram message to Kafka");
    let record = FutureRecord::to(kafka_in_topic.as_str())
        .payload(&json)
        .key("message");

    if let Err((e, _)) = producer.send(record, None).await {
        tracing::error!(topic = %kafka_in_topic, key = "message", message_id = %msg.id.0, chat_id = %msg.chat.id.0, error = %e, "Failed to send message to Kafka");
        return Err(Box::new(e));
    }
    Ok(())
}

async fn callback_query_handler(
    bot: Bot,
    query: CallbackQuery,
    producer: Arc<FutureProducer>,
    kafka_in_topic: Arc<String>,
) -> Result<(), Box<dyn Error + Send + Sync>> {
    let user_id = query.from.id.0;
    let query_id = query.id.clone();
    let data = query.data.as_deref().unwrap_or_default();
    let message_id = query.message.as_ref().map(|m| m.id().0); // Changed: m.id.0 -> m.id().0

    tracing::debug!(callback_query_id = %query_id, %user_id, message_id = ?message_id, callback_data = %data, "Received callback query");

    if let Err(e) = bot.answer_callback_query(query.id.clone()).await {
        tracing::warn!(callback_query_id = %query_id, user_id = %user_id, error = %e, "Failed to answer callback query");
        // Continue processing even if answering fails, as the main goal is to get it to Kafka
    }

    let incoming_msg = prepare_incoming_callback_message(&query);

    let json = match serde_json::to_string(&incoming_msg) {
        Ok(json_string) => json_string,
        Err(e) => {
            tracing::error!(callback_query_id = %query_id, user_id = %user_id, error = %e, "Failed to serialize IncomingCallbackMessage to JSON");
            return Err(Box::new(e));
        }
    };

    tracing::info!(topic = %kafka_in_topic, key = "callback_query", callback_query_id = %query_id, user_id = %user_id, "Sending callback data to Kafka");
    let record = FutureRecord::to(kafka_in_topic.as_str())
        .payload(&json)
        .key("callback_query");

    if let Err((e, _)) = producer.send(record, None).await {
        tracing::error!(topic = %kafka_in_topic, key = "callback_query", callback_query_id = %query_id, user_id = %user_id, error = %e, "Failed to send callback data to Kafka");
        return Err(Box::new(e));
    }

    Ok(())
}

// Public function to create InlineKeyboardMarkup from ButtonInfo
pub fn create_markup(buttons_opt: &Option<Vec<Vec<ButtonInfo>>>) -> Option<InlineKeyboardMarkup> {
    buttons_opt.as_ref().map(|buttons| {
        InlineKeyboardMarkup::new(buttons.iter().map(|row| {
            row.iter().map(|button_info| {
                InlineKeyboardButton::callback(
                    button_info.text.clone(),
                    button_info.callback_data.clone(),
                )
            })
        }))
    })
}

// Public function to prepare IncomingCallbackMessage from CallbackQuery
pub fn prepare_incoming_callback_message(query: &CallbackQuery) -> IncomingCallbackMessage {
    let chat_id = query.message.as_ref().map_or(0, |m| m.chat().id.0); // Changed: m.chat.id.0 -> m.chat().id.0
    let user_id = query.from.id.0;
    let message_id = query.message.as_ref().map_or(0, |m| m.id().0); // Changed: m.id.0 -> m.id().0
    let callback_data = query.data.clone().unwrap_or_default();
    let callback_query_id = query.id.clone();

    IncomingCallbackMessage {
        chat_id,
        user_id,
        message_id,
        callback_data,
        callback_query_id,
    }
}

#[tokio::main]
async fn main() {
    dotenv().ok();

    // Initialize tracing subscriber
    // RUST_LOG environment variable can be used to control log levels (e.g., RUST_LOG=info,ratatoskr=debug)
    let subscriber = fmt::Subscriber::builder()
        .with_env_filter(EnvFilter::from_default_env())
        .with_writer(std::io::stderr) // Or std::io::stdout
        .finish();
    if let Err(e) = tracing::subscriber::set_global_default(subscriber) {
        // Fallback to basic logging if tracing setup fails, though it's unlikely.
        eprintln!("Failed to set global default tracing subscriber: {}", e);
    }

    tracing::info!("Starting Ratatoskr bot...");

    let telegram_token =
        env::var("TELEGRAM_BOT_TOKEN").expect("FATAL: TELEGRAM_BOT_TOKEN not set in environment");

    let kafka_broker = env::var("KAFKA_BROKER").unwrap_or_else(|_| {
        tracing::info!("KAFKA_BROKER not set, defaulting to localhost:9092");
        "localhost:9092".to_string()
    });
    tracing::info!(kafka_broker = %kafka_broker, "Using Kafka broker");

    let kafka_in_topic_val = env::var("KAFKA_IN_TOPIC").unwrap_or_else(|_| {
        tracing::info!("KAFKA_IN_TOPIC not set, defaulting to com.sectorflabs.ratatoskr.in");
        "com.sectorflabs.ratatoskr.in".to_string()
    });
    let kafka_in_topic = Arc::new(kafka_in_topic_val.clone());
    tracing::info!(kafka_in_topic = %kafka_in_topic, "Using Kafka IN topic");

    let kafka_out_topic = env::var("KAFKA_OUT_TOPIC").unwrap_or_else(|_| {
        tracing::info!("KAFKA_OUT_TOPIC not set, defaulting to com.sectorflabs.ratatoskr.out");
        "com.sectorflabs.ratatoskr.out".to_string()
    });
    tracing::info!(kafka_out_topic = %kafka_out_topic, "Using Kafka OUT topic");

    let bot = Bot::new(telegram_token.clone());

    // Kafka producer
    let producer: Arc<FutureProducer> = Arc::new(
        ClientConfig::new()
            .set("bootstrap.servers", &kafka_broker)
            .create()
            .unwrap_or_else(|e| {
                tracing::error!(error = %e, "Kafka producer creation error");
                panic!("Kafka producer creation error: {}", e);
            }),
    );
    tracing::info!("Kafka producer created successfully.");

    // Kafka consumer
    let consumer: StreamConsumer = ClientConfig::new()
        .set("group.id", "ratatoskr-bot-consumer")
        .set("bootstrap.servers", &kafka_broker)
        .set("enable.partition.eof", "false")
        .set("session.timeout.ms", "6000")
        .set("enable.auto.commit", "true")
        .create()
        .unwrap_or_else(|e| {
            tracing::error!(error = %e, "Kafka consumer creation error");
            panic!("Kafka consumer creation error: {}", e);
        });
    tracing::info!("Kafka consumer created successfully.");

    consumer.subscribe(&[&kafka_out_topic]).unwrap_or_else(|e| {
        tracing::error!(topic = %kafka_out_topic, error = %e, "Failed to subscribe to Kafka topic");
        panic!(
            "Failed to subscribe to Kafka topic {}: {}",
            kafka_out_topic, e
        );
    });
    tracing::info!(topic = %kafka_out_topic, "Subscribed to Kafka topic successfully.");

    // Kafka -> Telegram task
    let bot_consumer_clone = bot.clone();
    let kafka_out_topic_clone = kafka_out_topic.clone(); // Clone for use in the spawned task
    tokio::spawn(async move {
        tracing::info!(topic = %kafka_out_topic_clone, "Starting Kafka consumer stream for Telegram output...");
        let mut stream = consumer.stream();
        while let Some(result) = stream.next().await {
            match result {
                Ok(kafka_msg) => {
                    tracing::debug!(topic = %kafka_msg.topic(), partition = %kafka_msg.partition(), offset = %kafka_msg.offset(), "Consumed message from Kafka");
                    if let Some(payload) = kafka_msg.payload() {
                        match serde_json::from_slice::<OutgoingKafkaMessage>(payload) {
                            Ok(out_msg) => {
                                let chat_id = ChatId(out_msg.chat_id);
                                tracing::info!(%chat_id, text_length = %out_msg.text.len(), has_buttons = %out_msg.buttons.is_some(), "Sending message to Telegram");
                                let mut msg_to_send =
                                    bot_consumer_clone.send_message(chat_id, out_msg.text.clone());
                                if let Some(markup) = create_markup(&out_msg.buttons) {
                                    msg_to_send = msg_to_send.reply_markup(markup);
                                }
                                if let Err(e) = msg_to_send.await {
                                    tracing::error!(%chat_id, error = ?e, "Error sending message to Telegram");
                                }
                            }
                            Err(e) => {
                                tracing::error!(topic = %kafka_msg.topic(), error = %e, "Error deserializing OutgoingKafkaMessage from Kafka payload");
                                tracing::debug!(raw_payload = ?String::from_utf8_lossy(payload), "Problematic Kafka payload");
                            }
                        }
                    } else {
                        tracing::warn!(topic = %kafka_msg.topic(), partition = %kafka_msg.partition(), offset = %kafka_msg.offset(), "Received Kafka message with empty payload");
                    }
                }
                Err(e) => {
                    tracing::error!(topic = %kafka_out_topic_clone, error = %e, "Error consuming message from Kafka");
                    // Depending on the error, may need to break or re-initialize consumer.
                    // For now, we just log and continue.
                }
            }
        }
        tracing::warn!(topic = %kafka_out_topic_clone, "Kafka consumer stream ended.");
    });

    // Telegram -> Kafka dispatcher
    let handler = dptree::entry()
        .branch(Update::filter_message().endpoint(message_handler))
        .branch(Update::filter_callback_query().endpoint(callback_query_handler));

    Dispatcher::builder(bot, handler)
        .dependencies(dptree::deps![producer, kafka_in_topic])
        .enable_ctrlc_handler()
        .build()
        .dispatch()
        .await;
}
