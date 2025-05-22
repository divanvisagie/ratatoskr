# Ratatoskr

A lightweight Telegram <-> Kafka bridge written in **Rust**, designed to decouple message ingestion from processing logic.

![Logo](docs/logo-256.png)

[![GitHub Repository](https://img.shields.io/badge/GitHub-Repository-blue.svg)](https://github.com/yourusername/ratatoskr)

## 🚀 Features

* Uses [`teloxide`](https://github.com/teloxide/teloxide) for Telegram bot integration
* Uses [`rdkafka`](https://github.com/fede1024/rust-rdkafka) for Kafka connectivity
* Forwards full Telegram message objects to Kafka
* Listens for outbound messages on a Kafka topic and sends them back to Telegram
* Minimal, event-driven, and easy to extend

## 📦 Prerequisites

* [Rust](https://www.rust-lang.org/tools/install)
* A Kafka broker (default: `localhost:9092`)
* A Telegram bot token from [@BotFather](https://t.me/BotFather)

## ⚙️ Setup

1. **Clone the repository:**

   ```sh
   git clone https://github.com/yourusername/ratatoskr.git
   cd ratatoskr
   ```

   Alternatively, you can download the source code directly from the [GitHub repository](https://github.com/yourusername/ratatoskr/releases).

2. **Set environment variables:**

   * `TELEGRAM_BOT_TOKEN` (**required**)
   * `KAFKA_BROKER` (optional, default: `localhost:9092`)
   * `KAFKA_IN_TOPIC` (optional, default: `com.sectorflabs.ratatoskr.in`)
   * `KAFKA_OUT_TOPIC` (optional, default: `com.sectorflabs.ratatoskr.out`)

   You can place these in a `.env` file or export them in your shell. A `.env.example` file is provided as a template.

3. **Build and run the bot:**

   ```sh
   cargo build --release
   ./target/release/ratatoskr
   ```

   Or simply:
   
   ```sh
   cargo run --release
   ```

## 🔄 Development

For development with auto-reload:

```sh
cargo install cargo-watch
cargo watch -x run
```

To run tests:

```sh
cargo test
```

## 🐳 Containerization

### Using Docker

To containerize the application for deployment:

1. **Using the provided Dockerfile:**

   The project includes a Dockerfile that sets up a multi-stage build for a lightweight container:

   ```dockerfile
   FROM rust:1.75-slim as builder
   WORKDIR /app
   COPY . .
   RUN cargo build --release

   FROM debian:bullseye-slim
   RUN apt-get update && apt-get install -y libssl-dev ca-certificates && rm -rf /var/lib/apt/lists/*
   WORKDIR /app
   COPY --from=builder /app/target/release/ratatoskr /app/ratatoskr
   COPY --from=builder /app/.env* /app/
   ENV RUST_LOG=info
   CMD ["./ratatoskr"]
   ```

2. **Build and run the Docker image:**

   ```sh
   docker build -t ratatoskr:latest .
   docker run -d --name ratatoskr \
     -e TELEGRAM_BOT_TOKEN=your_token_here \
     -e KAFKA_BROKER=kafka:9092 \
     ratatoskr:latest
   ```

### Using Docker Compose

For a complete development environment with Kafka, Zookeeper, and Kafdrop (a Kafka UI):

1. **Run with docker-compose:**

   ```sh
   # Make sure TELEGRAM_BOT_TOKEN is set in your environment or .env file
   docker-compose up -d
   ```

2. **Access services:**
   - Ratatoskr: Running in container
   - Kafka: localhost:9092
   - Kafdrop (Kafka UI): http://localhost:9000


---

## 📤 Kafka Message Format

**Incoming Topic:** `com.sectorflabs.ratatoskr.in`

```json
{
  "message_id": 123,
  "from": { "id": 456, "first_name": "User" },
  "chat": { "id": 789, "type": "private" },
  "date": 1678901234,
  "text": "Hello bot!"
  // Full Telegram Message object serialized as JSON
}
```

**Outgoing Topic:** `com.sectorflabs.ratatoskr.out`

```json
{
  "chat_id": 123456789,
  "text": "Hello from Ratatoskr!"
}
```

## 🧠 Why Ratatoskr?

Inspired by the mythical squirrel that relays messages across realms, Ratatoskr is built to relay messages between users and intelligent systems, using Kafka as the messaging backbone.

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the project
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📃 License

This project is licensed under the GNU General Public License v2.0 (GPL-2.0) - see the LICENSE.md file for details.
