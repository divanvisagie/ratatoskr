#![allow(deprecated)]
use message_types::RequestMessage;
use rumqttc::MqttOptions;
use serde::{Deserialize, Serialize};
use std::sync::mpsc;
use std::sync::mpsc::Receiver;
use std::thread;
use std::time::Duration;
use std::{error::Error, sync::Arc};
use tokio::sync::Mutex;
use tracing::{error, info};

use teloxide::{
    payloads::SendMessageSetters,
    prelude::*,
    types::{InlineKeyboardButton, InlineKeyboardMarkup},
};

use teloxide::types::{ChatAction, InputFile, ParseMode};

use crate::clients::muninn::{MunninClient, MunninClientImpl};

mod capabilities;
mod clients;
mod handler;
mod layers;
mod message_types;
mod repositories;

struct TelegramConverter;

trait BotConverter<T> {
    fn bot_type_to_request_message(&self, bot_message: &T) -> RequestMessage;
}

#[derive(Deserialize, Serialize, Debug)]
pub struct MessageEvent {
    pub username: String,
    pub hash: String,
    pub chat_id: i64,
}

impl BotConverter<Message> for TelegramConverter {
    fn bot_type_to_request_message(&self, message: &Message) -> RequestMessage {
        let chat_type = match message.chat.kind {
            teloxide::types::ChatKind::Private(_) => message_types::ChatType::Private,
            teloxide::types::ChatKind::Public(_) => {
                let title = message.chat.title().unwrap_or_default();
                message_types::ChatType::Group(title.to_string())
            }
        };

        let chat_id: i64 = message.chat.id.0;
        let text = message.text().unwrap_or_default().to_string();
        let username = message.chat.username().unwrap_or_default().to_string();

        let sender_username = message
            .from()
            .map(|u| u.username.clone().unwrap_or_default())
            .unwrap_or_default();
        let sender_name = message
            .from()
            .map(|u| {
                format!(
                    "{} {}",
                    u.first_name.clone(),
                    u.last_name.clone().unwrap_or_default()
                )
            })
            .unwrap_or_default();
        info!(
            ">>> Chat id: {}\nSender username: {}\nSender: {}",
            chat_id, sender_username, sender_name
        );
        let sent_by = format!("{} ({})", sender_username, sender_name);

        let text = match &chat_type {
            message_types::ChatType::Private => text,
            message_types::ChatType::Group(_) => format!("{}: {}", sent_by, text)
        };

        RequestMessage {
            text,
            username,
            context: Vec::new(),
            embedding: Vec::new(),
            chat_type,
            chat_id,
            sent_by,
        }
    }
}

impl TelegramConverter {
    fn new() -> Self {
        TelegramConverter {}
    }
}

async fn typing_loop(bot: &Bot, chat_id: ChatId) {
    loop {
        bot.send_chat_action(chat_id, ChatAction::Typing)
            .await
            .unwrap();
        tokio::time::sleep(Duration::from_secs(3)).await;
    }
}

async fn message_handler(
    bot: Bot,
    msg: Message,
    handler: Arc<Mutex<handler::Handler>>,
) -> Result<(), Box<dyn Error + Send + Sync>> {
    if let Some(_text) = msg.text() {
        let bc = TelegramConverter::new();
        let mut request_message: RequestMessage = bc.bot_type_to_request_message(&msg);
        let mut handler = handler.lock().await;

        let res = handler.handle_message(&mut request_message).await;

        // tokio::select! {
        //     _ = typing_loop(&bot, msg.chat.id) => (),
        //     _ = handler.handle_message(&mut request_message),
        // }

        if res.text.is_empty() {
            return Ok(());
        }

        bot.send_chat_action(msg.chat.id, ChatAction::Typing)
            .await?;

        if let Some(bytes) = res.bytes {
            bot.send_document(msg.chat.id, InputFile::memory(bytes).file_name(res.text))
                .await?;

            return Ok(());
        }

        if let Some(options) = res.options {
            let keyboard = make_keyboard(options);
            bot.send_message(msg.chat.id, res.text)
                .reply_markup(keyboard)
                .await?;

            return Ok(());
        }

        info!("responding with text: {:?}", res.text);
        match bot
            .send_message(msg.chat.id, res.text.clone())
            .parse_mode(ParseMode::Markdown)
            .reply_to_message_id(msg.id)
            .await
        {
            Ok(_) => (),
            Err(e) => {
                bot.send_message(msg.chat.id, res.text.clone()).await?;
                error!("Failed to send message: {}", e)
            }
        };
    }

    Ok(())
}

/// Creates a keyboard made by buttons in a big column.
fn make_keyboard(options_list: Vec<String>) -> InlineKeyboardMarkup {
    let mut keyboard: Vec<Vec<InlineKeyboardButton>> = vec![];

    for versions in options_list.chunks(3) {
        let row = versions
            .iter()
            .map(|version| InlineKeyboardButton::callback(version.to_owned(), version.to_owned()))
            .collect();

        keyboard.push(row);
    }

    InlineKeyboardMarkup::new(keyboard)
}

