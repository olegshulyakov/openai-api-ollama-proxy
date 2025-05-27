package com.example.proxy.service;

import com.example.proxy.dto.*;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;

import java.util.Collections;
import java.util.List;

import static org.junit.jupiter.api.Assertions.*;

class ResponseMappingServiceImplTest {

    private ResponseMappingServiceImpl responseMappingService;

    @BeforeEach
    void setUp() {
        responseMappingService = new ResponseMappingServiceImpl();
    }

    @Test
    void mapToOllamaResponse_Success() {
        OllamaRequest originalRequest = new OllamaRequest("original-model", "prompt");
        OpenAIMessage openAIMessage = new OpenAIMessage("assistant", "Hello there!");
        OpenAIChoice openAIChoice = new OpenAIChoice(openAIMessage, "stop");
        OpenAIResponse openAIResponse = new OpenAIResponse("id1", "chat.completion", System.currentTimeMillis(), "gpt-3.5-turbo", List.of(openAIChoice));

        OllamaResponse result = responseMappingService.mapToOllamaResponse(openAIResponse, originalRequest);

        assertNotNull(result);
        assertEquals("gpt-3.5-turbo", result.getModel());
        assertEquals("Hello there!", result.getResponse());
    }

    @Test
    void mapToOllamaResponse_NoChoices() {
        OllamaRequest originalRequest = new OllamaRequest("original-model", "prompt");
        OpenAIResponse openAIResponse = new OpenAIResponse("id1", "chat.completion", System.currentTimeMillis(), "gpt-3.5-turbo", Collections.emptyList());

        OllamaResponse result = responseMappingService.mapToOllamaResponse(openAIResponse, originalRequest);

        assertNotNull(result);
        assertEquals("gpt-3.5-turbo", result.getModel());
        assertEquals("Error: No content found in OpenAI response", result.getResponse());
    }

    @Test
    void mapToOllamaResponse_NullChoices() {
        OllamaRequest originalRequest = new OllamaRequest("original-model", "prompt");
        OpenAIResponse openAIResponse = new OpenAIResponse("id1", "chat.completion", System.currentTimeMillis(), "gpt-3.5-turbo", null);

        OllamaResponse result = responseMappingService.mapToOllamaResponse(openAIResponse, originalRequest);

        assertNotNull(result);
        assertEquals("gpt-3.5-turbo", result.getModel());
        assertEquals("Error: No content found in OpenAI response", result.getResponse());
    }
    
    @Test
    void mapToOllamaResponse_ChoiceWithMessageNullContent() {
        OllamaRequest originalRequest = new OllamaRequest("original-model", "prompt");
        OpenAIMessage openAIMessage = new OpenAIMessage("assistant", null);
        OpenAIChoice openAIChoice = new OpenAIChoice(openAIMessage, "stop");
        OpenAIResponse openAIResponse = new OpenAIResponse("id1", "chat.completion", System.currentTimeMillis(), "gpt-3.5-turbo", List.of(openAIChoice));

        OllamaResponse result = responseMappingService.mapToOllamaResponse(openAIResponse, originalRequest);

        assertNotNull(result);
        assertEquals("gpt-3.5-turbo", result.getModel());
        assertEquals("Error: No content found in OpenAI response", result.getResponse());
    }

    @Test
    void mapToOllamaResponse_ModelFallback() {
        OllamaRequest originalRequest = new OllamaRequest("fallback-model", "prompt");
        OpenAIMessage openAIMessage = new OpenAIMessage("assistant", "Response content");
        OpenAIChoice openAIChoice = new OpenAIChoice(openAIMessage, "stop");
        OpenAIResponse openAIResponse = new OpenAIResponse("id1", "chat.completion", System.currentTimeMillis(), null, List.of(openAIChoice));

        OllamaResponse result = responseMappingService.mapToOllamaResponse(openAIResponse, originalRequest);

        assertNotNull(result);
        assertEquals("fallback-model", result.getModel());
        assertEquals("Response content", result.getResponse());
    }

    @Test
    void mapToOllamaResponse_NullOpenAIResponse() {
        OllamaRequest originalRequest = new OllamaRequest("original-model", "prompt");
        OllamaResponse result = responseMappingService.mapToOllamaResponse(null, originalRequest);

        assertNotNull(result);
        assertEquals("original-model", result.getModel());
        assertEquals("Error: No response from OpenAI provider", result.getResponse());
    }
}
