# Stage 1: Build the Vite frontend
FROM node:22-alpine AS frontend-builder
WORKDIR /app/frontend
COPY frontend/package*.json ./
RUN npm install
COPY frontend/ ./
RUN npm run build

# Stage 2: Build the Go backend
FROM golang:alpine AS backend-builder
# Use latest alpine to support the latest go versions
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server/main.go

# Stage 3: Create the final minimal image
FROM alpine:latest
WORKDIR /app

# Copy backend binary
COPY --from=backend-builder /app/server .

# Copy frontend dist folder to the expected location
COPY --from=frontend-builder /app/frontend/dist ./frontend/dist

# Copy docs if required by backend swagger
COPY --from=backend-builder /app/docs ./docs

# Expose the application port
EXPOSE 8080

# Command to run the application
CMD ["./server"]
