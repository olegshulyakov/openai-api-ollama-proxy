![](./assets/banner.png)

# Server to use Enchanted app with OpenAI API

This is a simple server that can be used to interact with the OpenAI API via Ollama API. Made to work with [Enchanted](https://github.com/gluonfield/enchanted).

## Why

Enchanted app is a great tool to chatting with self hosted AI models. It supports Ollama API.

OpenAI API is different: it has other endpoints and requests format.

This server will convert a couple of endpoints mimic the Ollama API and converts data to work with OpenAI API.

## To do

- [ ] Add tests
- [ ] Add correct error handling

## How to use

### Docker

```
docker run -d \
  --name openai-api-server-for-enchanted \
  -e OPENAI_API_BASE_URL=https://openrouter.ai/api \
  -e OPENAI_ALLOWED_MODELS= \
  -p 3033:3033 \
  --restart unless-stopped \
  ghcr.io/talyguryn/openai-api-server-for-enchanted-app:latest
```

### Docker Compose

```yml
version: "3.4"
services:
  openai-api-server-for-enchanted:
    image: ghcr.io/talyguryn/openai-api-server-for-enchanted-app:latest
    container_name: openai-api-server-for-enchanted
    mem_limit: 64m
    environment:
      - OPENAI_API_BASE_URL=https://openrouter.ai/api
      - OPENAI_ALLOWED_MODELS=
    ports:
      - 3033:3033
    restart: unless-stopped
```

## How to develop

Clone the repository and install dependencies:

```bash
npm i
```

If you need to override vars create a `.env` file as a copy of `.env.example` and fill required variables.

Run the server:

```bash
npm start
```

## ENV variables

You can set the following environment variables:

| Env                     | Description                            | Default                  | Example                  |
| ----------------------- | -------------------------------------- | ------------------------ | ------------------------ |
| `OPENAI_API_BASE_URL`   | Base URL for the OpenAI API            | `https://api.openai.com` | `https://api.openai.com` |
| `OPENAI_ALLOWED_MODELS` | Comma separated list of allowed models | None                     | `gpt-3.5-turbo,gpt-4o`   |

All variables are optional.

## Server Endpoints

On each request Enchanted app will send `Authorization` header with the Bearer token. So you can use it to authenticate requests to OpenAI API.

### HEAD `/`

Health check.

Ollama returns `200` with no body.

### GET `/api/tags`

Get list of available models.

Example Ollama response:

```json
{
  "models": [
    {
      "name": "deepseek-r1:7b",
      "model": "deepseek-r1:7b",
      "modified_at": "2025-03-23T07:52:42.164862186Z",
      "size": 4683075271,
      "digest": "0a8c266910232fd3291e71e5ba1e058cc5af9d411192cf88b6d30e92b6e73163",
      "details": {
        "parent_model": "",
        "format": "gguf",
        "family": "qwen2",
        "families": ["qwen2"],
        "parameter_size": "7.6B",
        "quantization_level": "Q4_K_M"
      }
    },
    {
      "name": "llama3.2:3b",
      "model": "llama3.2:3b",
      "modified_at": "2025-03-23T07:52:41.339842922Z",
      "size": 2019393189,
      "digest": "a80c4f17acd55265feec403c7aef86be0c25983ab279d83f3bcd3abbcb5b8b72",
      "details": {
        "parent_model": "",
        "format": "gguf",
        "family": "llama",
        "families": ["llama"],
        "parameter_size": "3.2B",
        "quantization_level": "Q4_K_M"
      }
    }
  ]
}
```

Example OpenAI response on `/v1/models`:

```json
{
  "object": "list",
  "data": [
    {
      "id": "gpt-4",
      "object": "model",
      "created": 1687882411,
      "owned_by": "openai"
    },
    {
      "id": "o1-pro",
      "object": "model",
      "created": 1742251791,
      "owned_by": "system"
    }
  ]
}
```

### POST `/api/chat`

Chat with the model. Streaming allowed.

Request:

```json
{
  "messages": [{ "content": "hey", "images": [], "role": "user" }],
  "model": "deepseek-r1:1.5b",
  "stream": true,
  "options": { "temperature": 0 }
}
```

Ollama response chunks:

```json
{
  model: "deepseek-r1:1.5b",
  created_at: "2025-03-22T17:01:31.748659Z",
  message: { role: "assistant", content: "\n\n" },
  done: false,
}
{
  model: "deepseek-r1:1.5b",
  created_at: "2025-03-22T17:01:37.788115Z",
  message: { role: "assistant", content: "" },
  done_reason: "stop",
  done: true,
  total_duration: 12801095381,
  load_duration: 4858364600,
  prompt_eval_count: 4,
  prompt_eval_duration: 1511149623,
  eval_count: 41,
  eval_duration: 6429075170,
}
```

OpenAI response chunks on `/v1/chat/completions`:

```
data: {
  id: "chatcmpl-BDwZHr1gDRCmTsBDHgLXmlB90fRhU",
  object: "chat.completion.chunk",
  created: 1742663099,
  model: "gpt-4o-2024-08-06",
  service_tier: "default",
  system_fingerprint: "fp_90d33c15d4",
  choices: [
    {
      index: 0,
      delta: { content: "." },
      logprobs: null,
      finish_reason: null,
    },
  ],
}
data: {
  id: "chatcmpl-BDwZHr1gDRCmTsBDHgLXmlB90fRhU",
  object: "chat.completion.chunk",
  created: 1742663099,
  model: "gpt-4o-2024-08-06",
  service_tier: "default",
  system_fingerprint: "fp_90d33c15d4",
  choices: [{ index: 0, delta: {}, logprobs: null, finish_reason: "stop" }],
}
data: [DONE]
```
