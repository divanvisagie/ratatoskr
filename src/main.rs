#![allow(deprecated)]
use message_types::RequestMessage;
use std::error::Error;
use std::sync::mpsc;
use std::sync::mpsc::Receiver;
use std::thread;
use std::time::Duration;
use tracing::{error, info};

use teloxide::{
    payloads::SendMessageSetters,
    prelude::*,
    types::{
        InlineKeyboardButton, InlineKeyboardMarkup, InlineQueryResultArticle, InputMessageContent,
        InputMessageContentText, Me,
    },
    utils::command::BotCommands,
};

use teloxide::types::{
    ChatAction, InputFile, KeyboardButton, KeyboardMarkup, ParseMode, ReplyMarkup,
};

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
/// Parse the text wrote on Telegram and check if that text is a valid command
/// or not, then match the command. If the command is `/start` it writes a
/// markup with the `InlineKeyboardMarkup`.
async fn message_handler(bot: Bot, msg: Message) -> Result<(), Box<dyn Error + Send + Sync>> {
    info!(
        "{} sent the message: {}",
        msg.chat.username().unwrap_or_default(),
        msg.text().unwrap_or_default()
    );
    if let Some(_text) = msg.text() {
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
            let keyboard = make_keyboard(options);
            bot.send_message(msg.chat.id, res.text)
                .reply_markup(keyboard)
                .await?;
            // bot.edit_message_reply_markup(msg.chat.id, msg.id)
            //     .reply_markup(keyboard)
            //     .await?;
            // let keyboard_row: Vec<KeyboardButton> = options
            //     .iter()
            //     .map(|title| KeyboardButton::new(title.to_string()))
            //     .collect();
            //
            // let keyboard = KeyboardMarkup::default()
            //     .append_row(keyboard_row)
            //     .resize_keyboard(true);
            //
            // info!("responding with keyboard: {:?}", keyboard);
            // info!("responding with text: {:?}", res.text);
            // let r = bot
            //     .send_message(msg.chat.id, res.text)
            //     .parse_mode(ParseMode::MarkdownV2)
            //     .reply_markup(ReplyMarkup::Keyboard(keyboard))
            //     .await;
            // if let Err(e) = r {
            //     error!("Failed to send message: {}", e);
            // }

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

        // let keyboard = make_keyboard();
        // bot.send_message(msg.chat.id, "Debian versions:")
        //     .reply_markup(keyboard)
        //     .await?;
    }

    Ok(())
}

/// Creates a keyboard made by buttons in a big column.
fn make_keyboard(options_list: Vec<String>) -> InlineKeyboardMarkup {
    let mut keyboard: Vec<Vec<InlineKeyboardButton>> = vec![];

    // let debian_versions = [
    //     "Buzz", "Rex", "Bo", "Hamm", "Slink", "Potato", "Woody", "Sarge", "Etch", "Lenny",
    //     "Squeeze", "Wheezy", "Jessie", "Stretch", "Buster", "Bullseye",
    // ];

    for versions in options_list.chunks(3) {
        let row = versions
            .iter()
            .map(|version| InlineKeyboardButton::callback(version.to_owned(), version.to_owned()))
            .collect();

        keyboard.push(row);
    }

    InlineKeyboardMarkup::new(keyboard)
}
// async fn inline_query_handler(
//     bot: Bot,
//     q: InlineQuery,
// ) -> Result<(), Box<dyn Error + Send + Sync>> {
//     let choose_debian_version = InlineQueryResultArticle::new(
//         "0",
//         "Chose debian version",
//         InputMessageContent::Text(InputMessageContentText::new("Debian versions:")),
//     )
//     .reply_markup(make_keyboard());
//
//     bot.answer_inline_query(q.id, vec![choose_debian_version.into()])
//         .await?;
//
//     Ok(())
// }

