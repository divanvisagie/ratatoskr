Test response for saving API usage

```rust
use std::collections::HashMap;
use std::sync::Arc;

use async_trait::async_trait;

trait Storage {
    async fn save(&self, key: String, value: String) -> Result<(), String>;
}

struct StorageMock {
    storage: Arc<dyn Storage>,
    saved: Arc<HashMap<String, String>>,
}

#[async_trait]
impl Storage for StorageMock {
    async fn save(&self, key: String, value: String) -> Result<(), String> {
        self.saved.insert(key, value);
        Ok(())
    }
}
```
