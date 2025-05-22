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