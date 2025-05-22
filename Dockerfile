# Stage 1: Build the Go application
FROM golang:1.23-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy Go module files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the rest of the application code
COPY . .

# Build the Go application
RUN CGO_ENABLED=0 GOOS=linux go build -o ollama-openai-proxy .

# Stage 2: Create a lightweight runtime image
FROM alpine:latest

# Set the working directory
WORKDIR /app

# Copy the built binary from the builder stage
COPY --from=builder /app/ollama-openai-proxy .

# Expose the port your app runs on (default 3033)
EXPOSE 8080

# Define environment variables with default values if needed
ENV NODE_ENV=production
ENV OPENAI_API_BASE_URL="https://api.openai.com"
ENV OPENAI_ALLOWED_MODELS=""

# Set the entrypoint to run the application
ENTRYPOINT ["./ollama-openai-proxy"]