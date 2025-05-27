# Ollama to OpenAI Proxy

This Spring Boot application acts as a proxy, translating requests in Ollama format to an OpenAI-compatible API and forwarding them. It then translates the OpenAI response back to the Ollama format.

## Features

*   **Ollama to OpenAI Format Translation:** Converts request and response formats between Ollama and OpenAI.
*   **Custom OpenAI Server URL:** Allows configuration of the target OpenAI-compatible server URL.
*   **Model Filtering:** Supports filtering of allowed models using a regular expression.
*   **Access Token Forwarding:** Securely forwards the user-provided access token (from the `Authorization` header) to the target OpenAI server. The token is not stored on the proxy.

## Configuration

The application is configured primarily via environment variables. Properties in `src/main/resources/application.properties` can serve as fallbacks or for local development overrides.

| Environment Variable        | `application.properties` Key | Description                                                                 | Default (in Dockerfile)       |
| --------------------------- | ---------------------------- | --------------------------------------------------------------------------- | ----------------------------- |
| `PROXY_OPENAI_SERVER_URL`   | `proxy.openaiServerUrl`      | The base URL of the target OpenAI-compatible server (e.g., `https://api.openai.com/v1`). | `https://api.openai.com/v1` |
| `PROXY_MODEL_FILTER_REGEX`  | `proxy.modelFilterRegex`     | A regex to filter allowed model names (e.g., `^gpt-.*`, `^mistral-.*$`).     | `.*` (allow all)            |
| `SERVER_PORT`               | `server.port`                | The port on which the proxy application will run.                           | `8080`                        |

## Prerequisites

*   Java 17 or later
*   Maven 3.6.x or later (for building from source)
*   Docker (for running as a container)

## Building and Running

### Running Locally (from source)

1.  **Clone the repository:**
    ```bash
    git clone <repository-url>
    cd <repository-directory>
    ```
2.  **Configure (Optional):**
    You can set environment variables or modify `src/main/resources/application.properties`.
    ```bash
    export PROXY_OPENAI_SERVER_URL="your_openai_compatible_server_url"
    export PROXY_MODEL_FILTER_REGEX="^your-model-prefix.*"
    ```
3.  **Run the application using Maven:**
    ```bash
    mvn spring-boot:run
    ```
    The application will start on port 8080 by default.

### Building the JAR

To build the executable JAR file:
```bash
mvn package -DskipTests
```
The JAR file will be created in the `target/` directory.

### Running with Docker

1.  **Build the Docker image:**
    From the project root directory (where the `Dockerfile` is located):
    ```bash
    docker build -t ollama-openai-proxy:latest .
    ```

2.  **Run the Docker container:**
    ```bash
    docker run -p 8080:8080 \
      -e PROXY_OPENAI_SERVER_URL="https://your-openai-compatible-server.com/v1" \
      -e PROXY_MODEL_FILTER_REGEX="^gpt-.*" \
      ollama-openai-proxy:latest
    ```
    Adjust the environment variables and port mapping as needed.

## API Usage

The primary endpoint is:

*   **`POST /api/v1/chat`**

    This endpoint expects a JSON request body in a format similar to Ollama's chat or completion requests. It also requires an `Authorization` header containing the Bearer token for the target OpenAI service.

    **Request Example (`OllamaRequest` format):**
    ```json
    {
      "model": "your-target-model",
      "prompt": "What is the capital of France?"
    }
    ```
    *(Note: The DTOs support more fields, this is a basic example)*

    **Headers:**
    *   `Content-Type: application/json`
    *   `Authorization: Bearer YOUR_OPENAI_API_KEY`

    **Response Example (`OllamaResponse` format):**
    ```json
    {
      "model": "your-target-model", // Or the model returned by OpenAI
      "response": "The capital of France is Paris."
      // Other fields like 'done', 'context' might be added based on actual mapping
    }
    ```

## Development

(Details about setting up development environment, running tests, etc., can be added here if needed.)
Unit tests: `mvn test`
Integration tests: Verify during the `mvn verify` phase or run specific test classes from your IDE.

---

This README provides a good starting point for users of the proxy.
