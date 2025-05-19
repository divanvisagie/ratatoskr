# Ratatoskr

A lightweight Telegram <-> Kafka bridge written in **Deno 2+**, designed to decouple message ingestion from processing logic.

![Logo](docs/logo-256.png)

## 🚀 Features

* Uses [`grammY`](https://grammy.dev/) for Telegram bot integration
* Uses [`kafkajs`](https://kafka.js.org/) for Kafka connectivity
* Forwards full Telegram update objects to Kafka
* Listens for outbound messages on a Kafka topic and sends them back to Telegram
* Minimal, event-driven, and easy to extend

## 📦 Prerequisites

* [Deno 2+](https://deno.com/manual@v1.35.0/introduction)
* A Kafka broker (default: `localhost:9092`)
* A Telegram bot token from [@BotFather](https://t.me/BotFather)

## ⚙️ Setup

1. **Install dependencies:**

   ```sh
   deno add npm:grammy npm:kafkajs
   ```

2. **Set environment variables:**

   * `TELEGRAM_BOT_TOKEN` (**required**)
   * `KAFKA_BROKER` (optional, default: `localhost:9092`)
   * `KAFKA_TOPIC_IN` (optional, default: `com.sectorflabs.ratatoskr.in`)
   * `KAFKA_TOPIC_OUT` (optional, default: `com.sectorflabs.ratatoskr.out`)

   You can place these in a `.env` file or export them in your shell.

3. **Run the bot:**

   ```sh
   deno run -A main.ts
   ```

## 🔄 Development

To auto-reload on file changes:

```sh
deno run --watch -A main.ts
```

## 🧱 Project Structure

| File        | Purpose                                |
| ----------- | -------------------------------------- |
| `main.ts`   | Entry point, sets up the bot and Kafka |
| `deno.json` | Deno config (imports, permissions)     |

---

## 📤 Kafka Message Format

**Incoming Topic:** `com.sectorflabs.ratatoskr.in`

```json
{
  "update": { ... },  // Full Telegram update object
  "received_at": "2025-05-19T12:34:56.789Z"
}
```

**Outgoing Topic:** `com.sectorflabs.ratatoskr.out`

```json
{
  "chat_id": 123456789,
  "text": "Hello from Denny!"
}
```

## 🧠 Why Ratatoskr?

Inspired by the mythical squirrel that relays messages across realms, Ratatoskr is built to relay messages between users and intelligent systems, using Kafka as the spine.
