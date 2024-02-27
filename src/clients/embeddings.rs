use std::env;

use reqwest::header;
use serde::{Deserialize, Serialize};
use tracing::error;

pub struct EmbeddingsClient {}

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

impl EmbeddingsClient {
    pub fn new() -> Self {
        EmbeddingsClient {}
    }

    pub async fn get_embeddings(
        &self,
        text: String,
    ) -> Result<Vec<f32>, Box<dyn std::error::Error>> {
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
                error!("Error: {}", e);
                "Error".to_string()
            }
        };

        let response_object: EmbeddingsResponse = serde_json::from_str(&response).unwrap();

        let embeddings = response_object.data[0].embedding.clone();
        Ok(embeddings)
    }
}
