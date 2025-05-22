use std::env;
use dotenv::dotenv;
use teloxide::{prelude::*, dptree};
use teloxide::dispatching::UpdateFilterExt;
use teloxide::types::{Update, InlineKeyboardButton, InlineKeyboardMarkup, CallbackQuery};
use rdkafka::config::ClientConfig;
use rdkafka::producer::{FutureProducer, FutureRecord};
use rdkafka::consumer::{StreamConsumer, Consumer};
use rdkafka::message::Message as KafkaMessageRd; // Renamed to avoid conflict
use serde::{Serialize, Deserialize};
use futures_util::StreamExt;
use std::sync::Arc;
use std::error::Error;

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
    let json = serde_json::to_string(&msg)?;
    let record = FutureRecord::to(kafka_in_topic.as_str())
        .payload(&json)
        .key("message");
    producer.send(record, None).await.map_err(|(e, _)| Box::new(e) as Box<dyn Error + Send + Sync>)?;
    Ok(())
}

async fn callback_query_handler(
    bot: Bot,
    query: CallbackQuery,
    producer: Arc<FutureProducer>,
    kafka_in_topic: Arc<String>,
) -> Result<(), Box<dyn Error + Send + Sync>> {
    bot.answer_callback_query(query.id.clone()).await?;

    let chat_id = query.message.as_ref().map_or(0, |m| m.chat.id.0);
    let user_id = query.from.id.0;
    let message_id = query.message.as_ref().map_or(0, |m| m.id.0);
    let callback_data = query.data.unwrap_or_default();
    let callback_query_id = query.id.clone();

    let incoming_callback_message = IncomingCallbackMessage {
        chat_id,
        user_id,
        message_id,
        callback_data,
        callback_query_id,
    };

    let json = serde_json::to_string(&incoming_callback_message)?;
    let record = FutureRecord::to(kafka_in_topic.as_str())
        .payload(&json)
        .key("callback_query"); // Different key for callback queries

    producer.send(record, None).await.map_err(|(e, _)| Box::new(e) as Box<dyn Error + Send + Sync>)?;

    Ok(())
}

// Public function to create InlineKeyboardMarkup from ButtonInfo
pub fn create_markup(buttons_opt: &Option<Vec<Vec<ButtonInfo>>>) -> Option<InlineKeyboardMarkup> {
    buttons_opt.as_ref().map(|buttons| {
        InlineKeyboardMarkup::new(buttons.iter().map(|row| {
            row.iter().map(|button_info| {
                InlineKeyboardButton::callback(button_info.text.clone(), button_info.callback_data.clone())
            })
        }))
    })
}

