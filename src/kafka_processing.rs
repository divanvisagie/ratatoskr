use crate::structs::OutgoingKafkaMessage;
use crate::utils::create_markup;
use futures_util::StreamExt;
use rdkafka::consumer::StreamConsumer; // Consumer might not be needed directly if StreamConsumer is passed
use rdkafka::message::Message as KafkaMessageRd; // Alias if Message is ambiguous
use serde_json;
use std::sync::Arc;
use teloxide::{
    payloads::SendMessageSetters,
    prelude::{Bot, ChatId, Requester},
};
use tokio; // tokio is not directly used in the function body but good to have if the function itself is spawned.
use tracing; // Arc is not used in the function signature or body, but might be used by the caller.

pub async fn start_kafka_consumer_loop(
    bot_consumer_clone: Bot,
    consumer: StreamConsumer, // The StreamConsumer itself is moved.
    kafka_out_topic_clone: String,
) {
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
}
