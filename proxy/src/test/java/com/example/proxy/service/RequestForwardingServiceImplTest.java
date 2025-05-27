package com.example.proxy.service;

import com.example.proxy.config.ProxyConfigProperties;
import com.example.proxy.dto.OllamaRequest;
import com.example.proxy.dto.OpenAIResponse;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.Answers;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;
import org.springframework.http.HttpHeaders;
import org.springframework.http.MediaType;
import org.springframework.web.reactive.function.client.WebClient;
import org.springframework.web.reactive.function.client.WebClientResponseException;
import reactor.core.publisher.Mono;
import reactor.test.StepVerifier;

import static org.mockito.ArgumentMatchers.any;
import static org.mockito.ArgumentMatchers.anyString;
import static org.mockito.Mockito.*;

@ExtendWith(MockitoExtension.class)
class RequestForwardingServiceImplTest {

    @Mock
    private WebClient.Builder webClientBuilder;

    @Mock(answer = Answers.RETURNS_DEEP_STUBS) // For fluent API: webClient.post().uri()...
    private WebClient webClient;

    @Mock
    private WebClient.RequestBodyUriSpec requestBodyUriSpec;
    @Mock
    private WebClient.RequestBodySpec requestBodySpec;
    @Mock
    private WebClient.ResponseSpec responseSpec;

    @Mock
    private ProxyConfigProperties proxyConfigProperties;

    private RequestForwardingServiceImpl requestForwardingService;

    @BeforeEach
    void setUp() {
        // Common WebClient mock setup
        when(webClientBuilder.baseUrl(anyString())).thenReturn(webClientBuilder);
        when(webClientBuilder.build()).thenReturn(webClient);
        when(webClient.post()).thenReturn(requestBodyUriSpec);
        when(requestBodyUriSpec.uri(anyString())).thenReturn(requestBodySpec);
        when(requestBodySpec.contentType(any(MediaType.class))).thenReturn(requestBodySpec);
        when(requestBodySpec.bodyValue(any())).thenReturn(requestBodySpec);
        when(requestBodySpec.retrieve()).thenReturn(responseSpec);

        // Default config for most tests
        when(proxyConfigProperties.openaiServerUrl()).thenReturn("http://localhost:8080");
    }

    private void initializeServiceWithRegex(String regex) {
        when(proxyConfigProperties.modelFilterRegex()).thenReturn(regex);
        requestForwardingService = new RequestForwardingServiceImpl(webClientBuilder, proxyConfigProperties);
    }

    @Test
    void forwardRequest_ModelAllowed_WithToken() {
        initializeServiceWithRegex("allowed-model|another-model");
        OpenAIResponse mockResponse = new OpenAIResponse(); // Customize as needed
        when(responseSpec.bodyToMono(OpenAIResponse.class)).thenReturn(Mono.just(mockResponse));
        // Crucially, for the header check, we need to return the mock itself
        when(requestBodySpec.header(eq(HttpHeaders.AUTHORIZATION), anyString())).thenReturn(requestBodySpec);


        OllamaRequest ollamaRequest = new OllamaRequest("allowed-model", "prompt");
        String token = "Bearer test-token";

        StepVerifier.create(requestForwardingService.forwardRequest(ollamaRequest, token))
                .expectNext(mockResponse)
                .verifyComplete();

        verify(webClientBuilder).baseUrl("http://localhost:8080");
        verify(webClient).post();
        verify(requestBodyUriSpec).uri("/v1/chat/completions");
        verify(requestBodySpec).contentType(MediaType.APPLICATION_JSON);
        verify(requestBodySpec).header(HttpHeaders.AUTHORIZATION, token);
        verify(requestBodySpec).bodyValue(any()); // Could be more specific here
        verify(responseSpec).bodyToMono(OpenAIResponse.class);
    }

    @Test
    void forwardRequest_ModelAllowed_NoToken() {
        initializeServiceWithRegex(".*"); // Allow all
        OpenAIResponse mockResponse = new OpenAIResponse();
        when(responseSpec.bodyToMono(OpenAIResponse.class)).thenReturn(Mono.just(mockResponse));

        OllamaRequest ollamaRequest = new OllamaRequest("any-model", "prompt");

        StepVerifier.create(requestForwardingService.forwardRequest(ollamaRequest, null))
                .expectNext(mockResponse)
                .verifyComplete();

        verify(webClientBuilder).baseUrl("http://localhost:8080");
        verify(webClient).post();
        verify(requestBodyUriSpec).uri("/v1/chat/completions");
        verify(requestBodySpec).contentType(MediaType.APPLICATION_JSON);
        verify(requestBodySpec, never()).header(eq(HttpHeaders.AUTHORIZATION), anyString());
        verify(requestBodySpec).bodyValue(any());
        verify(responseSpec).bodyToMono(OpenAIResponse.class);
    }

    @Test
    void forwardRequest_ModelFilteredOut() {
        initializeServiceWithRegex("^only-this-model$"); // Strict regex

        OllamaRequest ollamaRequest = new OllamaRequest("disallowed-model", "prompt");

        StepVerifier.create(requestForwardingService.forwardRequest(ollamaRequest, "token"))
                .expectError(IllegalArgumentException.class)
                .verify();

        verify(webClient, never()).post(); // Crucial: WebClient should not be interacted with
    }
    
    @Test
    void forwardRequest_ModelAllowed_EmptyRegexMeansAllowAll() {
        initializeServiceWithRegex(""); // Empty regex should default to ".*"
        OpenAIResponse mockResponse = new OpenAIResponse();
        when(responseSpec.bodyToMono(OpenAIResponse.class)).thenReturn(Mono.just(mockResponse));

        OllamaRequest ollamaRequest = new OllamaRequest("any-model-allowed", "prompt");

        StepVerifier.create(requestForwardingService.forwardRequest(ollamaRequest, null))
                .expectNext(mockResponse)
                .verifyComplete();
        
        verify(webClient).post(); // Ensure webClient was called
    }


    @Test
    void forwardRequest_WebClientError() {
        initializeServiceWithRegex("allow-all.*");
        WebClientResponseException webClientException = WebClientResponseException.create(500, "Internal Server Error", null, null, null);
        when(responseSpec.bodyToMono(OpenAIResponse.class)).thenReturn(Mono.error(webClientException));
        // Need to mock the onStatus error handling part if it's complex
        // For simple cases like this, just returning the error from bodyToMono is often enough.
        // If onStatus logic is critical to test, you'd need a more involved mock for responseSpec.onStatus(...)
        when(responseSpec.onStatus(any(), any())).thenReturn(responseSpec);


        OllamaRequest ollamaRequest = new OllamaRequest("allow-all-model", "prompt");

        StepVerifier.create(requestForwardingService.forwardRequest(ollamaRequest, "token"))
                .expectErrorMatches(throwable -> throwable instanceof RuntimeException && throwable.getCause() instanceof WebClientResponseException)
                .verify();

        verify(webClient).post();
        verify(responseSpec).bodyToMono(OpenAIResponse.class);
    }
}