// Public function to prepare IncomingCallbackMessage from CallbackQuery
pub fn prepare_incoming_callback_message(query: &CallbackQuery) -> IncomingCallbackMessage {
    let chat_id = query.message.as_ref().map_or(0, |m| m.chat.id.0);
    let user_id = query.from.id.0;
    let message_id = query.message.as_ref().map_or(0, |m| m.id.0);
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
    pretty_env_logger::init(); // For better logging
    log::info!("Starting Ratatoskr bot...");

    let telegram_token = env::var("TELEGRAM_BOT_TOKEN").expect("TELEGRAM_BOT_TOKEN not set");
    let kafka_broker = env::var("KAFKA_BROKER").unwrap_or_else(|_| "localhost:9092".to_string());
    let kafka_in_topic = Arc::new(env::var("KAFKA_IN_TOPIC").unwrap_or_else(|_| "com.sectorflabs.ratatoskr.in".to_string()));
    let kafka_out_topic = env::var("KAFKA_OUT_TOPIC").unwrap_or_else(|_| "com.sectorflabs.ratatoskr.out".to_string());

    let bot = Bot::new(telegram_token.clone());

    // Kafka producer
    let producer: Arc<FutureProducer> = Arc::new(ClientConfig::new()
        .set("bootstrap.servers", &kafka_broker)
        .create()
        .expect("Producer creation error"));

    // Kafka consumer
    let consumer: StreamConsumer = ClientConfig::new()
        .set("group.id", "ratatoskr-bot-consumer")
        .set("bootstrap.servers", &kafka_broker)
        .set("enable.partition.eof", "false")
        .set("session.timeout.ms", "6000")
        .set("enable.auto.commit", "true")
        .create()
        .expect("Consumer creation failed");
    consumer.subscribe(&[&kafka_out_topic]).expect("Can't subscribe to topic");

    // Kafka -> Telegram task
    let bot_consumer_clone = bot.clone();
    tokio::spawn(async move {
        let mut stream = consumer.stream();
        while let Some(Ok(kafka_msg)) = stream.next().await { // Renamed msg to kafka_msg
            if let Some(payload) = kafka_msg.payload() {
                match serde_json::from_slice::<OutgoingKafkaMessage>(payload) {
                    Ok(out_msg) => {
                        let mut msg_to_send = bot_consumer_clone.send_message(ChatId(out_msg.chat_id), out_msg.text);
                        if let Some(markup) = create_markup(&out_msg.buttons) {
                            msg_to_send = msg_to_send.reply_markup(markup);
                        }
                        if let Err(e) = msg_to_send.await {
                            log::error!("Error sending message to Telegram: {:?}", e);
                        }
                    }
                    Err(e) => {
                        log::error!("Error deserializing OutgoingKafkaMessage: {:?}", e);
                        log::debug!("Problematic payload: {:?}", String::from_utf8_lossy(payload));
                    }
                }
            }
        }
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

#[cfg(test)]
mod tests {
    use super::*;
    use teloxide::types::{
        User, Chat, ChatKind, MessageKind, MessageCommon, MessageId, ChatId, UserId,
        PrivateChat, MessageText,
    };
    use std::time::{SystemTime, UNIX_EPOCH};

    // Helper to create a basic ButtonInfo
    fn bi(text: &str, cbd: &str) -> ButtonInfo {
        ButtonInfo { text: text.to_string(), callback_data: cbd.to_string() }
    }

    #[test]
    fn test_create_markup_none() {
        assert_eq!(create_markup(&None), None);
    }

    #[test]
    fn test_create_markup_empty_buttons() {
        let buttons: Option<Vec<Vec<ButtonInfo>>> = Some(vec![]);
        let expected_markup = InlineKeyboardMarkup::new(Vec::<Vec<InlineKeyboardButton>>::new());
        assert_eq!(create_markup(&buttons), Some(expected_markup));
    }

    #[test]
    fn test_create_markup_empty_row() {
        let buttons: Option<Vec<Vec<ButtonInfo>>> = Some(vec![vec![]]);
        let expected_markup = InlineKeyboardMarkup::new(vec![Vec::<InlineKeyboardButton>::new()]);
        assert_eq!(create_markup(&buttons), Some(expected_markup));
    }

    #[test]
    fn test_create_markup_single_button() {
        let buttons = Some(vec![vec![bi("Test", "cb_test")]]);
        let expected_button = InlineKeyboardButton::callback("Test".to_string(), "cb_test".to_string());
        let expected_markup = InlineKeyboardMarkup::new(vec![vec![expected_button]]);
        assert_eq!(create_markup(&buttons), Some(expected_markup));
    }

    #[test]
    fn test_create_markup_multiple_buttons_one_row() {
        let buttons = Some(vec![vec![
            bi("B1", "cb1"),
            bi("B2", "cb2"),
        ]]);
        let expected_buttons = vec![
            InlineKeyboardButton::callback("B1".to_string(), "cb1".to_string()),
            InlineKeyboardButton::callback("B2".to_string(), "cb2".to_string()),
        ];
        let expected_markup = InlineKeyboardMarkup::new(vec![expected_buttons]);
        assert_eq!(create_markup(&buttons), Some(expected_markup));
    }

    #[test]
    fn test_create_markup_multiple_rows() {
        let buttons = Some(vec![
            vec![bi("R1B1", "cb_r1b1")],
            vec![bi("R2B1", "cb_r2b1"), bi("R2B2", "cb_r2b2")],
        ]);
        let expected_markup = InlineKeyboardMarkup::new(vec![
            vec![InlineKeyboardButton::callback("R1B1".to_string(), "cb_r1b1".to_string())],
            vec![
                InlineKeyboardButton::callback("R2B1".to_string(), "cb_r2b1".to_string()),
                InlineKeyboardButton::callback("R2B2".to_string(), "cb_r2b2".to_string()),
            ],
        ]);
        assert_eq!(create_markup(&buttons), Some(expected_markup));
    }

    // Helper to create a mock CallbackQuery
    fn mock_callback_query(
        query_id: &str,
        user_id: u64,
        chat_id: i64,
        msg_id: i32,
        data: Option<String>,
        message_present: bool,
    ) -> CallbackQuery {
        let user = User {
            id: UserId(user_id),
            is_bot: false,
            first_name: "Test".to_string(),
            last_name: None,
            username: Some("testuser".to_string()),
            language_code: Some("en".to_string()),
            is_premium: false,
            added_to_attachment_menu: false,
        };

        let message = if message_present {
            Some(Arc::new(Message {
                id: MessageId(msg_id),
                date: SystemTime::now().duration_since(UNIX_EPOCH).unwrap().as_secs() as i64, // u64 in teloxide 0.12, cast to i64
                chat: Chat {
                    id: ChatId(chat_id),
                    kind: ChatKind::Private(PrivateChat {
                        username: Some("testuser".to_string()),
                        first_name: Some("Test".to_string()),
                        last_name: None,
                        bio: None,
                        has_private_forwards: None,
                        has_restricted_voice_and_video_messages: None,
                        photo: None,
                    }),
                    photo: None,
                    pinned_message: None, // Box<Option<Message>>
                    message_auto_delete_time: None,
                },
                kind: MessageKind::Common(MessageCommon{
                    from: Some(user.clone()),
                     sender_chat: None, // Option<Chat>
                     author_signature: None, // Option<String>
                     reply_to_message: None, // Box<Option<Message>>
                     edit_date: None, // Option<i32>
                     forward: None, // Option<MessageForward>
                     via_bot: None, // Option<User>
                }),
                 // Other fields like `via_bot`, `edit_date`, etc. are not directly used by the handler
                // but might be needed for full Message construction.
                // For this test, we only need what `prepare_incoming_callback_message` uses.
                // Let's assume a simple text message for now.
                // kind: MessageKind::Text(MessageText { text: "Hello".to_string(), entities: vec![] }),
            }))
        } else {
            None
        };

        CallbackQuery {
            id: query_id.to_string(),
            from: user,
            message,
            inline_message_id: None,
            chat_instance: "instance1".to_string(),
            data,
            game_short_name: None,
        }
    }

    #[test]
    fn test_prepare_incoming_callback_message_with_message() {
        let query = mock_callback_query("q1", 123, 456, 789, Some("cb_data_1".to_string()), true);
        let result = prepare_incoming_callback_message(&query);

        assert_eq!(result.chat_id, 456);
        assert_eq!(result.user_id, 123);
        assert_eq!(result.message_id, 789);
        assert_eq!(result.callback_data, "cb_data_1");
        assert_eq!(result.callback_query_id, "q1");
    }

    #[test]
    fn test_prepare_incoming_callback_message_without_message() {
        // This tests the .map_or(0, ...) default for chat_id and message_id
        let query = mock_callback_query("q2", 234, 0, 0, Some("cb_data_2".to_string()), false);
        let result = prepare_incoming_callback_message(&query);

        assert_eq!(result.chat_id, 0); // Defaulted
        assert_eq!(result.user_id, 234);
        assert_eq!(result.message_id, 0); // Defaulted
        assert_eq!(result.callback_data, "cb_data_2");
        assert_eq!(result.callback_query_id, "q2");
    }

    #[test]
    fn test_prepare_incoming_callback_message_no_data() {
        let query = mock_callback_query("q3", 345, 678, 901, None, true);
        let result = prepare_incoming_callback_message(&query);

        assert_eq!(result.chat_id, 678);
        assert_eq!(result.user_id, 345);
        assert_eq!(result.message_id, 901);
        assert_eq!(result.callback_data, ""); // Defaulted to empty string
        assert_eq!(result.callback_query_id, "q3");
    }
}
