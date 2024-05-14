use std::env;

use async_trait::async_trait;
use reqwest::header;
use serde::{Deserialize, Serialize};
use tracing::{error, info};

use super::{EmbeddingsClient, EmbeddingsRequest, EmbeddingsResponse};

/// Ollama Client
/// Implementation of the EmbeddingsClient trait which uses the Ollama service
pub struct OllamaEmbeddingsClient {
    base_url: String,
}

/*
* https://www.sbert.net/docs/pretrained_models.html
* curl http://localhost:11434/api/embeddings -d '{
*  "model": "all-minilm",
*  "prompt": "Here is an article about llamas..."
* }'
**/

#[derive(Serialize)]
struct OllamaRequest {
    model: String,
    prompt: String,
}

#[derive(Deserialize)]
struct OllamaResponse {
    embedding: Vec<f32>,
}

impl OllamaEmbeddingsClient {
    //ignore unused
    #[allow(dead_code)]
    pub fn new() -> Self {
        OllamaEmbeddingsClient {
            base_url: "http://localhost:11434".to_string(),
        }
    }
}

#[async_trait]
impl EmbeddingsClient for OllamaEmbeddingsClient {
    async fn get_embeddings(
        &self,
        text: String,
    ) -> Result<Vec<f32>, Box<dyn std::error::Error + Send + Sync>> {
        info!("Ollama embeddings for: {}", text);
        let url = format!("{}/api/embeddings", self.base_url,);
        let client = reqwest::Client::new();

        let request_body = serde_json::to_string(&OllamaRequest {
            model: "all-minilm".to_string(),
            prompt: text.to_string(),
        });

        let response = client.post(&url).body(request_body.unwrap()).send().await;

        let ollama_response = match response {
            Ok(response) => response.text().await.unwrap(),
            Err(e) => {
                error!("Error in response: {}", e);
                return Err(Box::new(e));
            }
        };
        let response_object: OllamaResponse = match serde_json::from_str(&ollama_response) {
            Ok(object) => object,
            Err(e) => {
                error!("Error in respone object: {}", e);
                return Err(Box::new(e));
            }
        };

        Ok(response_object.embedding)
    }
}
