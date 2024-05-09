use serde::{Deserialize, Serialize};
use tracing::{error, info};

use std::{env, fmt};

use reqwest::header;
use serde_json::Result;
use super::{parse_response, ChatClient, ChatRequest, Message};

pub struct GptClient;
#[allow(dead_code)]
impl GptClient {
    pub fn new() -> Self {
        GptClient {}
    }
}

#[allow(dead_code)]
#[async_trait::async_trait]
impl ChatClient for GptClient {
    //complete method
    async fn complete(&mut self, context: Vec<Message>) -> String {
        // Retrieve the API key from the environment variable
        let api_key =
            env::var("OPENAI_API_KEY").expect("Missing OPENAI_API_KEY environment variable");

        let client = reqwest::Client::new();
        let url = "https://api.openai.com/v1/chat/completions";

        let mut headers = header::HeaderMap::new();
        headers.insert(
            header::CONTENT_TYPE,
            header::HeaderValue::from_static("application/json"),
        );
        headers.insert(
            header::AUTHORIZATION,
            header::HeaderValue::from_str(&format!("Bearer {}", api_key)).unwrap(),
        );

        let chat_request = ChatRequest {
            model: "gpt-4-turbo-preview".to_string(),
            messages: context.clone(),
        };

        let request_body = serde_json::to_string(&chat_request).unwrap();

        let response = client
            .post(url)
            .headers(headers)
            .body(request_body)
            .send()
            .await;

        let response = match response {
            Ok(response) => response.text().await,
            Err(e) => {
                error!("Error: {}", e);
                return "Error".to_string();
            }
        };

        let response_text = response.unwrap();

        let response_object = parse_response(&response_text).unwrap();

        response_object.choices[0].message.content.clone()
    }
}
