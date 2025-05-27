package com.example.proxy.service;

import com.example.proxy.dto.OllamaRequest;
import com.example.proxy.dto.OllamaResponse;
import com.example.proxy.dto.OpenAIChoice;
import com.example.proxy.dto.OpenAIMessage;
import com.example.proxy.dto.OpenAIResponse;
import org.springframework.stereotype.Service;

import java.util.List;

@Service
public class ResponseMappingServiceImpl implements ResponseMappingService {

    @Override
    public OllamaResponse mapToOllamaResponse(OpenAIResponse openAIResponse, OllamaRequest originalRequest) {
        if (openAIResponse == null) {
            return new OllamaResponse(originalRequest.getModel(), "Error: No response from OpenAI provider");
        }

        String model = openAIResponse.getModel();
        if (model == null || model.isEmpty()) {
            model = originalRequest.getModel();
        }

        String responseContent = "Error: No content found in OpenAI response";
        List<OpenAIChoice> choices = openAIResponse.getChoices();
        if (choices != null && !choices.isEmpty()) {
            OpenAIChoice firstChoice = choices.get(0);
            if (firstChoice != null) {
                OpenAIMessage message = firstChoice.getMessage();
                if (message != null && message.getContent() != null && !message.getContent().isEmpty()) {
                    responseContent = message.getContent();
                }
            }
        }
        return new OllamaResponse(model, responseContent);
    }
}
