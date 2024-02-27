use async_trait::async_trait;
use serde::{Deserialize, Serialize};

#[derive(Debug, Serialize, Deserialize, Clone)]
struct BERTEmbedding {
    pub values: Vec<f32>,
}

use crate::{
    capabilities::cosine_similarity, clients::embeddings::EmbeddingsClient,
    message_types::ResponseMessage, RequestMessage,
};

use super::Capability;

pub struct PrivacyCapability {
    description: String,
}

impl PrivacyCapability {
    pub fn new() -> Self {
        let description = "What is the privacy policy of this service?".to_string();
        PrivacyCapability { description }
    }
}

#[async_trait]
impl Capability for PrivacyCapability {
    async fn check(&mut self, message: &RequestMessage) -> f32 {
        let cl = EmbeddingsClient::new();

        let description_embedding = cl.get_embeddings(self.description.clone()).await.unwrap();

        cosine_similarity(
            message.embedding.as_slice(),
            description_embedding.as_slice(),
        )
    }

    async fn execute(&mut self, _message: &RequestMessage) -> ResponseMessage {
        let res = include_str!("canned_response.md");
        ResponseMessage::new(res.to_string())
    }
}
