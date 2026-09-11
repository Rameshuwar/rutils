# Utility File Converter Microservice

A simple Go microservice for converting files. Built as a Minimum Viable Product (MVP) to demonstrate core HTTP API capabilities.

## Getting Started

1. Run the server from the repository root:
   ```bash
   go run ./cmd/server/main.go
   ```
   *The server will start on http://localhost:8080 (API) and http://localhost:3000 (UI)*

## How to Test the API

You can use `curl` from a new terminal window to test the endpoints.

### 1. Image Conversion (JPG to PNG)

Create a dummy JPG file (or use a real one):
```bash
# Example: If you have an image named photo.jpg
curl -X POST -F "file=@photo.jpg" http://localhost:8080/convert/image --output result.png
```
This will take the JPG and save the output as `result.png`.

### 2. Document Conversion (CSV to PDF)

Create a simple CSV file:
```bash
echo "Name,Role,City\nAlice,Developer,New York\nBob,Manager,London" > data.csv
```

Send it to the API:
```bash
curl -X POST -F "file=@data.csv" http://localhost:8080/convert/document --output result.pdf
```
This will read the CSV and save the output as `result.pdf`.
