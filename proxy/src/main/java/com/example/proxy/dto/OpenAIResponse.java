package com.example.proxy.dto;

import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.util.List;

@Data
@NoArgsConstructor
@AllArgsConstructor
public class OpenAIResponse {
    private String id;
    private String object;
    private long created;
    private String model;
    private List<OpenAIChoice> choices;
}
