use async_trait::async_trait;
use regex::Regex;
use tracing::info;

use crate::clients::chat::ContextBuilder;
use crate::{
    clients::chat::{ChatClient, Role},
    message_types::ResponseMessage,
    RequestMessage,
};

use super::Capability;
use scraper::{Html, Selector};

pub struct SummaryCapability<C: ChatClient> {
    client: C,
}

fn is_link(string: &str) -> bool {
    let url_regex = Regex::new(r#"(?i)\b((?:https?://|www\d{0,3}[.]|[a-z0-9.\-]+[.][a-z]{2,4}/)(?:[^\s()<>]+|\(([^\s()<>]+|(\([^\s()<>]+\)))*\))+(?:\(([^\s()<>]+|(\([^\s()<>]+\)))*\)|[^\s`!()\[\]{};:'".,<>?«»“”‘’]))"#).unwrap();
    url_regex.is_match(string)
}

impl<C: ChatClient> SummaryCapability<C> {
    pub fn new(client: C) -> Self {
        SummaryCapability { client }
    }
}

#[async_trait]
impl<C: ChatClient> Capability for SummaryCapability<C> {
    async fn check(&mut self, message: &RequestMessage) -> f32 {
        if message.text == "Summarize" {
            return 1.0;
        }
        if is_link(&message.text) {
            return 1.0;
        }
        0.0
    }

    async fn execute(&mut self, message: &RequestMessage) -> ResponseMessage {
        if message.text == "Summarize" {
            let input_message = message.context.iter().filter(|m| m.role.to_string() == "user").last().unwrap();
            let article_text = fetch_and_summarize(&input_message.text)
                .await
                .unwrap_or_else(|_| "".to_string());

            info!("article_text: {}", article_text);
            let prompt = "The following is an article that the user has sent you, send them a brief TLDR summary describing any main takeaways that might be useful";
            let mut context = ContextBuilder::new();
            context.add_message(Role::System, prompt.to_string());
            context.add_message(Role::User, message.text.clone());

            // check if article is empty string or just whitepace
            if article_text.trim().is_empty() {
                return ResponseMessage::new("I was unable to read the article".to_string());
            }
            context.add_message(
                Role::System,
                format!("The system then created the summary:\n {} ", article_text),
            );

            let summary = self.client.complete(context.build()).await;
            return ResponseMessage::new(summary);
        }
        let text = "What would you like todo with this article?".to_string();

        // shorten article text to just under what telegram bots can handle
        // let article_text = if article_text.len() > 4000 {
        //     &article_text[..4000]
        // } else {
        //     &article_text
        // };

        let options = vec![
            "Save".to_string(),
            "Summarize".to_string(),
            "Discuss".to_string(),
        ];
        // let options = None

        ResponseMessage {
            bytes: None,
            options: Some(options),
            text,
        }
    }
}

async fn fetch_and_summarize(url: &str) -> Result<String, ()> {
    let html = reqwest::get(url).await.unwrap().text().await.unwrap();
    let document = Html::parse_document(&html);

    // Attempt to find the main article content
    let article_selector = Selector::parse("article, .article, .post, .content").unwrap();
    let mut article_texts = Vec::new();

    for element in document.select(&article_selector) {
        article_texts.push(element.text().collect::<Vec<_>>().join(" "));
    }

    // If still empty use meta description
    if article_texts.is_empty() {
        let meta_description_selector = Selector::parse("meta[name=description]").unwrap();
        for element in document.select(&meta_description_selector) {
            article_texts.push(element.value().attr("content").unwrap().to_string());
        }
    }

    let summary = article_texts.join(" ");

    // If no article content was found, you might fallback to another strategy or return an error
    if summary.is_empty() {
        return Err(());
    }

    Ok(summary)
}

#[cfg(test)]
mod tests {

    use crate::clients::chat::Message;

    use super::*;

    struct MockChatClient {}

    #[async_trait]
    impl ChatClient for MockChatClient {
        async fn complete(&mut self, context: Vec<Message>) -> String {
            "Summary of https://www.google.com goes here".to_string()
        }
    }

    #[tokio::test]
    async fn test_execute() {
        let mock_client = MockChatClient {};
        let mut summary_capability = SummaryCapability::new(mock_client);
        let message = RequestMessage {
            text: "https://divanv.com/post/structured-intelligence/".to_string(),
            username: "test".to_string(),
            context: Vec::new(),
            embedding: Vec::new(),
        };
        let response = summary_capability.execute(&message).await.clone();
        assert_eq!(response.text, "Summary of https://www.google.com goes here");
    }

    #[tokio::test]
    async fn test_check_when_given_link() {
        let mock_client = MockChatClient {};
        let mut summary_capability = SummaryCapability::new(mock_client);
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
        let mock_client = MockChatClient {};
        let mut summary_capability = SummaryCapability::new(mock_client);

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
