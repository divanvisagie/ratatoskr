use std::env;
use dotenv::dotenv;
use teloxide::prelude::*;
use rdkafka::config::ClientConfig;
use rdkafka::producer::{FutureProducer, FutureRecord};
use rdkafka::consumer::{StreamConsumer, Consumer};
use rdkafka::message::Message as KafkaMessage;
use serde::{Serialize, Deserialize};
use futures_util::StreamExt;
use std::sync::Arc;

#[derive(Serialize, Deserialize)]
struct OutgoingKafkaMessage {
    chat_id: i64,
    text: String,
}

#[tokio::main]
async fn main() {
    dotenv().ok();
    let telegram_token = env::var("TELEGRAM_BOT_TOKEN").expect("TELEGRAM_BOT_TOKEN not set");
    let kafka_broker = env::var("KAFKA_BROKER").unwrap_or_else(|_| "localhost:9092".to_string());
    let kafka_in = env::var("KAFKA_IN_TOPIC").unwrap_or_else(|_| "com.sectorflabs.ratatoskr.in".to_string());
    let kafka_out = env::var("KAFKA_OUT_TOPIC").unwrap_or_else(|_| "com.sectorflabs.ratatoskr.out".to_string());

    let bot = Bot::new(telegram_token.clone());

    // Kafka producer
    let producer: FutureProducer = ClientConfig::new()
        .set("bootstrap.servers", &kafka_broker)
        .create()
        .expect("Producer creation error");

    // Kafka consumer
    let consumer: StreamConsumer = ClientConfig::new()
        .set("group.id", "ratatoskr-bot-consumer")
        .set("bootstrap.servers", &kafka_broker)
        .set("enable.partition.eof", "false")
        .set("session.timeout.ms", "6000")
        .set("enable.auto.commit", "true")
        .create()
        .expect("Consumer creation failed");
    consumer.subscribe(&[&kafka_out]).expect("Can't subscribe to topic");

    // Kafka -> Telegram
    let bot2 = bot.clone();
    tokio::spawn(async move {
        let mut stream = consumer.stream();
        while let Some(Ok(msg)) = stream.next().await {
            if let Some(payload) = msg.payload() {
                if let Ok(out) = serde_json::from_slice::<OutgoingKafkaMessage>(payload) {
                    let _ = bot2.send_message(ChatId(out.chat_id), out.text).await;
                }
            }
        }
    });

    // Telegram -> Kafka
    let producer = Arc::new(producer);
    let kafka_in = Arc::new(kafka_in);
    teloxide::repl(bot, move |bot: Bot, msg: Message| {
        let producer = Arc::clone(&producer);
        let kafka_in = Arc::clone(&kafka_in);
        async move {
            let json = serde_json::to_string(&msg).unwrap();
            let record = FutureRecord::to(&kafka_in)
                .payload(&json)
                .key("message");
            let _ = producer.send(record, None).await;
            Ok(())
        }
    })
    .await;
}
