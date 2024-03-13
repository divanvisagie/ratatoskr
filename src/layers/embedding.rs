use async_trait::async_trait;

use crate::{
    clients::embeddings::{EmbeddingsClient, OllamaEmbeddingsClient},
    message_types::ResponseMessage,
    RequestMessage,
};

use super::Layer;

pub struct EmbeddingLayer<E: EmbeddingsClient, L: Layer> {
    embedding_client: E,
    next: L,
}

impl <E: EmbeddingsClient, L: Layer>EmbeddingLayer <E, L> {
    pub fn new(next: L) -> Self {
        EmbeddingLayer {
            embedding_client: OllamaEmbeddingsClient::new(),
            next,
        }
    }
}

#[async_trait]
impl <E: EmbeddingsClient, L: Layer>Layer for EmbeddingLayer<E, L> {
    async fn execute(&mut self, message: &mut RequestMessage) -> ResponseMessage {
        let embedding = self
            .embedding_client
            .get_embeddings(message.text.clone())
            .await
            .unwrap();

        message.embedding = embedding;

        self.next.execute(message).await
    }
}
