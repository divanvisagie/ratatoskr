#![allow(unused)]
use anyhow::Result;

use async_trait::async_trait;
pub use dalle::DalleClient;

pub mod dalle;

pub enum ImageGenerationClientImpl {
    Dalle(DalleClient),
    Mock(MockImageGenerationClient),
}

#[async_trait]
impl ImageGenerationClient for ImageGenerationClientImpl {
    async fn generate_image(&self, text: String) -> Result<Vec<u8>> {
        match self {
            ImageGenerationClientImpl::Dalle(client) => client.generate_image(text).await,
            ImageGenerationClientImpl::Mock(client) => client.generate_image(text).await,
        }
    }
}

#[async_trait::async_trait]
pub trait ImageGenerationClient {
    async fn generate_image(&self, text: String) -> Result<Vec<u8>>;
}

pub struct MockImageGenerationClient {}

impl MockImageGenerationClient {

    fn new() -> Self {
        MockImageGenerationClient {}
    }
}

#[async_trait::async_trait]
impl ImageGenerationClient for MockImageGenerationClient {
    async fn generate_image(&self, text: String) -> Result<Vec<u8>> {
        Ok(vec![0, 0, 0])
    }
}


#[cfg(test)]
mod tests {
    use super::*;

    #[tokio::test]
    async fn test_mock_image_generation() {
        let client = ImageGenerationClientImpl::Mock(MockImageGenerationClient::new());
        let result = client.generate_image("test".to_string()).await;
        assert_eq!(result.unwrap(), vec![0, 0, 0]);
    }
}
