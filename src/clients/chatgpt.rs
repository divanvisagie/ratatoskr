use std::{env, fmt};

use reqwest::header;
use serde::{Deserialize, Serialize};
use serde_json::Result;
use tracing::error;
#[derive(Debug, Serialize, Deserialize)]
struct ChatRequest {
    pub model: String,
    pub messages: Vec<Message>,
}

#[derive(Debug, Serialize, Deserialize, Clone)]
struct Message {
    role: String,
    content: String,
}

impl fmt::Display for Message {
    fn fmt(&self, f: &mut fmt::Formatter) -> fmt::Result {
        write!(f, "{}: {}", self.role, self.content)
    }
}

#[derive(Debug, Serialize, Deserialize)]
pub struct ChatResponse {
    pub id: String,
    pub object: String,
    pub created: u64,
    pub model: String,
    usage: Usage,
    choices: Vec<Choice>,
}

#[derive(Debug, Serialize, Deserialize)]
struct Usage {
    prompt_tokens: u64,
    completion_tokens: u64,
    total_tokens: u64,
}

#[derive(Debug, Serialize, Deserialize)]
struct Choice {
    message: Message,
    finish_reason: String,
    index: u64,
}

fn parse_response(json_str: &str) -> Result<ChatResponse> {
    serde_json::from_str(json_str)
}

#[derive(Debug)]
pub struct GptClient {
    messages: Vec<Message>,
}

#[derive(Debug, Serialize, Deserialize, Clone)]
pub enum Role {
    System,
    User,
    Assistant,
}

impl fmt::Display for Role {
    fn fmt(&self, f: &mut fmt::Formatter) -> fmt::Result {
        match self {
            Role::System => write!(f, "system"),
            Role::User => write!(f, "user"),
            Role::Assistant => write!(f, "assistant"),
        }
    }
}

impl GptClient {
    pub fn new() -> Self {
        GptClient {
            messages: Vec::new(),
        }
    }

    pub fn add_message(&mut self, role: Role, text: String) -> &mut Self {
        self.messages.push(Message {
            role: role.to_string(),
            content: text.trim().to_string(),
        });
        self
    }

    //complete method
    pub async fn complete(&mut self) -> String {
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
            messages: self.messages.clone(),
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
