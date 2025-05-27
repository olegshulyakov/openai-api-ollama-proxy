package com.example.proxy.service;

import com.example.proxy.config.ProxyConfigProperties;
import com.example.proxy.dto.*;
import org.springframework.http.HttpHeaders;
import org.springframework.http.MediaType;
import org.springframework.stereotype.Service;
import org.springframework.web.reactive.function.client.WebClient;
import reactor.core.publisher.Mono;

import java.util.Collections;
import java.util.regex.Pattern;

@Service
public class RequestForwardingServiceImpl implements RequestForwardingService {

    private final WebClient webClient;
    private final ProxyConfigProperties configProperties;
    private final Pattern modelPattern;

    public RequestForwardingServiceImpl(WebClient.Builder webClientBuilder, ProxyConfigProperties configProperties) {
        this.configProperties = configProperties;
        this.webClient = webClientBuilder.baseUrl(configProperties.openaiServerUrl()).build();

        String regex = configProperties.modelFilterRegex();
        if (regex != null && !regex.isEmpty()) {
            this.modelPattern = Pattern.compile(regex, Pattern.CASE_INSENSITIVE);
        } else {
            this.modelPattern = Pattern.compile(".*"); // Default to allow all if regex is empty or null
        }
    }

    @Override
    public Mono<OpenAIResponse> forwardRequest(OllamaRequest ollamaRequest, String authorizationToken) {
        String requestedModel = ollamaRequest.getModel();

        if (!this.modelPattern.matcher(requestedModel).matches()) {
            System.out.println("Model '" + requestedModel + "' filtered out by regex: " + this.modelPattern.pattern());
            return Mono.error(new IllegalArgumentException("Requested model '" + requestedModel + "' is not allowed by filter."));
        }

        OpenAIRequest openAIRequest = transformRequest(ollamaRequest);
        String openAiEndpoint = "/v1/chat/completions"; // Common endpoint

        WebClient.RequestBodySpec requestBodySpec = webClient.post()
                .uri(openAiEndpoint)
                .contentType(MediaType.APPLICATION_JSON)
                .bodyValue(openAIRequest);

        if (authorizationToken != null && !authorizationToken.isEmpty()) {
            requestBodySpec.header(HttpHeaders.AUTHORIZATION, authorizationToken); // Assumes "Bearer <token>" is passed in authorizationToken
        }

        return requestBodySpec.retrieve()
                .onStatus(httpStatus -> httpStatus.isError(), clientResponse -> {
                    System.err.println("Error from OpenAI provider: " + clientResponse.statusCode());
                    return clientResponse.bodyToMono(String.class)
                            .flatMap(errorBody -> Mono.error(new RuntimeException("Error from OpenAI provider: " + clientResponse.statusCode() + " - " + errorBody)));
                })
                .bodyToMono(OpenAIResponse.class)
                .doOnError(error -> System.err.println("Error during WebClient call: " + error.getMessage()));
    }

    private OpenAIRequest transformRequest(OllamaRequest ollamaRequest) {
        OpenAIMessage openAIMessage = new OpenAIMessage("user", ollamaRequest.getPrompt());
        return new OpenAIRequest(ollamaRequest.getModel(), Collections.singletonList(openAIMessage));
    }
}
