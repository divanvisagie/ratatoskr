use async_trait::async_trait;

use crate::{
    clients::embeddings::EmbeddingsClient, message_types::ResponseMessage, RequestMessage,
};

use super::Layer;

pub struct EmbeddingLayer {
    embedding: EmbeddingsClient,
    next: Box<dyn Layer>,
}

impl EmbeddingLayer {
    pub fn new(next: Box<dyn Layer>) -> Self {
        EmbeddingLayer {
            embedding: EmbeddingsClient::new(),
            next,
        }
    }
}

#[async_trait]
impl Layer for EmbeddingLayer {
    async fn execute(&mut self, message: &mut RequestMessage) -> ResponseMessage {
        let embedding = self
            .embedding
            .get_embeddings(message.text.clone())
            .await
            .unwrap();

        message.embedding = embedding;

        self.next.execute(message).await
    }
}
