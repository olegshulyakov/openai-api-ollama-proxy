# Ollama-OpenAI Proxy

This is a lightweight proxy server written in GoLang that allows you to interact with the OpenAI API using the Ollama API format. It was primarily developed to work seamlessly with [Enchanted](https://github.com/gluonfield/enchanted), an excellent tool for chatting with self-hosted AI models, which supports the Ollama API.

The main purpose of this proxy is to bridge the gap between the Ollama and OpenAI APIs by translating requests and responses between the two formats. This enables applications designed to communicate with the Ollama API to instead utilize OpenAI-compatible endpoints.

## Features

- Converts Ollama API requests to OpenAI API requests
- Translates OpenAI API responses back into Ollama-compatible formats
- Supports streaming chat completions
- Easy-to-use Docker container for quick deployment

## Requirements

- Docker (for containerized deployment)
- Go 1.21+ (for local development)

## Usage

### Docker Deployment

You can run the proxy as a Docker container using the following command:

```bash
docker run -d \
  --name ollama-openai-proxy \
  -e OPENAI_API_BASE_URL=https://api.openai.com \
  -e OPENAI_ALLOWED_MODELS=gpt-3.5-turbo,gpt-4o \
  -p 8080:8080 \
  --restart unless-stopped \
  ghcr.io/olegshulyakov/ollama-openai-proxy:latest
```

### Docker Compose

Alternatively, use the `docker-compose.yml` file:

```yaml
services:
  ollama-openai-proxy:
    image: ghcr.io/olegshulyakov/ollama-openai-proxy:latest
    container_name: ollama-openai-proxy
    environment:
      - OPENAI_API_BASE_URL=https://api.openai.com
      - OPENAI_ALLOWED_MODELS=gpt-3.5-turbo,gpt-4o
    ports:
      - 8080:8080
    restart: unless-stopped
```

Then start it using:

```bash
docker-compose up -d
```

## Environment Variables

| Env | Description | Default | Example |
|-----|-------------|---------|---------|
| `OPENAI_API_BASE_URL` | Base URL for the OpenAI API | `https://api.openai.com` | `https://openrouter.ai/api` |
| `OPENAI_ALLOWED_MODELS` | Comma-separated list of allowed models | None | `gpt-3.5-turbo,gpt-4o` |

> **Note**: Enchanted sends an `Authorization` header with a Bearer token, which this proxy forwards to the OpenAI API endpoint for authentication (https://github.com/olegshulyakov/ollama-openai-proxy).

## Endpoints Supported

- **GET /api/tags** – Returns a list of available models in Ollama format.
- **POST /api/chat** – Chat with a model, supporting both streaming and non-streaming modes.

For more information on how the translation works between the two APIs, refer to the example payloads in the project's repository.

## Building from Source

If you prefer to build and run the application locally:

1. Clone the repository:

```bash
git clone https://github.com/olegshulyakov/ollama-openai-proxy.git
cd ollama-openai-proxy
```

2. Install dependencies:

```bash
go mod download
```

3. Run the server:

```bash
go run main.go
```

Or build and run the binary:

```bash
go build -o ollama-openai-proxy
./ollama-openai-proxy
```

## Contributing

Feel free to fork the repo, create pull requests, or open issues if you'd like to contribute or enhance functionality.

## License

MIT License – see the [LICENSE](https://github.com/olegshulyakov/ollama-openai-proxy/blob/main/LICENSE) file for details.

---

For more information about the Enchanted app and its integration with Ollama, visit [Enchanted GitHub](https://github.com/gluonfield/enchanted).