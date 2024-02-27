# Build in special rust container
FROM rust:latest as build

WORKDIR /app

COPY Cargo.toml ./
COPY Cargo.lock ./
COPY src ./src

RUN cargo build --release


# Transfer to debian container for production
FROM rust:latest

WORKDIR /app

ENV TELOXIDE_TOKEN ""
ENV REDIS_URL ""
ENV OPENAI_API_KEY ""

COPY --from=build /app/target/release/ .
RUN chmod +x ./ratatoskr
# Set the entrypoint command for the container
RUN ls
CMD ["./ratatoskr"]
