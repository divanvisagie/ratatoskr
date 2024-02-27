use crate::{
    capabilities::{cosine_similarity, Capability},
    clients::{
        chatgpt::{GptClient, Role},
        embeddings::EmbeddingsClient,
    },
    message_types::ResponseMessage,
    RequestMessage,
};
use async_trait::async_trait;

#[derive(Debug)]
pub struct ChatCapability <'a> {
    // fields omitted
    client: GptClient,
    description: String,
    prompt: &'a str
}

#[async_trait]
impl <'a>Capability for ChatCapability<'a> {
    async fn check(&mut self, message: &RequestMessage) -> f32 {
        let cl = EmbeddingsClient::new();

        let description_embedding = cl.get_embeddings(self.description.clone()).await.unwrap();

        cosine_similarity(
            message.embedding.as_slice(),
            description_embedding.as_slice(),
        )
    }

    async fn execute(&mut self, message: &RequestMessage) -> ResponseMessage {
        self.client.add_message(Role::System, self.prompt.to_string());
        message.context.iter().for_each(|m| {
            self.client.add_message(m.role.clone(), m.text.clone());
        });
        
        self.client.add_message(Role::User, message.text.clone());
        let response = self.client.complete().await;

        ResponseMessage::new(response)
    }
}

impl <'a>ChatCapability<'a> {
    pub fn new() -> Self {
        //include bytes from prompt.txt
        let prompt = include_str!("prompt.txt");
        ChatCapability {
            client: GptClient::new(),
            description: "Any question a user may have".to_string(),
            prompt
        }
    }
}
