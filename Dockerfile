# Use official Node.js runtime as a parent image
FROM node:18-alpine

# Set working directory in the container
WORKDIR /app

# Copy package.json and package-lock.json (if available)
COPY package*.json ./

# Install app dependencies
RUN npm install --omit=dev

# Copy app source code to the working directory
COPY . .

# Expose the port your app runs on (default 3033)
EXPOSE 3033

# Define environment variables with default values if needed
ENV NODE_ENV=production
ENV OPENAI_API_BASE_URL="https://api.openai.com"
ENV OPENAI_ALLOWED_MODELS=""

# Define the command to run your app
CMD ["npm", "start"]