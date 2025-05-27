package com.example.proxy.config;

import jakarta.validation.constraints.NotBlank;
import org.springframework.boot.context.properties.ConfigurationProperties;
import org.springframework.validation.annotation.Validated;

@ConfigurationProperties(prefix = "proxy")
@Validated
public record ProxyConfigProperties(
        @NotBlank String openaiServerUrl,
        String modelFilterRegex
) {
}
