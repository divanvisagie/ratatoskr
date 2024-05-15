#[allow(dead_code)]
use async_trait::async_trait;
use tracing::info;

use crate::{
    clients::{
        embeddings::{EmbeddingsClient, EmbeddingsClientImpl},
        image::{ImageGenerationClientImpl, ImageGenerationClient},
    },
    message_types::{RequestMessage, ResponseMessage},
};

use super::{cosine_similarity, Capability};

pub struct ImageGenerationCapability {
    description: String,
    embedding_client: EmbeddingsClientImpl,
    image_client_type: ImageGenerationClientImpl,
}

impl ImageGenerationCapability {
    pub fn new(
        image_client_type: ImageGenerationClientImpl,
        embedding_client: EmbeddingsClientImpl,
    ) -> Self {
        ImageGenerationCapability {
            description: "Could you generate me an image?".to_string(),
            embedding_client,
            image_client_type,
        }
    }
}

#[async_trait]
impl Capability for ImageGenerationCapability {
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
        // extract the client out of the type
        let x = self
            .image_client_type
            .generate_image(message.text.clone())
            .await;

        let bytes = match x {
            Ok(bytes) => bytes,
            Err(_) => return ResponseMessage {
                bytes: None,
                options: None,
                text: "Failed to generate image".to_string(),
            },
        };

        info!("Generated image with {} bytes", bytes.len());
        // create new response message stub
        let response = ResponseMessage {
            bytes: Some(bytes),
            options: None,
            text: "image.png".to_string(),
        };
        response
    }
}
