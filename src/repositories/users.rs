use async_trait::async_trait;
use std::env;
use tracing::warn;

#[async_trait]
pub trait UserRepository: Send + Sync {
    async fn get_usernames(&mut self) -> Vec<String>;
}

pub struct FsUserRepository {}

impl FsUserRepository {
    pub fn new() -> Self {
        FsUserRepository {}
    }
}

#[async_trait]
impl UserRepository for FsUserRepository {
    async fn get_usernames(&mut self) -> Vec<String> {
        // get admin user from env
        let admin_user = env::var("TELEGRAM_ADMIN").unwrap_or_else(|_| {
            panic!("TELEGRAM_ADMIN must be set");
        });
        let telegram_users = env::var("TELEGRAM_USERS").unwrap_or_else(|_| {
            warn!("TELEGRAM_USERS not set, using only admin user");
            "".to_string()
        });
        // split the string into a vector of strings
        let telegram_users: Vec<String> =
            telegram_users.split(",").map(|s| s.to_string()).collect();

        let mut usernames = vec![admin_user];
        usernames.extend(telegram_users);
        usernames
    }
}
