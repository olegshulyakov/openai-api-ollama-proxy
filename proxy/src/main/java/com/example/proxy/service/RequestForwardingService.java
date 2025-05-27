package com.example.proxy.service;

import com.example.proxy.dto.OllamaRequest;
import com.example.proxy.dto.OpenAIResponse;
import reactor.core.publisher.Mono;

public interface RequestForwardingService {
    Mono<OpenAIResponse> forwardRequest(OllamaRequest ollamaRequest, String authorizationToken);
}
