use std::{env, fs::File, io::{Read, Write}, sync::Arc};

use async_openai::{
    types::{CreateImageRequestArgs, Image, ImageModel, ImageSize, ResponseFormat},
    Client,
};
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
    async fn generate_image(&self, text: String) -> Result<Vec<u8>, ()> {
        let client = Client::new();

        let request = CreateImageRequestArgs::default()
            .prompt(text)
            .n(1)
            .response_format(ResponseFormat::Url)
            .size(ImageSize::S1024x1024)
            .user("async-openai")
            .model(ImageModel::DallE3)
            .build();

        let request = match request {
            Ok(request) => request,
            Err(e) => {
                error!("Error creating request: {}", e);
                return Err(());
            }
        };
        let response = client.images().create(request).await;
        let response = match response {
            Ok(response) => response,
            Err(e) => {
                error!("Error in response: {}", e);
                return Err(());
            }
        };

        // &Arc<Image> to bytes
        let paths = response.save("/tmp/ratatoskr").await;
        let paths = match paths {
            Ok(paths) => paths,
            Err(e) => {
                error!("Error saving image: {}", e);
                return Err(());
            }
        };

        let path = paths.first();
        let path = match path {
            Some(path) => path,
            None => {
                error!("No path found");
                return Err(());
            }
        };
        let mut file = File::open(path).unwrap();
        let mut buffer = Vec::new();
        file.read_to_end(&mut buffer);
        Ok(buffer)
    }
}

async fn parse_response(response_text: &str) -> Result<Vec<u8>, ()> {
    // first convert to response type and get the url
    let response_object: DalleResponse = serde_json::from_str(&response_text).unwrap();

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
            return Err(());
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
            return Err(());
        }
    };

    info!("Image downloaded successfully {}", image);

    let image = base64::decode(image).unwrap();

    // also save to disk
    // let mut file = File::create("/tmp/image.png")?;
    // file.write_all(&image)?;

    Ok(image.to_vec())
}
