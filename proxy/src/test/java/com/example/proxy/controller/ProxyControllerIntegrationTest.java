package com.example.proxy.controller;

import com.example.proxy.dto.OllamaRequest;
import com.example.proxy.dto.OllamaResponse;
import com.example.proxy.dto.OpenAIChoice;
import com.example.proxy.dto.OpenAIMessage;
import com.example.proxy.dto.OpenAIRequest;
import com.example.proxy.dto.OpenAIResponse;
import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.github.tomakehurst.wiremock.WireMockServer;
import com.github.tomakehurst.wiremock.client.WireMock;
import org.junit.jupiter.api.AfterAll;
import org.junit.jupiter.api.BeforeAll;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.context.SpringBootTest;
import org.springframework.http.HttpHeaders;
import org.springframework.http.MediaType;
import org.springframework.test.context.DynamicPropertyRegistry;
import org.springframework.test.context.DynamicPropertySource;
import org.springframework.test.web.reactive.server.WebTestClient;

import java.util.List;

import static com.github.tomakehurst.wiremock.client.WireMock.*;
import static com.github.tomakehurst.wiremock.core.WireMockConfiguration.options;
import static org.assertj.core.api.Assertions.assertThat;


@SpringBootTest(webEnvironment = SpringBootTest.WebEnvironment.RANDOM_PORT)
public class ProxyControllerIntegrationTest {

    @Autowired
    private WebTestClient webTestClient;

    @Autowired
    private ObjectMapper objectMapper;

    static WireMockServer wireMockServer;

    @BeforeAll
    static void startWireMock() {
        wireMockServer = new WireMockServer(options().dynamicPort());
        wireMockServer.start();
        WireMock.configureFor("localhost", wireMockServer.port()); // Ensure WireMock client is configured
    }

    @AfterAll
    static void stopWireMock() {
        if (wireMockServer != null) {
            wireMockServer.stop();
        }
    }

    @DynamicPropertySource
    static void registerPgProperties(DynamicPropertyRegistry registry) {
        registry.add("proxy.openaiServerUrl", () -> wireMockServer.baseUrl());
        // Restrictive regex for testing model filtering effectively
        registry.add("proxy.modelFilterRegex", () -> "^allowed-model.*$");
    }

    @BeforeEach
    void resetWireMock() {
        if (wireMockServer != null) {
            wireMockServer.resetAll(); // Reset mappings and request logs before each test
        }
    }

    @Test
    void testChatEndpoint_Success() throws JsonProcessingException {
        // Arrange
        String modelToUse = "allowed-model-gpt4";
        OllamaRequest ollamaRequest = new OllamaRequest(modelToUse, "Tell me a joke about proxies.");
        String authToken = "Bearer test-token-success";

        OpenAIMessage openAIMsg = new OpenAIMessage("assistant", "Why was the proxy server so good at stand-up? Because it always had a great intermediary!");
        OpenAIChoice openAIChoice = new OpenAIChoice(openAIMsg, "stop");
        OpenAIResponse mockOpenAIResponse = new OpenAIResponse("chatcmpl-test", "chat.completion", System.currentTimeMillis(), modelToUse, List.of(openAIChoice));
        String mockOpenAIResponseBody = objectMapper.writeValueAsString(mockOpenAIResponse);

        wireMockServer.stubFor(post(urlEqualTo("/v1/chat/completions"))
                .withHeader(HttpHeaders.AUTHORIZATION, equalTo(authToken))
                .withRequestBody(matchingJsonPath("$.model", equalTo(modelToUse)))
                .willReturn(aResponse()
                        .withStatus(200)
                        .withHeader(HttpHeaders.CONTENT_TYPE, MediaType.APPLICATION_JSON_VALUE)
                        .withBody(mockOpenAIResponseBody)));

        // Act & Assert
        webTestClient.post().uri("/api/v1/chat")
                .header(HttpHeaders.AUTHORIZATION, authToken)
                .contentType(MediaType.APPLICATION_JSON)
                .bodyValue(ollamaRequest)
                .exchange()
                .expectStatus().isOk()
                .expectBody(OllamaResponse.class)
                .value(ollamaResponse -> {
                    assertThat(ollamaResponse.getModel()).isEqualTo(modelToUse);
                    assertThat(ollamaResponse.getResponse()).isEqualTo("Why was the proxy server so good at stand-up? Because it always had a great intermediary!");
                });

        // Verify WireMock received the request
        OpenAIRequest expectedOpenAIRequest = new OpenAIRequest(modelToUse, List.of(new OpenAIMessage("user", "Tell me a joke about proxies.")));
        wireMockServer.verify(1, postRequestedFor(urlEqualTo("/v1/chat/completions"))
                .withHeader(HttpHeaders.AUTHORIZATION, equalTo(authToken))
                .withRequestBody(equalToJson(objectMapper.writeValueAsString(expectedOpenAIRequest))));
    }

