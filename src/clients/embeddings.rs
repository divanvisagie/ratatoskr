use std::env;

use async_trait::async_trait;
use reqwest::header;
use serde::{Deserialize, Serialize};
use tracing::{error, info};
pub struct OpenAiEmbeddingsClient {}

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

/// Barnstokker Client
/// Implementation of the EmbeddingsClient trait which uses the Barnstokkr service
pub struct BarnstokkrClient<'a> {
    base_url: &'a str,
}

#[derive(Serialize)]
struct BarnstokkrRequest {
    text: String,
}

#[derive(Deserialize)]
struct BarnstokkrResponse {
    embeddings: Vec<f32>,
}


impl <'a> BarnstokkrClient<'a> {
    //ignore unused 
    #[allow(dead_code)]
    pub fn new() -> Self {
        BarnstokkrClient {
            base_url: "http://127.0.0.1:8000",
        }
    }
}

#[async_trait]
impl <'a> EmbeddingsClient for BarnstokkrClient <'a> {
    async fn get_embeddings(
        &self,
        text: String,
    ) -> Result<Vec<f32>, Box<dyn std::error::Error + Send + Sync>> {
        info!("Barnstokkr embeddings for: {}", text);
        let url = format!(
            "{}/embeddings",
            self.base_url,
        );
        let client = reqwest::Client::new();

        let request_body = serde_json::to_string(&BarnstokkrRequest{
            text: text.to_string(),
        });

        let response = client
            .post(&url)
            .body(request_body.unwrap())
            .send()
            .await;

        let barnstokkr_response = match response {
            Ok(response) => response.text().await.unwrap(),
            Err(e) => {
                error!("Error in response: {}", e);
                return Err(Box::new(e));
            }
        };
        let response_object: BarnstokkrResponse = match serde_json::from_str(&barnstokkr_response) {
            Ok(object) => object,
            Err(e) => {
                error!("Error in respone object: {}", e);
                return Err(Box::new(e));
            }
        };

        Ok(response_object.embeddings)
    }
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
