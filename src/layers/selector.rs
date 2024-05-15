use std::env;

use super::Layer;
use crate::capabilities::chat::ChatCapability;
use crate::capabilities::debug::DebugCapability;
use crate::capabilities::group_chat::GroupChatCapability;
use crate::capabilities::image_generation::ImageGenerationCapability;
use crate::capabilities::privacy::PrivacyCapability;
use crate::capabilities::summarize::SummaryCapability;
use crate::capabilities::test::TestCapability;
use crate::clients::chat::{ChatClientImpl, GptClient, OllamaClient};
use crate::clients::embeddings::{EmbeddingsClientImpl, OllamaEmbeddingsClient};
use crate::clients::image::{DalleClient, ImageGenerationClientImpl};
use crate::message_types;
use crate::message_types::ResponseMessage;
use crate::{capabilities::Capability, RequestMessage};

use async_trait::async_trait;
use tracing::info;

pub struct SelectorLayer {
    private_capabilities: Vec<Box<dyn Capability>>,
    group_capabilities: Vec<Box<dyn Capability>>,
}

#[async_trait]
impl Layer for SelectorLayer {
    async fn execute(&mut self, message: &mut RequestMessage) -> ResponseMessage {
        match &message.chat_type {
            message_types::ChatType::Private => self.execute_private(message).await,
            message_types::ChatType::Group(_) => self.execute_group(message).await,
        }
    }
}

impl SelectorLayer {
    pub fn new() -> Self {
        if cfg!(debug_assertions) {
            info!("Running in debug mode");
            SelectorLayer {
                private_capabilities: vec![
                    Box::new(DebugCapability::new()),
                    Box::new(PrivacyCapability::new()),
                    Box::new(ChatCapability::new(
                        ChatClientImpl::OpenAi(GptClient::new()),
                        EmbeddingsClientImpl::Ollama(OllamaEmbeddingsClient::new()),
                    )),
                    Box::new(SummaryCapability::new(ChatClientImpl::OpenAi(
                        GptClient::new(),
                    ))),
                    Box::new(ImageGenerationCapability::new(
                        ImageGenerationClientImpl::Dalle(DalleClient::new()),
                        EmbeddingsClientImpl::Ollama(OllamaEmbeddingsClient::new()),
                    )),
                    Box::new(TestCapability::new()),
                ],
                group_capabilities: vec![
                    Box::new(SummaryCapability::new(ChatClientImpl::Ollama(
                        OllamaClient::new(),
                    ))),
                    Box::new(ImageGenerationCapability::new(
                        ImageGenerationClientImpl::Dalle(DalleClient::new()),
                        EmbeddingsClientImpl::Ollama(OllamaEmbeddingsClient::new()),
                    )),
                    Box::new(GroupChatCapability::new(
                        OllamaClient::new(),
                        OllamaEmbeddingsClient::new(),
                    )),
                    Box::new(TestCapability::new()),
                ],
            }
        } else {
            info!("Running in production mode");
            SelectorLayer {
                private_capabilities: vec![
                    Box::new(DebugCapability::new()),
                    Box::new(PrivacyCapability::new()),
                    Box::new(ChatCapability::new(
                        ChatClientImpl::OpenAi(GptClient::new()),
                        EmbeddingsClientImpl::Ollama(OllamaEmbeddingsClient::new()),
                    )),
                    Box::new(SummaryCapability::new(ChatClientImpl::OpenAi(
                        GptClient::new(),
                    ))),
                    Box::new(ImageGenerationCapability::new(
                        ImageGenerationClientImpl::Dalle(DalleClient::new()),
                        EmbeddingsClientImpl::Ollama(OllamaEmbeddingsClient::new()),
                    )),
                    Box::new(TestCapability::new()),
                ],
                group_capabilities: vec![
                    Box::new(SummaryCapability::new(ChatClientImpl::OpenAi(
                        GptClient::new(),
                    ))),
                    Box::new(GroupChatCapability::new(
                        GptClient::new(),
                        OllamaEmbeddingsClient::new(),
                    )),
                    Box::new(ImageGenerationCapability::new(
                        ImageGenerationClientImpl::Dalle(DalleClient::new()),
                        EmbeddingsClientImpl::Ollama(OllamaEmbeddingsClient::new()),
                    )),
                ],
            }
        }
    }

    async fn execute_private(&mut self, message: &mut RequestMessage) -> ResponseMessage {
        let mut best: Option<&mut Box<dyn Capability>> = None;
        let mut best_score = -1.0;

        for capability in &mut self.private_capabilities {
            let score = capability.check(message).await;
            info!("{} similarity: {}", capability.get_name(), score);
            if score > best_score {
                best_score = score;
                best = Some(capability);
            }
        }
        match best {
            Some(capability) => {
                info!("Selected capability: {}", capability.get_name());
                capability.execute(message).await
            }
            None => ResponseMessage::new("No capability found".to_string()),
        }
    }

    async fn execute_group(&mut self, message: &mut RequestMessage) -> ResponseMessage {
        let bot_name = env::var("BOT_NAME").unwrap_or_else(|_| "@".to_string());
        if !message.text.contains(&bot_name) {
            return ResponseMessage::new("".to_string())
        }

        let mut best: Option<&mut Box<dyn Capability>> = None;
        let mut best_score = -1.0;

        for capability in &mut self.group_capabilities {
            let score = capability.check(message).await;
            info!("{} similarity: {}", capability.get_name(), score);
            if score > best_score {
                best_score = score;
                best = Some(capability);
            }
        }
        match best {
            Some(capability) => {
                info!("Selected capability: {}", capability.get_name());
                capability.execute(message).await
            }
            None => ResponseMessage::new("".to_string()),
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use async_trait::async_trait;

    struct MockCapability {}
    #[async_trait]
    impl Capability for MockCapability {
        async fn check(&mut self, message: &RequestMessage) -> f32 {
            if message.text == "Hello" {
                1.0
            } else {
                0.0
            }
        }
        async fn execute(&mut self, message: &RequestMessage) -> ResponseMessage {
            ResponseMessage::new(format!("Hello, {}!", message.username))
        }
    }

    #[tokio::test]
    async fn test_selector_layer() {
        let mut layer = SelectorLayer {
            private_capabilities: vec![Box::new(MockCapability {})],
            group_capabilities: Vec::new(),
        };

        let mut message = RequestMessage {
            text: "Hello".to_string(),
            username: "test".to_string(),
            context: Vec::new(),
            embedding: Vec::new(),
            chat_type: message_types::ChatType::Private,
            chat_id: 0,
            sent_by: "".to_string(),
        };
        let response = layer.execute(&mut message).await;
        assert_eq!(response.text, "Hello, test!");
    }
}
