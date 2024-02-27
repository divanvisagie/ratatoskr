use std::vec;

use async_trait::async_trait;
use regex::Regex;

use crate::{message_types::ResponseMessage, RequestMessage};

use super::Capability;

pub struct SummaryCapability {}

fn is_link(string: &str) -> bool {
    let url_regex = Regex::new(r#"(?i)\b((?:https?://|www\d{0,3}[.]|[a-z0-9.\-]+[.][a-z]{2,4}/)(?:[^\s()<>]+|\(([^\s()<>]+|(\([^\s()<>]+\)))*\))+(?:\(([^\s()<>]+|(\([^\s()<>]+\)))*\)|[^\s`!()\[\]{};:'".,<>?«»“”‘’]))"#).unwrap();
    url_regex.is_match(string)
}

#[async_trait]
impl Capability for SummaryCapability {
    async fn check(&mut self, message: &RequestMessage) -> f32 {
        if is_link(&message.text) {
            return 1.0;
        }
        0.0
    }

    async fn execute(&mut self, message: &RequestMessage) -> ResponseMessage {
        let summary = format!("Summary of {} goes here", message.text);

        let options = vec!["Save".to_string(), "Discuss".to_string()];

        ResponseMessage::new_with_options(summary, options)
    }
}

impl SummaryCapability {
    pub fn new() -> Self {
        SummaryCapability {}
    }
}

#[cfg(test)]
mod tests {

    use super::*;

    #[tokio::test]
    async fn test_execute() {
        let mut summary_capability = SummaryCapability::new();
        let message = RequestMessage {
            text: "https://www.google.com".to_string(),
            username: "test".to_string(),
            context: Vec::new(),
            embedding: Vec::new(),
        };
        let response = summary_capability.execute(&message).await.clone();
        assert_eq!(response.text, "Summary of https://www.google.com goes here");
    }

    #[tokio::test]
    async fn test_check_when_given_link() {
        let mut summary_capability = SummaryCapability::new();
        let message = RequestMessage {
            text: "https://www.google.com".to_string(),
            username: "test".to_string(),
            context: Vec::new(),
            embedding: Vec::new(),
        };
        let score = summary_capability.check(&message).await;
        assert_eq!(score, 1.0);
    }

    #[tokio::test]
    async fn test_check_when_given_non_link() {
        let mut summary_capability = SummaryCapability::new();
        let message = RequestMessage {
            text: "Hello".to_string(),
            username: "test".to_string(),
            context: Vec::new(),
            embedding: Vec::new(),
        };
        let score = summary_capability.check(&message).await;
        assert_eq!(score, 0.0);
    }
}
