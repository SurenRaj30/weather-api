# Weather API
- Project URL: https://roadmap.sh/projects/weather-api-wrapper-service

## Overview
This project is a simple weather API built using Go, designed to retrieve weather data and integrate with Redis for caching. The application is configured to run on port 9000 and is easily scalable with Docker.

## Features
- Retrieves weather data from external APIs
- Caches responses using Redis
- Simple Go server setup

## Requirements
- Go 1.16+
- Redis
- Docker (optional, for containerization)

## Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/SurenRaj30/weather-api.git
   cd weather-api
   ```

2. Install Go dependencies:
   ```bash
   go mod download
   ```

3. Configure Redis (ensure Redis is running).

4. Run the application:
   ```bash
   go run main.go
   ```

## Running with Docker

1. Build the Docker image:
   ```bash
   docker-compose build
   ```

2. Run the container:
   ```bash
   docker-compose up
   ```

## Usage

Send a GET request to the following endpoint:
```
http://localhost:9000/weather?city=<city_name>
```

## Configuration

- Ensure Redis is running locally or update the Redis host settings in `main.go`.

## License
This project is licensed under the MIT License.
