// Deno 2+ Telegram <-> Kafka bridge
// Requires: npm:grammy, npm:kafkajs
// Ensure you have a .env file or set TELEGRAM_BOT_TOKEN in your environment

import { Bot } from "npm:grammy";
import { Kafka, logLevel } from "npm:kafkajs";

// Load environment variables (Deno 2+)
const TELEGRAM_BOT_TOKEN = Deno.env.get("TELEGRAM_BOT_TOKEN");
if (!TELEGRAM_BOT_TOKEN) throw new Error("TELEGRAM_BOT_TOKEN not set");

const KAFKA_BROKER = Deno.env.get("KAFKA_BROKER") ?? "localhost:9092";
const KAFKA_IN = Deno.env.get("KAFKA_IN_TOPIC") ??
  "com.sectorflabs.ratatoskr.in";
const KAFKA_OUT = Deno.env.get("KAFKA_OUT_TOPIC") ??
  "com.sectorflabs.ratatoskr.out";

const bot = new Bot(TELEGRAM_BOT_TOKEN);

// Kafka setup
const kafka = new Kafka({
  clientId: "deno-bot",
  brokers: [KAFKA_BROKER],
  logLevel: logLevel.ERROR,
});
const producer = kafka.producer();
const consumer = kafka.consumer({ groupId: "denny-bot-consumer" });

// Kafka → Telegram
async function kafkaToTelegram() {
  await consumer.connect();
  await consumer.subscribe({ topic: KAFKA_OUT, fromBeginning: false });
  await consumer.run({
    eachMessage: async ({ message }) => {
      if (!message.value) return;
      try {
        const out = JSON.parse(message.value.toString());
        if (typeof out.chat_id === "number" && typeof out.text === "string") {
          await bot.api.sendMessage(out.chat_id, out.text);
        }
      } catch (_) {}
    },
  });
}

// Telegram → Kafka
bot.on("message:text", async (ctx) => {
  const msg = ctx.message;
  const json = JSON.stringify(msg);
  const key = String(msg.chat.id);
  try {
    await producer.send({
      topic: KAFKA_IN,
      messages: [{ key, value: json }],
    });
    if (Deno.env.get("DEBUG")) {
      console.log("Kafka message sent:", { key, value: json });
      await ctx.reply("✅ Message forwarded to Kafka.");
    }
  } catch (e) {
    await ctx.reply(`Kafka error: ${e}`);
  }
});

// Main
async function main() {
  await producer.connect();
  kafkaToTelegram(); // don't await, run in background
  await bot.start();
}

if (import.meta.main) {
  main();
}
