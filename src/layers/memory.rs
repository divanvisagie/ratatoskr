use std::borrow::Borrow;

use serde::{Deserialize, Serialize};
use tracing::error;

use async_trait::async_trait;

use crate::{
    clients::{
        chatgpt::Role,
        muninn::{MunninClient, MunninClientImpl},
    },
    message_types::ResponseMessage,
    RequestMessage,
};

use super::Layer;

#[derive(Clone, Debug, Serialize, Deserialize)]
pub struct StoredMessage {
    pub username: String,
    pub text: String,
    pub role: Role,
}

pub struct MemoryLayer {
    next: Box<dyn Layer>,
}

#[async_trait]
impl Layer for MemoryLayer {
    async fn execute(&mut self, message: &mut RequestMessage) -> ResponseMessage {
        let munnin_client = MunninClientImpl::new();

        let context = match munnin_client.get_context(message.username.clone()).await {
            Ok(context) => context,
            Err(err) => {
                error!("Failed to get context: {:?}", err);
                return ResponseMessage {
                    bytes: None,
                    options: None,
                    text: "Failed to get context".to_string(),
                };
            }
        };

        // Convert context to stored messages
        let mut stored_context: Vec<StoredMessage> = Vec::new();
        for chat_response in context {
            stored_context.push(StoredMessage {
                username: message.username.clone(),
                text: chat_response.content,
                role: match chat_response.role.as_str() {
                    "user" => Role::User,
                    "assistant" => Role::Assistant,
                    _ => Role::System,
                },
            });
        }

        message.context = stored_context;
        let res = self.next.execute(message).await;

        munnin_client
            .save(
                message.username.to_string().borrow(),
                "user".to_string(),
                message.text.clone(),
            )
            .await
            .unwrap();

        munnin_client
            .save(
                message.username.to_string().borrow(),
                "assistant".to_string(),
                res.text.clone(),
            )
            .await
            .unwrap();

        res
    }
}

impl MemoryLayer {
    pub fn new(next: Box<dyn Layer>) -> Self {
        MemoryLayer { next }
    }
}
