use std::{env, fs::File, io::Write};

use anyhow::Result;
use async_trait::async_trait;
use reqwest::header;
use serde::{Deserialize, Serialize};
use tracing::{error, info};

use super::ImageGenerationClient;

pub struct DalleClient {}

#[derive(Debug, Serialize, Deserialize)]
struct DalleRequest {
    model: String,
    prompt: String,
    n: usize,
    size: String,
}

/// {
///   "created": 1715298390,
///   "data": [
///     {
///       "revised_prompt": "A Siamese cat characterized by its cream-colored coat and vivid blue almond-shaped eyes. The cat's distinctive coat furthers darkens around its ears, face, paws, and tail, displaying the typical pointed pattern of the breed. The full-bodied and muscular feline wears an expression of serene curiosity, its tail casually curling around its body.",
///       "url": "https://oaidalleapiprodscus.blob.core.windows.net/private/org-X57IW9msRf3rtj0FNayeakQR/user-G39nQfXoeEVObYhIVH9idoVK/img-w3FNekGhpRrOVQNmYoM9Hj1O.png?st=2024-05-09T22%3A46%3A30Z&se=2024-05-10T00%3A46%3A30Z&sp=r&sv=2021-08-06&sr=b&rscd=inline&rsct=image/png&skoid=6aaadede-4fb3-4698-a8f6-684d7786b067&sktid=a48cca56-e6da-484e-a814-9c849652bcb3&skt=2024-05-09T21%3A38%3A23Z&ske=2024-05-10T21%3A38%3A23Z&sks=b&skv=2021-08-06&sig=Wy9a33SOyG6678bHvoBn4D3bZtB5V0YK%2B9UZt%2B3w/Qc%3D"
///     }
///   ]
/// }
#[derive(Deserialize)]
struct DalleResponse {
    data: Vec<DalleDataItem>,
}

#[derive(Serialize, Deserialize)]
struct DalleDataItem {
    revised_prompt: String,
    url: String,
}

/// curl https://api.openai.com/v1/images/generations \
///   -H "Content-Type: application/json" \
///   -H "Authorization: Bearer $OPENAI_API_KEY" \
///   -d '{
///     "model": "dall-e-3",
///     "prompt": "a white siamese cat",
///     "n": 1,
///     "size": "1024x1024"
///   }'
impl DalleClient {
    pub fn new() -> Self {
        DalleClient {}
    }
}

#[async_trait]
impl ImageGenerationClient for DalleClient {
    async fn generate_image(&self, text: String) -> Result<Vec<u8>> {
        let api_key =
            env::var("OPENAI_API_KEY").expect("Missing OPENAI_API_KEY environment variable");

        let client = reqwest::Client::new();
        let url = "https://api.openai.com/v1/images/generations";

        let mut headers = header::HeaderMap::new();
        headers.insert(
            header::CONTENT_TYPE,
            header::HeaderValue::from_static("application/json"),
        );
        headers.insert(
            header::AUTHORIZATION,
            header::HeaderValue::from_str(&format!("Bearer {}", api_key)).unwrap(),
        );

        let request_body = serde_json::to_string(&DalleRequest {
            model: "dall-e-3".to_string(),
            prompt: text,
            n: 1,
            size: "1024x1024".to_string(),
        });

        let request_body = match request_body {
            Ok(body) => body,
            Err(e) => panic!("Error serializing request body: {}", e),
        };

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
                return Err(e.into());
            }
        };

        let response_text = match response {
            Ok(response) => response,
            Err(e) => {
                error!("Error after trying to unwrap: {}", e);
                return Err(e.into());
            }
        };

        let response_object = parse_response(&response_text).await?;

        Ok(response_object)
    }
}

async fn parse_response(response_text: &str) -> Result<Vec<u8>> {
    // first convert to response type and get the url
    let response_object: DalleResponse = serde_json::from_str(&response_text)?;

    // then download the image from the url
    let url = response_object.data.first().unwrap().url.clone();
    let api_key = env::var("OPENAI_API_KEY").expect("Missing OPENAI_API_KEY environment variable");
    let mut headers = header::HeaderMap::new();
    headers.insert(
        header::AUTHORIZATION,
        header::HeaderValue::from_str(&format!("Bearer {}", api_key)).unwrap(),
    );
    let client = reqwest::Client::new();

    info!("Downloading image from: {}", &url);
    let response = match client.get(url).headers(headers).send().await {
        Ok(response) => response,
        Err(e) => {
            error!("Error: {}", e);
            return Err(e.into());
        }
    };

    let content_type = response
        .headers()
        .get(reqwest::header::CONTENT_TYPE)
        .map(|v| v.to_str().unwrap_or("unknown"));
    info!("Content-Type: {}", content_type.unwrap_or("unknown"));

    let image = match response.text().await {
        Ok(image) => image,
        Err(e) => {
            error!("Error getting bytes: {}", e);
            return Err(e.into());
        }
    };

    info!("Image downloaded successfully {}", image);

    let image = base64::decode(image)?;

    // also save to disk
    // let mut file = File::create("/tmp/image.png")?;
    // file.write_all(&image)?;

    Ok(image.to_vec())
}
