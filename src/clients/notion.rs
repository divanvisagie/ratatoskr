//allow dead code in this file
#![allow(dead_code)]
use chrono::Utc;
use reqwest::Client;
use serde::{Deserialize, Serialize};
use serde_json::json;
use std::env;

#[derive(Debug, Serialize, Deserialize)]
struct Notion {
    journal_page_id: String,
    token: String,
}

#[derive(Debug, Serialize, Deserialize, Clone)]
struct ResultObject {
    object: String,
    id: String,
    created_time: String,
    url: String,
}

#[derive(Debug, Serialize, Deserialize)]
struct TodaysPageResult {
    object: String,
    results: Vec<ResultObject>,
}

impl Notion {
    fn new() -> Self {
        let token = env::var("NOTION_TOKEN").expect("NOTION_TOKEN environment variable not set");
        let journal_page_id =
            env::var("NOTION_JOURNAL_DB").expect("NOTION_JOURNAL_DB environment variable not set");

        Notion {
            journal_page_id,
            token,
        }
    }

    async fn do_request_to_notion(
        &self,
        method: &str,
        url: &str,
        body: Option<Vec<u8>>,
    ) -> Result<Vec<u8>, Box<dyn std::error::Error>> {
        let client = Client::new();
        let request_builder = client.request(reqwest::Method::from_bytes(method.as_bytes())?, url);

        let request_builder = match body {
            Some(body) => request_builder.body(body),
            None => request_builder,
        };

        let request_builder = request_builder
            .header("Content-Type", "application/json")
            .header("Authorization", format!("Bearer {}", self.token))
            .header("Notion-Version", "2022-06-28");

        let response = request_builder.send().await?;
        let response_body = response.bytes().await?;

        Ok(response_body.to_vec())
    }

    async fn get_todays_page(&self) -> Result<TodaysPageResult, Box<dyn std::error::Error>> {
        let today = Utc::now().format("%Y-%m-%d").to_string();

        let request_body = json!({
            "filter": {
                "property": "Date",
                "date": {
                    "on_or_after": today
                }
            }
        });

        let json_body = serde_json::to_vec(&request_body)?;
        let url = format!(
            "https://api.notion.com/v1/databases/{}/query",
            self.journal_page_id
        );

        let body = self
            .do_request_to_notion("POST", &url, Some(json_body))
            .await?;
        let response: TodaysPageResult = serde_json::from_slice(&body)?;

        Ok(response)
    }

    async fn add_link_to_todays_page(
        &self,
        link: &str,
        summary: &str,
    ) -> Result<ResultObject, Box<dyn std::error::Error>> {
        let todays_page = self.get_todays_page().await?;

        let result = if todays_page.results.is_empty() {
            self.create_page_for_today(&["Ratatoskr"]).await? // Assuming create_page_for_today function exists
        } else {
            todays_page.results[0].clone()
        };

        let request_body = json!({
            "children": [
                {
                    "object": "block",
                    "type": "paragraph",
                    "paragraph": {
                        "rich_text": [
                            {
                                "type": "text",
                                "text": {
                                    "content": summary
                                }
                            }
                        ]
                    }
                },
                {
                    "object": "block",
                    "type": "bookmark",
                    "bookmark": {
                        "url": link
                    }
                }
            ]
        });

        let json_body = serde_json::to_vec(&request_body)?;
        let url = format!("https://api.notion.com/v1/blocks/{}/children", result.id);

        self.do_request_to_notion("PATCH", &url, Some(json_body))
            .await?;

        Ok(result)
    }

    async fn create_page_for_today(
        &self,
        tags: &[&str],
    ) -> Result<ResultObject, Box<dyn std::error::Error>> {
        let day_of_week = Utc::now().format("%A").to_string();
        let today = Utc::now().format("%Y-%m-%d").to_string();

        let tag_options = tags
            .iter()
            .map(|tag| json!({ "name": tag }))
            .collect::<Vec<_>>();

        let request_body = json!({
            "parent": {
                "database_id": self.journal_page_id
            },
            "properties": {
                "Tags": {
                    "multi_select": tag_options
                },
                "Date": {
                    "date": {
                        "start": today
                    }
                },
                "Name": {
                    "title": [
                        {
                            "text": {
                                "content": day_of_week
                            }
                        }
                    ]
                }
            }
        });

        let json_body = serde_json::to_vec(&request_body)?;
        let url = "https://api.notion.com/v1/pages";

        let body = self
            .do_request_to_notion("POST", url, Some(json_body))
            .await?;
        let response: ResultObject = serde_json::from_slice(&body)?;

        Ok(response)
    }
}
