package com.example.proxy.controller;

import com.example.proxy.dto.OllamaRequest;
import com.example.proxy.dto.OllamaResponse;
import com.example.proxy.service.RequestForwardingService;
import com.example.proxy.service.ResponseMappingService;
import org.springframework.web.bind.annotation.*;
import reactor.core.publisher.Mono;

@RestController
@RequestMapping("/api/v1")
public class ProxyController {

    private final RequestForwardingService requestForwardingService;
    private final ResponseMappingService responseMappingService;

    public ProxyController(RequestForwardingService requestForwardingService, ResponseMappingService responseMappingService) {
        this.requestForwardingService = requestForwardingService;
        this.responseMappingService = responseMappingService;
    }

    @PostMapping("/chat")
    public Mono<OllamaResponse> handleChat(@RequestBody OllamaRequest ollamaRequest, @RequestHeader(name = "Authorization", required = false) String authorizationHeader) {
        System.out.println("Received model: " + ollamaRequest.getModel());
        System.out.println("Received prompt: " + ollamaRequest.getPrompt());
        if (authorizationHeader != null) {
            System.out.println("Forwarding with Authorization Header: " + authorizationHeader.substring(0, Math.min(authorizationHeader.length(), 15)) + "..."); // Log only a prefix
        } else {
            System.out.println("No Authorization Header received. Forwarding without it.");
        }

        return requestForwardingService.forwardRequest(ollamaRequest, authorizationHeader)
            .map(openAIResp -> responseMappingService.mapToOllamaResponse(openAIResp, ollamaRequest))
            .defaultIfEmpty(new OllamaResponse(ollamaRequest.getModel(), "Error or empty response from OpenAI provider"));
    }
}
