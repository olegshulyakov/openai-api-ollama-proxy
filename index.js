const express = require("express");
const axios = require("axios");

require("dotenv").config();

const PORT = process.env.PORT || 3033;
const OPENAI_API_BASE_URL =
  process.env.OPENAI_API_BASE_URL || "https://api.openai.com";
const OPENAI_ALLOWED_MODELS = process.env.OPENAI_ALLOWED_MODELS || "";

const app = express();
app.use(express.json());

/**
 * Temp middleware to log all requests
 */
app.use((req, res, next) => {
  console.log(`New request > ${req.method} ${req.url}`);
  console.log("    headers > ", req.headers["authorization"]);
  console.log("    body    > ", req.body);
  next();
});

/**
 * Health check endpoint
 */
app.head("/", (req, res) => {
  return res.send();
});

/**
 * Endpoint to get all available models
 */
app.get("/api/tags", async (req, res) => {
  // Example response from Ollama
  // {
  //   "models": [
  //     {
  //       "name": "deepseek-r1:7b",
  //       "model": "deepseek-r1:7b",
  //       "modified_at": "2025-03-23T07:52:42.164862186Z",
  //       "size": 4683075271,
  //       "digest": "0a8c266910232fd3291e71e5ba1e058cc5af9d411192cf88b6d30e92b6e73163",
  //       "details": {
  //         "parent_model": "",
  //         "format": "gguf",
  //         "family": "qwen2",
  //         "families": [
  //           "qwen2"
  //         ],
  //         "parameter_size": "7.6B",
  //         "quantization_level": "Q4_K_M"
  //       }
  //     },
  //     {
  //       "name": "llama3.2:3b",
  //       "model": "llama3.2:3b",
  //       "modified_at": "2025-03-23T07:52:41.339842922Z",
  //       "size": 2019393189,
  //       "digest": "a80c4f17acd55265feec403c7aef86be0c25983ab279d83f3bcd3abbcb5b8b72",
  //       "details": {
  //         "parent_model": "",
  //         "format": "gguf",
  //         "family": "llama",
  //         "families": [
  //           "llama"
  //         ],
  //         "parameter_size": "3.2B",
  //         "quantization_level": "Q4_K_M"
  //       }
  //     }
  //   ]
  // }

  // Example response from OpenAI API
  // {
  //   object: "list",
  //   data: [
  //     {
  //       id: "gpt-4",
  //       object: "model",
  //       created: 1687882411,
  //       owned_by: "openai",
  //     },
  //     {
  //       id: "o1-pro",
  //       object: "model",
  //       created: 1742251791,
  //       owned_by: "system",
  //     },
  //   ],
  // }
  try {
    const response = await axios.get(`${OPENAI_API_BASE_URL}/v1/models`, {
      headers: {
        Authorization: req.headers["authorization"],
      },
    });
    const responseModels = response.data.data;

    const modelFilter = OPENAI_ALLOWED_MODELS
      ? OPENAI_ALLOWED_MODELS.split(",")
      : [];

    const filteredModels = responseModels.filter((model) =>
      modelFilter.length ? modelFilter.includes(model.id) : true
    );

    const formattedModels = filteredModels.map((model) => ({
      name: model.id,
      model: model.id,
      modified_at: new Date(model.created * 1000).toISOString(),
      size: 0,
      digest: "",
      details: {
        parent_model: "",
        format: "",
      },
    }));

    res.send({
      models: formattedModels,
    });
  } catch (error) {
    console.error(error);
    res.status(500).json({ error: error.message });
  }
});

/**
 * Endpoint to chat with the assistant
 */
app.post("/api/chat", async (req, res) => {
  const { model, messages, stream, options } = req.body;

  const response = await axios({
    method: "post",
    url: `${OPENAI_API_BASE_URL}/v1/chat/completions`,
    data: {
      model: model,
      messages: messages,
      stream: stream,
    },
    headers: {
      Authorization: req.headers["authorization"],
    },
    responseType: stream ? "stream" : "json",
  });

  const openStream = response.data;

  // [ollama] EXAMPLE CHUNK:
  // {
  //   model: "deepseek-r1:1.5b",
  //   created_at: "2025-03-22T17:01:31.748659Z",
  //   message: { role: "assistant", content: "\n\n" },
  //   done: false,
  // }

  // [ollama] EXAMPLE FINAL CHUNK:
  // {
  //   model: "deepseek-r1:1.5b",
  //   created_at: "2025-03-22T17:01:37.788115Z",
  //   message: { role: "assistant", content: "" },
  //   done_reason: "stop",
  //   done: true,
  //   total_duration: 12801095381,
  //   load_duration: 4858364600,
  //   prompt_eval_count: 4,
  //   prompt_eval_duration: 1511149623,
  //   eval_count: 41,
  //   eval_duration: 6429075170,
  // }

  // [openai] EXAMPLE CHUNK:
  // data: {
  //   id: "chatcmpl-BDwZHr1gDRCmTsBDHgLXmlB90fRhU",
  //   object: "chat.completion.chunk",
  //   created: 1742663099,
  //   model: "gpt-4o-2024-08-06",
  //   service_tier: "default",
  //   system_fingerprint: "fp_90d33c15d4",
  //   choices: [
  //     {
  //       index: 0,
  //       delta: { content: "." },
  //       logprobs: null,
  //       finish_reason: null,
  //     },
  //   ],
  // }

  // [openai] EXAMPLE FINAL CHUNK:
  // data: {
  //   id: "chatcmpl-BDwZHr1gDRCmTsBDHgLXmlB90fRhU",
  //   object: "chat.completion.chunk",
  //   created: 1742663099,
  //   model: "gpt-4o-2024-08-06",
  //   service_tier: "default",
  //   system_fingerprint: "fp_90d33c15d4",
  //   choices: [{ index: 0, delta: {}, logprobs: null, finish_reason: "stop" }],
  // }
  // data: [DONE]
  openStream.on("data", (data) => {
    const stringData = data.toString();

    const lines = stringData.split("\n").filter((line) => line.trim() !== "");

    for (const line of lines) {
      // String has 6 chars at the beginning: "data: "
      const message = line.substring(6);

      // Sometimes string is "undefined:1"
      if (message.startsWith("undefined:")) {
        continue;
      }

      if (message === "[DONE]") {
        return;
      }

      let text = "";

      try {
        const parsed = JSON.parse(message);

        text = parsed.choices[0].delta.content;
      } catch (error) {
        console.error(error);
        text = "";
      }

      const responseData = {
        model: model,
        message: {
          role: "assistant",
          content: text,
        },
      };

      console.log(text);
      res.write(JSON.stringify(responseData));
      res.write("\n");
    }
  });

  openStream.on("end", () => {
    console.log("stream done");
    res.end();
  });

  openStream.on("error", (error) => {
    console.error(error);
    res.status(500).json({ error: error.message });
  });
});

/**
 * Run the server
 */
app.listen(PORT, () => {
  console.log(`Server is running on http://localhost:${PORT}`);
});
