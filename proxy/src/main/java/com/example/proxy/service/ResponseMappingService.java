package com.example.proxy.service;

import com.example.proxy.dto.OllamaRequest;
import com.example.proxy.dto.OllamaResponse;
import com.example.proxy.dto.OpenAIResponse;

public interface ResponseMappingService {
    OllamaResponse mapToOllamaResponse(OpenAIResponse openAIResponse, OllamaRequest originalRequest);
}
