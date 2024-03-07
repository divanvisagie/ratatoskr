use std::vec;

use async_trait::async_trait;
use regex::Regex;
use tracing::info;

use crate::{
    clients::chatgpt::{GptClient, Role},
    message_types::ResponseMessage,
    RequestMessage,
};

use super::Capability;
use scraper::{Html, Selector};

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
        let article_text = fetch_and_summarize(&message.text)
            .await
            .unwrap_or_else(|_| "".to_string());

        info!("article_text: {}", article_text);
        let mut  gpt_client = GptClient::new();
        let prompt = "The following is an article that the user has sent you, send them a brief TLDR summary describing any main takeaways that might be useful";
        gpt_client.add_message(Role::System, prompt.to_string());
        gpt_client.add_message(Role::User, message.text.clone());

        // check if article is empty string or just whitepace
        if article_text.trim().is_empty() {
            return ResponseMessage::new("I was unable to read the article".to_string());
        }

        gpt_client.add_message(
            Role::System,
            format!("The system then created the summary:\n {} ", article_text),
        );

        let summary = gpt_client.complete().await;

        // shorten article text to just under what telegram bots can handle
        // let article_text = if article_text.len() > 4000 {
        //     &article_text[..4000]
        // } else {
        //     &article_text
        // };

        let options = vec!["Save".to_string(), "Discuss".to_string()];

        ResponseMessage::new_with_options(summary, options)
    }
}

impl SummaryCapability {
    pub fn new() -> Self {
        SummaryCapability {}
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
