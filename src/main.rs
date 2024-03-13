#![allow(deprecated)]
use message_types::RequestMessage;
use std::sync::mpsc;
use std::sync::mpsc::Receiver;
use std::thread;
use std::time::Duration;
use tracing::{error, info};

use teloxide::prelude::*;

use teloxide::types::{
    ChatAction, InputFile, KeyboardButton, KeyboardMarkup, ParseMode, ReplyMarkup,
};

use crate::layers::security::SecurityLayer;

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

impl BotConverter<Message> for TelegramConverter {
    fn bot_type_to_request_message(&self, message: &Message) -> RequestMessage {
        RequestMessage::new(
            message.text().unwrap_or_default().to_string(),
            message.chat.username().unwrap_or_default().to_string(),
        )
    }
}

impl TelegramConverter {
    fn new() -> Self {
        TelegramConverter {}
    }
}

pub async fn start_bot() {
    let bot = Bot::from_env();

    teloxide::repl(bot, move |bot: Bot, msg: Message| {
        info!(
            "{} sent the message: {}",
            msg.chat.username().unwrap_or_default(),
            msg.text().unwrap_or_default()
        );

        async move {
            let bc = TelegramConverter::new();
            let handler = handler::Handler::new();

            bot.send_chat_action(msg.chat.id, ChatAction::Typing)
                .await?;

            let mut request_message: RequestMessage = bc.bot_type_to_request_message(&msg);

            let res = handler.handle_message(&mut request_message).await;

            if let Some(bytes) = res.bytes {
                bot.send_document(msg.chat.id, InputFile::memory(bytes).file_name(res.text))
                    .await?;

                return Ok(());
            }

            if let Some(options) = res.options {
                let keyboard_row: Vec<KeyboardButton> = options
                    .iter()
                    .map(|title| KeyboardButton::new(title.to_string()))
                    .collect();

                let keyboard = KeyboardMarkup::default()
                    .append_row(keyboard_row)
                    .resize_keyboard(true);
                
                info!("responding with keyboard: {:?}", keyboard);
                info!("responding with text: {:?}", res.text);
                let r = bot.send_message(msg.chat.id, res.text)
                    .parse_mode(ParseMode::MarkdownV2)
                    .reply_markup(ReplyMarkup::Keyboard(keyboard))
                    .await;
                if let Err(e) = r {
                    error!("Failed to send message: {}", e);
                }
                

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

            Ok(())
        }
    })
    .await;
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
