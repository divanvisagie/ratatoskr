use std::env;

use async_trait::async_trait;

use crate::{message_types::ResponseMessage, repositories::users::UserRepository, RequestMessage};

use super::Layer;
pub struct SecurityLayer<T: Layer, R: UserRepository> {
    next: T,
    admin: String,
    user_repository: R,
    allowed_groups: Vec<String>,
}

#[async_trait]
impl<T: Layer, R: UserRepository> Layer for SecurityLayer<T, R> {
    async fn execute(&mut self, message: &mut RequestMessage) -> ResponseMessage {
        match message.chat_type.clone() {
            crate::message_types::ChatType::Private => {
                let users = self.user_repository.get_usernames().await;

                if users.contains(&message.username) || message.username == self.admin {
                    self.next.execute(message).await
                } else {
                    return ResponseMessage::new(format!(
                        "You need to contact @{} to use this bot.",
                        self.admin
                    ));
                }
            }
            crate::message_types::ChatType::Group(group) => {
                if self.allowed_groups.contains(&group) {
                    self.next.execute(message).await
                } else {
                    return ResponseMessage::new(format!(
                        "You need to contact @{} to use this bot.",
                        self.admin
                    ));
                }
            }
        }
    }
}

impl<T: Layer, R: UserRepository> SecurityLayer<T, R> {
    pub fn new(next: T, repo: R) -> Self {
        let admin =
            env::var("TELEGRAM_ADMIN").expect("Missing TELEGRAM_ADMIN environment variable");
        let allowed_groups = env::var("TELEGRAM_ALLOWED_GROUPS")
            .expect("Missing TELEGRAM_ALLOWED_GROUPS environment variable");
        SecurityLayer {
            next,
            admin,
            allowed_groups: allowed_groups.split(",").map(|s| s.to_string()).collect(),
            user_repository: repo,
        }
    }

    #[allow(dead_code)]
    pub fn with_admin(next: T, user_repository: R, admin: String) -> Self {
        SecurityLayer {
            next,
            admin,
            user_repository,
            allowed_groups: Vec::new(),
        }
    }
}

#[cfg(test)]
mod tests {

    use super::*;
    use async_trait::async_trait;

    struct MockRepository {}

    #[async_trait]
    impl UserRepository for MockRepository {
        async fn get_usernames(&mut self) -> Vec<String> {
            vec!["valid_user".to_string()]
        }
    }

    struct MockLayer {}
    #[async_trait]
    impl Layer for MockLayer {
        async fn execute(&mut self, message: &mut RequestMessage) -> ResponseMessage {
            ResponseMessage {
                text: format!("Hello, {}!", message.username),
                bytes: None,
                options: None,
            }
        }
    }

    #[tokio::test]
    async fn test_security_layer_not_allowed() {
        let mock_repo = MockRepository {};
        let mut layer =
            SecurityLayer::with_admin(MockLayer {}, mock_repo, "valid_user".to_string());

        let mut message = RequestMessage {
            text: "Hello".to_string(),
            username: "invalid_user".to_string(),
            context: Vec::new(),
            embedding: Vec::new(),
            chat_type: crate::message_types::ChatType::Private,
        };

        let response = layer.execute(&mut message).await;
        assert_eq!(
            response.text,
            "You need to contact @valid_user to use this bot."
        );
    }

    #[tokio::test]
    async fn test_security_layer_allowed() {
        let mock_repo = MockRepository {};
        let mut layer =
            SecurityLayer::with_admin(MockLayer {}, mock_repo, "valid_user".to_string());

        let mut message = RequestMessage {
            text: "Hello".to_string(),
            username: "valid_user".to_string(),
            context: Vec::new(),
            embedding: Vec::new(),
            chat_type: crate::message_types::ChatType::Private,
        };

        let response = layer.execute(&mut message).await;
        assert_eq!(response.text, "Hello, valid_user!");
    }
}
