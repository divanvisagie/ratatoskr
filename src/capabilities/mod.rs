use crate::{message_types::ResponseMessage, RequestMessage};
use async_trait::async_trait;

pub mod chat;
pub mod debug;
pub mod privacy;
pub mod summarize;

/// A capability is a feature that the bot can perform. Capabilities are chosen
/// by a capability selector which expects to be able to call the `check` method
/// to determine a score between -1.0 and 1.0. The capability with the highest score
/// is chosen to handle the request.
#[async_trait]
pub trait Capability: Send {
    fn get_name(&self) -> String {
        let raw = std::any::type_name::<Self>().to_string();
        let parts: Vec<&str> = raw.split("::").collect();
        parts.last().unwrap().to_string()
    }
    /// Returns a score between -1.0 and 1.0 indicating how well this capability
    /// can handle the request.
    async fn check(&mut self, message: &RequestMessage) -> f32;

    /// Executes the capability and returns a response message.
    async fn execute(&mut self, message: &RequestMessage) -> ResponseMessage;
}

pub fn cosine_similarity(a: &[f32], b: &[f32]) -> f32 {
    let dot_product: f32 = a.iter().zip(b).map(|(x, y)| x * y).sum();
    let magnitude_a: f32 = a.iter().map(|x| x * x).sum::<f32>().sqrt();
    let magnitude_b: f32 = b.iter().map(|x| x * x).sum::<f32>().sqrt();
    dot_product / (magnitude_a * magnitude_b)
}
