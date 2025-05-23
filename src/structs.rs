use serde::{Deserialize, Serialize};

#[derive(Serialize, Deserialize, Debug)]
pub struct ButtonInfo {
    text: String,
    callback_data: String,
}

#[derive(Serialize, Deserialize, Debug)]
pub struct OutgoingKafkaMessage {
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
