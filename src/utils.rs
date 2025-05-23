use crate::structs::{ButtonInfo, IncomingCallbackMessage};
use teloxide::types::{CallbackQuery, InlineKeyboardButton, InlineKeyboardMarkup};

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
    let chat_id = query.message.as_ref().map_or(0, |m| m.chat().id.0);
    let user_id = query.from.id.0;
    let message_id = query.message.as_ref().map_or(0, |m| m.id().0);
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
