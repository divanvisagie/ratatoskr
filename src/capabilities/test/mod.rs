use super::Capability;
use async_trait::async_trait;

pub struct TestCapability {}

impl TestCapability {
    pub fn new() -> Self {
        TestCapability {}
    }
}

#[async_trait]
impl Capability for TestCapability {
    async fn check(&mut self, message: &crate::message_types::RequestMessage) -> f32 {
        if message.text.to_lowercase().starts_with("/test")
            || message.text.to_lowercase().starts_with("test")
        {
            1.0
        } else {
            -1.0
        }
    }

    async fn execute(
        &mut self,
        _message: &crate::message_types::RequestMessage,
    ) -> crate::message_types::ResponseMessage {
        let res = include_str!("canned_response.md");
        crate::message_types::ResponseMessage::new(res.to_string())
    }
}
