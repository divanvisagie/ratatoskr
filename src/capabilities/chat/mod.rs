use crate::clients::embeddings::EmbeddingsClient;
use crate::{
    capabilities::Capability,
    clients::chat::{ChatClient, ChatClientImpl, ContextBuilder, Role},
    message_types::ResponseMessage,
    RequestMessage,
};
use async_trait::async_trait;

use super::cosine_similarity;

pub struct ChatCapability<'a, E: EmbeddingsClient> {
    client: ChatClientImpl,
    embedding_client: E,
    description: String,
    prompt: &'a str,
}

#[async_trait]
impl<'a, E: EmbeddingsClient> Capability for ChatCapability<'a, E> {
    async fn check(&mut self, message: &RequestMessage) -> f32 {
        let description_embedding = self
            .embedding_client
            .get_embeddings(self.description.clone())
            .await
            .unwrap();

        cosine_similarity(
            message.embedding.as_slice(),
            description_embedding.as_slice(),
        )
    }

    async fn execute(&mut self, message: &RequestMessage) -> ResponseMessage {
        let mut builder = ContextBuilder::new();

        builder.add_message(Role::System, self.prompt.to_string());
        let context = message.context.iter().collect::<Vec<_>>();
        context.iter().for_each(|m| {
            builder.add_message(m.role.clone(), m.text.clone());
        });

        builder.add_message(Role::User, message.text.clone());
        let response = self.client.complete(builder.build()).await;

        ResponseMessage::new(response)
    }
}

impl<'a, E: EmbeddingsClient> ChatCapability<'a, E> {
    pub fn new(client: ChatClientImpl, embeddings_client: E) -> Self {
        //include bytes from prompt.txt
        let prompt = include_str!("prompt.txt");
        ChatCapability {
            client,
            embedding_client: embeddings_client,
            description: "Any question a user may have".to_string(),
            prompt,
        }
    }
}
