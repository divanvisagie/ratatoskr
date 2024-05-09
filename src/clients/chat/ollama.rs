use serde::{Deserialize, Serialize};
use tracing::{error, info};

use std::{env, fmt};

use reqwest::header;
use serde_json::Result;
use super::{ChatClient, Message};

/// Ollama client implementation
pub struct OllamaClient;
#[allow(dead_code)]
impl OllamaClient {
    pub fn new() -> Self {
        OllamaClient {}
    }
}

#[derive(Deserialize)]
struct OllamaResponse {
    pub message: Message,
}
#[derive(Serialize)]
struct OllamaRequest {
    pub model: String,
    pub messages: Vec<Message>,
    pub stream: bool,
}
#[allow(dead_code)]
#[async_trait::async_trait]
impl ChatClient for OllamaClient {
    async fn complete(&mut self, context: Vec<Message>) -> String {
        let client = reqwest::Client::new();
        let url = "http://127.0.0.1:11434/api/chat";

        let chat_request = OllamaRequest {
            model: "phi3".to_string(),
            messages: context.clone(),
            stream: false,
        };

        let request_body = serde_json::to_string(&chat_request).unwrap();

        let response = client.post(url).body(request_body).send().await;

        let response = match response {
            Ok(response) => response.text().await,
            Err(e) => {
                error!("Error getting text from response: {}", e);
                return "Error".to_string();
            }
        };

        let response_text = match response {
            Ok(response) => response,
            Err(e) => {
                error!("Error: {}", e);
                return "Error".to_string();
            }
        };

        info!("response_text: {}", response_text);
        let response_object: Result<OllamaResponse> = serde_json::from_str(&response_text);
        let response_object = match response_object {
            Ok(response) => response,
            Err(e) => {
                error!("Error: {}", e);
                return "Error".to_string();
            }
        };

        response_object.message.content
    }
}
