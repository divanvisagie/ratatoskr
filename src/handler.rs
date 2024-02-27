use crate::{
    layers::{
        embedding::EmbeddingLayer, memory::MemoryLayer, security::SecurityLayer,
        selector::SelectorLayer, Layer,
    },
    message_types::ResponseMessage,
    repositories::users::FsUserRepository,
    RequestMessage,
};

pub struct Handler {
    gateway_layer: Box<dyn Layer>,
}

impl Handler {
    pub fn new() -> Self {
        let selector_layer = SelectorLayer::new();
        let embedding_layer = EmbeddingLayer::new(Box::new(selector_layer));
        let memory_layer = MemoryLayer::new(Box::new(embedding_layer));

        let user_repository = FsUserRepository::new();
        let security_layer = SecurityLayer::new(Box::new(memory_layer), Box::new(user_repository));
        Self {
            gateway_layer: Box::new(security_layer),
        }
    }

    pub async fn handle_message(mut self, message: &mut RequestMessage) -> ResponseMessage {
        self.gateway_layer.execute(message).await
    }
}
