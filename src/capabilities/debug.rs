use async_trait::async_trait;

use crate::{message_types::ResponseMessage, RequestMessage};

use super::Capability;

pub struct DebugCapability {}

#[async_trait]
impl Capability for DebugCapability {
    async fn check(&mut self, message: &RequestMessage) -> f32 {
        if message.text == "/debug" {
            1.0
        } else {
            0.0
        }
    }

    async fn execute(&mut self, _message: &RequestMessage) -> ResponseMessage {
        let keyboard_functions = vec!["Memory Dump".to_string(), "Memory Clear".to_string()];
        ResponseMessage::new_with_options(
            "I've sent you some debug options, you should see the buttons below.".to_string(),
            keyboard_functions,
        )
    }
}

impl DebugCapability {
    pub fn new() -> Self {
        DebugCapability {}
    }
}
