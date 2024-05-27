use async_trait::async_trait;
use regex::Regex;
use tracing::info;

use crate::clients::chat::{ChatClientImpl, ContextBuilder};
use crate::{
    clients::chat::{ChatClient, Role},
    message_types::ResponseMessage,
    RequestMessage,
};

use super::Capability;
use scraper::{Html, Selector};

pub struct SummaryCapability {
    client: ChatClientImpl,
}

fn is_link(string: &str) -> bool {
    let url_regex = Regex::new(r#"(?i)\b((?:https?://|www\d{0,3}[.]|[a-z0-9.\-]+[.][a-z]{2,4}/)(?:[^\s()<>]+|\(([^\s()<>]+|(\([^\s()<>]+\)))*\))+(?:\(([^\s()<>]+|(\([^\s()<>]+\)))*\)|[^\s`!()\[\]{};:'".,<>?«»“”‘’]))"#).unwrap();
    url_regex.is_match(string)
}

impl SummaryCapability {
    pub fn new(client: ChatClientImpl) -> Self {
        SummaryCapability { client }
    }
}

#[async_trait]
impl Capability for SummaryCapability {
    async fn check(&mut self, message: &RequestMessage) -> f32 {
        if message.text == "Summarize" || message.text == "Nothing" {
            return 1.0;
        }
        if is_link(&message.text) {
            return 1.0;
        }
        -1.0
    }

    async fn execute(&mut self, message: &RequestMessage) -> ResponseMessage {
        if message.text == "Nothing" {
            return ResponseMessage::new(
                "Alright, let me know if you need anything else".to_string(),
            );
        }
        if message.text == "Summarize" {
            let input_message = message
                .context
                .iter()
                .filter(|m| m.role.to_string() == "user")
                .last()
                .unwrap();
            let article_text = fetch_content(&input_message.text)
                .await
                .unwrap_or_else(|_| "".to_string());

            info!("article_text: {}", article_text);
            let prompt = "The following is an article that the user has sent you, send them a brief
                TLDR summary describing any main takeaways that might be useful";
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
        let text = "What would you like to do with this article?".to_string();

        let options = vec!["Summarize".to_string(), "Nothing".to_string()];

        ResponseMessage {
            bytes: None,
            options: Some(options),
            text,
        }
    }
}

fn get_main_article_content(document: &Html) -> String {
    let article_selector = Selector::parse("article, .article, .post, .content").unwrap();
    let mut article_texts = Vec::new();

    for element in document.select(&article_selector) {
        article_texts.push(element.text().collect::<Vec<_>>().join(" "));
    }

    if article_texts.is_empty() {
        match Selector::parse("body") {
            Ok(selector) => {
                for element in document.select(&selector) {
                    article_texts.push(element.text().collect::<Vec<_>>().join(" "));
                }
            }
            Err(_) => return "".to_string(),
        };
    }

    article_texts.join(" ")
}

fn get_meta_description(document: &Html) -> String {
    let meta_description_selector = Selector::parse("meta[name=description]").unwrap();
    let mut article_texts = Vec::new();

    for element in document.select(&meta_description_selector) {
        article_texts.push(element.value().attr("content").unwrap().to_string());
    }

    article_texts.join(" ")
}

async fn fetch_content(url: &str) -> Result<String, String> {
    let html = reqwest::get(url).await.unwrap().text().await.unwrap();
    let document = Html::parse_document(&html);

    let meta_description = get_meta_description(&document);
    let main_article_content = get_main_article_content(&document);

    let summary = vec![url.to_string(), meta_description, main_article_content];
    let summary = summary.join("\n");

    if summary.is_empty() {
        return Err("Could not read the content of the website".to_string());
    }

    Ok(summary)
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::clients::chat::MockChatClient;

    #[tokio::test]
    async fn test_execute() {
        let mock_client = MockChatClient {};
        let client_impl = ChatClientImpl::Mock(mock_client);
        let mut summary_capability = SummaryCapability::new(client_impl);
        let message = RequestMessage {
            chat_id: 99,
            text: "https://divanv.com/post/structured-intelligence/".to_string(),
            username: "test".to_string(),
            context: Vec::new(),
            sent_by: "".to_string(),
            embedding: Vec::new(),
            chat_type: crate::message_types::ChatType::Private,
        };
        let response = summary_capability.execute(&message).await.clone();
        assert_eq!(
            response.text,
            "What would you like to do with this article?"
        );
    }

    #[tokio::test]
    async fn test_check_when_given_link() {
        let mock_client = MockChatClient {};
        let client_impl = ChatClientImpl::Mock(mock_client);
        let mut summary_capability = SummaryCapability::new(client_impl);
        let message = RequestMessage {
            chat_id: 99,
            text: "https://www.google.com".to_string(),
            username: "test".to_string(),
            sent_by: "".to_string(),
            context: Vec::new(),
            embedding: Vec::new(),
            chat_type: crate::message_types::ChatType::Private,
        };
        let score = summary_capability.check(&message).await;
        assert_eq!(score, 1.0);
    }

    #[tokio::test]
    async fn test_check_when_given_non_link() {
        let mock_client = MockChatClient {};
        let client_impl = ChatClientImpl::Mock(mock_client);
        let mut summary_capability = SummaryCapability::new(client_impl);

        let message = RequestMessage {
            chat_id: 99,
            text: "Hello".to_string(),
            username: "test".to_string(),
            context: Vec::new(),
            sent_by: "".to_string(),
            embedding: Vec::new(),
            chat_type: crate::message_types::ChatType::Private,
        };
        let score = summary_capability.check(&message).await;
        assert_eq!(score, 0.0);
    }
}
