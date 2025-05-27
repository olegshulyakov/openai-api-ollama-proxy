# Stage 1: Build the application
FROM maven:3.8.5-openjdk-17 AS builder

# Set the working directory
WORKDIR /app

# Copy the pom.xml file
COPY pom.xml .

# Download dependencies (this layer will be cached)
RUN mvn dependency:go-offline

# Copy the source code
COPY src ./src

# Package the application
RUN mvn package -DskipTests

# Stage 2: Run the application
FROM openjdk:17-jre-slim

# Set the working directory
WORKDIR /app

# Copy the JAR from the builder stage
COPY --from=builder /app/target/*.jar app.jar

# Expose the port the application runs on
EXPOSE 8080

# Environment variables that can be overridden at runtime
# These are for documentation and provide defaults if not set,
# but Spring Boot will pick up PROXY_OPENAI_SERVER_URL and PROXY_MODEL_FILTER_REGEX directly.
ENV PROXY_OPENAI_SERVER_URL="https://api.openai.com/v1"
ENV PROXY_MODEL_FILTER_REGEX=".*"

# Command to run the application
ENTRYPOINT ["java", "-jar", "app.jar"]
