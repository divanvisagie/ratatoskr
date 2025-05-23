use rdkafka::producer::{FutureProducer, FutureRecord};
use std::error::Error;
use std::sync::Arc;
use teloxide::prelude::{Bot, CallbackQuery, Message, Requester};

pub async fn message_handler(
    _bot: Bot,
    msg: Message,
    producer: Arc<FutureProducer>,
    kafka_in_topic: Arc<String>,
) -> Result<(), Box<dyn Error + Send + Sync>> {
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

pub async fn callback_query_handler(
    bot: Bot,
    query: CallbackQuery,
    producer: Arc<FutureProducer>,
    kafka_in_topic: Arc<String>,
) -> Result<(), Box<dyn Error + Send + Sync>> {
    let user_id = query.from.id.0;
    let query_id = query.id.clone();
    let data = query.data.as_deref().unwrap_or_default();
    let message_id = query.message.as_ref().map(|m| m.id().0);

    tracing::debug!(callback_query_id = %query_id, %user_id, message_id = ?message_id, callback_data = %data, "Received callback query");

    if let Err(e) = bot.answer_callback_query(query.id.clone()).await {
        tracing::warn!(callback_query_id = %query_id, user_id = %user_id, error = %e, "Failed to answer callback query");
    }

    let incoming_msg = crate::utils::prepare_incoming_callback_message(&query);

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