/// When it receives a callback from a button it edits the message with all
/// those buttons writing a text with the selected Debian version.
///
/// **IMPORTANT**: do not send privacy-sensitive data this way!!!
/// Anyone can read data stored in the callback button.
async fn callback_handler(bot: Bot, q: CallbackQuery) -> Result<(), Box<dyn Error + Send + Sync>> {
    if let Some(version) = q.data {
        let text = format!("You chose: {version}");

        // Tell telegram that we've seen this query, to remove 🕑 icons from the
        // clients. You could also use `answer_callback_query`'s optional
        // parameters to tweak what happens on the client side.
        bot.answer_callback_query(q.id).await?;

        // Edit text of the message to which the buttons were attached
        if let Some(Message { id, chat, .. }) = q.message {
            bot.edit_message_text(chat.id, id, text).await?;
        } else if let Some(id) = q.inline_message_id {
            bot.edit_message_text_inline(id, text).await?;
        }

        info!("You chose: {}", version);
    }

    Ok(())
}
pub async fn start_bot() {
    let bot = Bot::from_env();

    // let handler = dptree::entry()
    //     .branch(Update::filter_message().endpoint(message_handler))
    //     .branch(Update::filter_callback_query().endpoint(callback_handler))
    //     .branch(Update::filter_inline_query().endpoint(inline_query_handler));

    let handler = dptree::entry()
        .branch(
            Update::filter_message()
                .endpoint(move |bot: Bot, msg| async move { message_handler(bot, msg).await }),
        )
        .branch(
            Update::filter_callback_query()
                .endpoint(move |bot: Bot, q| async move { callback_handler(bot, q).await }),
        );
        // .branch(
        //     Update::filter_inline_query()
        //         .endpoint(move |bot: Bot, q| async move { inline_query_handler(bot, q).await }),
        // );

    Dispatcher::builder(bot, handler)
        .enable_ctrlc_handler()
        .build()
        .dispatch()
        .await;

    // teloxide::repl(bot, move |bot: Bot, msg: Message| {
    //     info!(
    //         "{} sent the message: {}",
    //         msg.chat.username().unwrap_or_default(),
    //         msg.text().unwrap_or_default()
    //     );
    //
    //     async move {
    //         let bc = TelegramConverter::new();
    //         let handler = handler::Handler::new();
    //
    //         bot.send_chat_action(msg.chat.id, ChatAction::Typing)
    //             .await?;
    //
    //         let mut request_message: RequestMessage = bc.bot_type_to_request_message(&msg);
    //
    //         let res = handler.handle_message(&mut request_message).await;
    //
    //         if let Some(bytes) = res.bytes {
    //             bot.send_document(msg.chat.id, InputFile::memory(bytes).file_name(res.text))
    //                 .await?;
    //
    //             return Ok(());
    //         }
    //
    //         if let Some(options) = res.options {
    //             let keyboard_row: Vec<KeyboardButton> = options
    //                 .iter()
    //                 .map(|title| KeyboardButton::new(title.to_string()))
    //                 .collect();
    //
    //             let keyboard = KeyboardMarkup::default()
    //                 .append_row(keyboard_row)
    //                 .resize_keyboard(true);
    //
    //             info!("responding with keyboard: {:?}", keyboard);
    //             info!("responding with text: {:?}", res.text);
    //             let r = bot.send_message(msg.chat.id, res.text)
    //                 .parse_mode(ParseMode::MarkdownV2)
    //                 .reply_markup(ReplyMarkup::Keyboard(keyboard))
    //                 .await;
    //             if let Err(e) = r {
    //                 error!("Failed to send message: {}", e);
    //             }
    //
    //
    //             return Ok(());
    //         }
    //
    //         info!("responding with text: {:?}", res.text);
    //         match bot
    //             .send_message(msg.chat.id, res.text.clone())
    //             .parse_mode(ParseMode::Markdown)
    //             .reply_to_message_id(msg.id)
    //             .await
    //         {
    //             Ok(_) => (),
    //             Err(e) => {
    //                 bot.send_message(msg.chat.id, res.text.clone()).await?;
    //                 error!("Failed to send message: {}", e)
    //             }
    //         };
    //
    //         Ok(())
    //     }
    // })
    // .await;
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
