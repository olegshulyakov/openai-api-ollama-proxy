package com.example.proxy.controller.advice;

import com.example.proxy.dto.OllamaResponse;
import org.springframework.http.HttpStatus;
import org.springframework.web.bind.annotation.ExceptionHandler;
import org.springframework.web.bind.annotation.ResponseStatus;
import org.springframework.web.bind.annotation.RestControllerAdvice;
import org.springframework.web.reactive.function.client.WebClientResponseException;
import reactor.core.publisher.Mono;
// import org.slf4j.Logger;
// import org.slf4j.LoggerFactory;

@RestControllerAdvice // Use RestControllerAdvice for reactive applications
public class GlobalExceptionHandler {

    // private static final Logger logger = LoggerFactory.getLogger(GlobalExceptionHandler.class);

    @ExceptionHandler(IllegalArgumentException.class)
    @ResponseStatus(HttpStatus.BAD_REQUEST)
    public Mono<OllamaResponse> handleIllegalArgumentException(IllegalArgumentException ex) {
        // logger.warn("Bad request: {}", ex.getMessage());
        System.err.println("Handled IllegalArgumentException: " + ex.getMessage()); // Simple logging
        return Mono.just(new OllamaResponse(null, "Error: " + ex.getMessage()));
    }

    @ExceptionHandler(WebClientResponseException.class)
    // We can't use @ResponseStatus for dynamic status codes easily here with Mono<OllamaResponse>
    // The status code will be based on what WebClient throws or a default.
    // For more control, the controller method itself would return Mono<ResponseEntity<OllamaResponse>>
    // For now, let's set a generic server error and include details in the body.
    @ResponseStatus(HttpStatus.SERVICE_UNAVAILABLE) // Or INTERNAL_SERVER_ERROR
    public Mono<OllamaResponse> handleWebClientResponseException(WebClientResponseException ex) {
        // logger.error("Error from downstream service: {} {} - {}", ex.getStatusCode(), ex.getResponseBodyAsString(), ex);
         System.err.println("Handled WebClientResponseException: " + ex.getStatusCode() + " - " + ex.getResponseBodyAsString());
        return Mono.just(new OllamaResponse(null, "Upstream service error: " + ex.getStatusCode() + " - " + ex.getResponseBodyAsString()));
    }

    @ExceptionHandler(Exception.class)
    @ResponseStatus(HttpStatus.INTERNAL_SERVER_ERROR)
    public Mono<OllamaResponse> handleGenericException(Exception ex) {
        // logger.error("An unexpected error occurred: {}", ex.getMessage(), ex);
        System.err.println("Handled Generic Exception: " + ex.getMessage());
        return Mono.just(new OllamaResponse(null, "An unexpected internal server error occurred: " + ex.getMessage()));
    }
}