    @Test
    void testChatEndpoint_ModelFilteredOut() {
        // Arrange
        OllamaRequest ollamaRequest = new OllamaRequest("filtered-out-model", "This model should be filtered.");
        String authToken = "Bearer test-token-filtered";

        // Act & Assert
        webTestClient.post().uri("/api/v1/chat")
                .header(HttpHeaders.AUTHORIZATION, authToken)
                .contentType(MediaType.APPLICATION_JSON)
                .bodyValue(ollamaRequest)
                .exchange()
                .expectStatus().isBadRequest() // Expecting 400 due to IllegalArgumentException from filter
                .expectBody(OllamaResponse.class)
                .value(ollamaResponse -> {
                    assertThat(ollamaResponse.getModel()).isNull(); // Model might be null or original depending on error handling
                    assertThat(ollamaResponse.getResponse()).contains("Requested model 'filtered-out-model' is not allowed by filter");
                });

        // Verify WireMock did NOT receive any request
        wireMockServer.verify(0, postRequestedFor(urlEqualTo("/v1/chat/completions")));
    }

    @Test
    void testChatEndpoint_OpenAIError() throws JsonProcessingException {
        // Arrange
        String modelToUse = "allowed-model-for-error"; // Must match the regex
        OllamaRequest ollamaRequest = new OllamaRequest(modelToUse, "A prompt that will lead to an error.");
        String authToken = "Bearer test-token-error";

        wireMockServer.stubFor(post(urlEqualTo("/v1/chat/completions"))
                .withHeader(HttpHeaders.AUTHORIZATION, equalTo(authToken))
                .withRequestBody(matchingJsonPath("$.model", equalTo(modelToUse)))
                .willReturn(aResponse()
                        .withStatus(500)
                        .withHeader(HttpHeaders.CONTENT_TYPE, MediaType.APPLICATION_JSON_VALUE)
                        .withBody("{\"error\":{\"message\":\"The server had an error while processing your request.\",\"type\":\"server_error\",\"code\":\"internal_error\"}}")));
        
        // Act & Assert
        webTestClient.post().uri("/api/v1/chat")
                .header(HttpHeaders.AUTHORIZATION, authToken)
                .contentType(MediaType.APPLICATION_JSON)
                .bodyValue(ollamaRequest)
                .exchange()
                .expectStatus().isServiceUnavailable() // Expecting 503 due to WebClientResponseException
                .expectBody(OllamaResponse.class)
                .value(ollamaResponse -> {
                    assertThat(ollamaResponse.getModel()).isNull(); // Or original model based on GlobalExceptionHandler
                    assertThat(ollamaResponse.getResponse()).contains("Upstream service error: 500 INTERNAL_SERVER_ERROR");
                    assertThat(ollamaResponse.getResponse()).contains("The server had an error while processing your request.");
                });

        // Verify WireMock received the request
        OpenAIRequest expectedOpenAIRequest = new OpenAIRequest(modelToUse, List.of(new OpenAIMessage("user", "A prompt that will lead to an error.")));
        wireMockServer.verify(1, postRequestedFor(urlEqualTo("/v1/chat/completions"))
                .withHeader(HttpHeaders.AUTHORIZATION, equalTo(authToken))
                .withRequestBody(equalToJson(objectMapper.writeValueAsString(expectedOpenAIRequest))));
    }
}
