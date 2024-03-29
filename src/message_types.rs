#![allow(dead_code)]
use crate::layers::memory::StoredMessage;

#[derive(Clone, PartialEq)]
pub enum ChatType {
    Private,
    Group(String),
}

pub struct RequestMessage {
    pub text: String,
    pub username: String,
    pub context: Vec<StoredMessage>,
    pub embedding: Vec<f32>,
    pub chat_type: ChatType,
    pub chat_id: i64
}

impl RequestMessage {
    pub fn new(text: String, username: String, chat_type: ChatType, chatId: i64) -> Self {
        RequestMessage {
            text,
            username,
            context: Vec::new(),
            embedding: Vec::new(),
            chat_type,
            chat_id: chatId
        }
    }
}

#[derive(Clone, Debug)]
pub struct ResponseMessage {
    pub text: String,

    /// If the response is a file, it will be sent to the user as a
    /// document and the text will be used as the filename
    pub bytes: Option<Vec<u8>>,

    /// Inline button options
    pub options: Option<Vec<String>>,
}

impl ResponseMessage {
    pub fn new(text: String) -> Self {
        ResponseMessage {
            text,
            bytes: None,
            options: None,
        }
    }

    pub fn new_with_bytes(text: String, bytes: Vec<u8>) -> Self {
        ResponseMessage {
            text,
            bytes: Some(bytes),
            options: None,
        }
    }

    pub fn new_with_options(text: String, options: Vec<String>) -> Self {
        ResponseMessage {
            text,
            bytes: None,
            options: Some(options),
        }
    }
}
