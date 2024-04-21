use crate::{
    clients::embeddings::{OllamaEmbeddingsClient}, layers::{
        embedding::EmbeddingLayer, memory::MemoryLayer, security::SecurityLayer,
        selector::SelectorLayer, Layer,
    }, message_types::ResponseMessage, repositories::users::FsUserRepository, RequestMessage
};

pub struct Handler {
    gateway_layer: Box<dyn Layer>,
}

impl Handler {
    pub fn new() -> Self {
        if cfg!(debug_assertions) {
            let embeddings_client = BarnstokkrClient::new();
            let selector_layer = SelectorLayer::new();
            let embedding_layer = EmbeddingLayer::new(selector_layer,embeddings_client);
            let memory_layer = MemoryLayer::new(embedding_layer);
            let user_repository = FsUserRepository::new();
            let security_layer = SecurityLayer::new(memory_layer, user_repository);
            Self {
                gateway_layer: Box::new(security_layer),
            }
        } else {
            let embedding_client = OllamaEmbeddingsClient::new();
            let selector_layer = SelectorLayer::new();
            let embedding_layer = EmbeddingLayer::new(selector_layer, embedding_client);
            let memory_layer = MemoryLayer::new(embedding_layer);
            let user_repository = FsUserRepository::new();
            let security_layer = SecurityLayer::new(memory_layer, user_repository);
            Self {
                gateway_layer: Box::new(security_layer),
            }
        }
    }

    pub async fn handle_message(&mut self, message: &mut RequestMessage) -> ResponseMessage {
        self.gateway_layer.execute(message).await
    }
}