/// **IMPORTANT**: do not send privacy-sensitive data this way!!!
/// Anyone can read data stored in the callback button.
async fn callback_handler(
    bot: Bot,
    q: CallbackQuery,
    handler: Arc<Mutex<handler::Handler>>,
) -> Result<(), Box<dyn Error + Send + Sync>> {
    if let Some(option) = q.data {
        let text = format!("User chose: {option}");
        info!("{}", text);

        let username = q
            .message
            .clone()
            .unwrap()
            .chat
            .username()
            .unwrap_or_default()
            .to_string();

        let chat_type = match q.message.clone().unwrap().chat.kind {
            teloxide::types::ChatKind::Private(_) => message_types::ChatType::Private,
            teloxide::types::ChatKind::Public(p) => {
                let title = p.title.unwrap_or_default();
                message_types::ChatType::Group(title)
            }
        };
        // Tell telegram that we've seen this query, to remove 🕑 icons from the
        // clients. You could also use `answer_callback_query`'s optional
        // parameters to tweak what happens on the client side.
        match bot.answer_callback_query(q.id).await {
            Ok(_) => (),
            Err(e) => error!("Failed to answer callback query: {}", e),
        }
        let wait_response_text = format!("You selected {}. Please wait...", option);

        if let Some(Message { id, chat, .. }) = q.message {
            info!("Editing message: {}", id);
            match bot.edit_message_text(chat.id, id, wait_response_text).await {
                Ok(_) => (),
                Err(e) => error!("Failed to edit message: {}", e),
            }
            // convert Telegram ChatId to i64
            let chat_id: i64 = chat.id.0;
            info!("Chat id: {}", chat_id);
            let mut request_message =
                RequestMessage::new(option.clone(), username, chat_type, chat_id);
            let mut handler = handler.lock().await;
            let response = handler.handle_message(&mut request_message).await;

            match bot.send_message(chat.id, &response.text).await {
                Ok(_) => (),
                Err(e) => error!("Failed to send message with response: {}", e),
            }
        }
    }

    Ok(())
}
pub async fn start_bot() {
    let bot = Bot::from_env();

    let handler = handler::Handler::new();
    let ham = Arc::new(Mutex::new(handler));
    let handler = dptree::entry()
        .branch(Update::filter_message().endpoint({
            let ham = Arc::clone(&ham);
            move |bot: Bot, msg| {
                let ham = Arc::clone(&ham); // Clone again for each async invocation
                async move { message_handler(bot, msg, ham).await }
            }
        }))
        .branch(Update::filter_callback_query().endpoint({
            let ham = Arc::clone(&ham);
            move |bot: Bot, q| {
                let ham = Arc::clone(&ham);
                async move { callback_handler(bot, q, ham).await }
            }
        }));

    let bota = bot.clone();
    let bot_task = async {
        Dispatcher::builder(bota, handler)
            .enable_ctrlc_handler()
            .build()
            .dispatch()
            .await;
    };

    let mqtt_listener = async {
        let mut mqttoptions = MqttOptions::new("ratatoskr", "127.0.0.1", 1883);
        mqttoptions.set_keep_alive(Duration::from_secs(5));

        let (client, mut eventloop) = rumqttc::AsyncClient::new(mqttoptions, 10);
        client
            .subscribe("messages/assistant", rumqttc::QoS::AtLeastOnce)
            .await
            .unwrap();
        while let Ok(notification) = eventloop.poll().await {
            match notification {
                rumqttc::Event::Incoming(incoming) => match incoming {
                    rumqttc::Packet::Publish(publish) => {
                        let payload = publish.payload.to_vec();
                        let deserialized: MessageEvent =
                            rmp_serde::from_read_ref(&payload).unwrap();
                        info!("Received publish message from MQTT: {:?}", deserialized);

                        let mc = MunninClientImpl::new();
                        let msg = mc
                            .get_chat_message(&deserialized.username, &deserialized.hash)
                            .await;
                        match msg {
                            Ok(msg) => {
                                bot.send_message(ChatId(deserialized.chat_id), msg.content)
                                    .parse_mode(ParseMode::Markdown)
                                    .await
                                    .unwrap();
                            }
                            Err(e) => {
                                error!("Failed to get message from muninn: {:?}", e);
                            }
                        }
                    }
                    _ => {}
                },
                _ => {}
            }
        }
    };

    tokio::join!(bot_task, mqtt_listener);
}

pub async fn start_receiver(receiver: Receiver<String>) {
    loop {
        match receiver.try_recv() {
            Ok(message) => {
                info!("Received message from channel: {}", message);
            }
            Err(mpsc::TryRecvError::Empty) => {
                thread::sleep(Duration::from_secs(1));
            }
            Err(mpsc::TryRecvError::Disconnected) => {
                info!("Sender has been disconnected.");
                break;
            }
        }
    }
}

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error + Send + Sync>> {
    tracing_subscriber::fmt::init();
    info!("Starting bot...");

    start_bot().await;

    Ok(())
}
