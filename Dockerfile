# Use the official Go image as the base image
FROM golang:1.17 as builder

# Set the working directory inside the container
WORKDIR /app

# Copy the go.mod and go.sum files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the entire project
COPY . .

# Build the application
RUN go build -o ratatoskr cmd/ratatoskr/main.go

# Use a minimal image for the final container
FROM gcr.io/distroless/base-debian10

# Copy the built application from the builder stage
COPY --from=builder /app/ratatoskr /ratatoskr

# Expose the port your app runs on
EXPOSE 8080

# Command to run the application
CMD ["/ratatoskr"]
