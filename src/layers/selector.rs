use super::Layer;
use crate::capabilities::chat::ChatCapability;
use crate::capabilities::debug::DebugCapability;
use crate::capabilities::privacy::PrivacyCapability;
use crate::capabilities::summarize::SummaryCapability;
use crate::capabilities::test::TestCapability;
use crate::clients;
use crate::clients::chat::{GptClient, OllamaClient};
use crate::message_types::ResponseMessage;
use crate::{capabilities::Capability, RequestMessage};
use async_trait::async_trait;
use tracing::info;

pub struct SelectorLayer {
    capabilities: Vec<Box<dyn Capability>>,
}

#[async_trait]
impl Layer for SelectorLayer {
    async fn execute(&mut self, message: &mut RequestMessage) -> ResponseMessage {
        let mut best: Option<&mut Box<dyn Capability>> = None;
        let mut best_score = 0.0;

        for capability in &mut self.capabilities {
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
}

impl SelectorLayer {
    pub fn new() -> Self {
        if cfg!(debug_assertions) {
            info!("Running in debug mode");
            let chat_client = OllamaClient::new();
            let embeddings_client = clients::embeddings::OllamaEmbeddingsClient::new();
            SelectorLayer {
                capabilities: vec![
                    Box::new(DebugCapability::new()),
                    Box::new(PrivacyCapability::new()),
                    Box::new(ChatCapability::new(chat_client, embeddings_client)),
                    Box::new(SummaryCapability::new(OllamaClient::new())),
                    Box::new(TestCapability::new()),
                ],
            }
        } else {
            info!("Running in production mode");
            let chat_client = clients::chat::GptClient::new();
            let embeddings_client = clients::embeddings::BarnstokkrClient::new();
            SelectorLayer {
                capabilities: vec![
                    Box::new(DebugCapability::new()),
                    Box::new(PrivacyCapability::new()),
                    Box::new(ChatCapability::new(chat_client, embeddings_client)),
                    Box::new(SummaryCapability::new(GptClient::new())),
                    Box::new(TestCapability::new()),
                ],
            }
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
            capabilities: vec![Box::new(MockCapability {})],
        };

        let mut message = RequestMessage {
            text: "Hello".to_string(),
            username: "test".to_string(),
            context: Vec::new(),
            embedding: Vec::new(),
        };
        let response = layer.execute(&mut message).await;
        assert_eq!(response.text, "Hello, test!");
    }
}
