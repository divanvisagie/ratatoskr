use std::env;

use async_trait::async_trait;
use serde::{Deserialize, Serialize};
use tracing::{error, info};

pub use openai::OpenAiEmbeddingsClient;
pub use ollama::OllamaEmbeddingsClient;

pub mod openai;
pub mod ollama;

pub enum EmbeddingsClientImpl {
    OpenAi(OpenAiEmbeddingsClient),
    Ollama(OllamaEmbeddingsClient),
    Mock(MockEmbeddingsClient),
}

#[async_trait]
impl EmbeddingsClient for EmbeddingsClientImpl {
    async fn get_embeddings(
        &self,
        text: String,
    ) -> Result<Vec<f32>, Box<dyn std::error::Error + Send + Sync>> {
        match self {
            EmbeddingsClientImpl::OpenAi(client) => client.get_embeddings(text).await,
            EmbeddingsClientImpl::Ollama(client) => client.get_embeddings(text).await,
            EmbeddingsClientImpl::Mock(client) => client.get_embeddings(text).await,
        }
    }
}

#[derive(Debug, Serialize, Deserialize)]
struct EmbeddingsRequest {
    input: String,
    model: String,
}

#[derive(Debug, Serialize, Deserialize)]
struct EmbeddingsResponse {
    object: String,
    data: Vec<Data>,
}

#[derive(Debug, Serialize, Deserialize)]
struct Data {
    object: String,
    index: usize,
    embedding: Vec<f32>,
}

#[async_trait]
pub trait EmbeddingsClient: Send + Sync {
    async fn get_embeddings(
        &self,
        text: String,
    ) -> Result<Vec<f32>, Box<dyn std::error::Error + Send + Sync>>;
}

/**
 * Mocking the embeddings client
 */
pub struct MockEmbeddingsClient {}
impl MockEmbeddingsClient {
    #[allow(dead_code)]
    pub fn new() -> Self {
        MockEmbeddingsClient {}
    }
}

#[async_trait]
impl EmbeddingsClient for MockEmbeddingsClient {
    async fn get_embeddings(
        &self,
        text: String,
    ) -> Result<Vec<f32>, Box<dyn std::error::Error + Send + Sync>> {
        info!("Mocking embeddings for: {}", text);
        Ok(vec![0.0, 0.0, 0.0])
    }
}
