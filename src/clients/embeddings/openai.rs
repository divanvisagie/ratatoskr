use std::env;

use async_trait::async_trait;
use reqwest::header;
use tracing::error;

use super::{EmbeddingsClient, EmbeddingsRequest, EmbeddingsResponse};

pub struct OpenAiEmbeddingsClient {}
impl OpenAiEmbeddingsClient {
    #[allow(dead_code)]
    pub fn new() -> Self {
        OpenAiEmbeddingsClient {}
    }
}

#[async_trait]
impl EmbeddingsClient for OpenAiEmbeddingsClient {
    async fn get_embeddings(
        &self,
        text: String,
    ) -> Result<Vec<f32>, Box<dyn std::error::Error + Send + Sync>> {
        let api_key =
            env::var("OPENAI_API_KEY").expect("Missing OPENAI_API_KEY environment variable");

        let client = reqwest::Client::new();

        let url = "https://api.openai.com/v1/embeddings";

        let mut headers = header::HeaderMap::new();
        headers.insert(
            header::CONTENT_TYPE,
            header::HeaderValue::from_static("application/json"),
        );
        headers.insert(
            header::AUTHORIZATION,
            header::HeaderValue::from_str(&format!("Bearer {}", api_key)).unwrap(),
        );

        let request_body = serde_json::to_string(&EmbeddingsRequest {
            input: text.to_string(),
            model: "text-embedding-ada-002".to_string(),
        })
        .unwrap();
        let response = client
            .post(url)
            .headers(headers)
            .body(request_body)
            .send()
            .await;

        let response = match response {
            Ok(response) => response.text().await.unwrap(),
            Err(e) => {
                error!("Error in response: {}", e);
                return Err(Box::new(e));
            }
        };

        let response_object: EmbeddingsResponse = match serde_json::from_str(&response) {
            Ok(object) => object,
            Err(e) => {
                error!("Error in respone object: {}", e);
                return Err(Box::new(e));
            }
        };

        let embeddings = response_object.data[0].embedding.clone();
        Ok(embeddings)
    }
}
