## unique-request-counter-app
An API which logs the count of unique requests every minute.

## Setup Instructions

### Basic Solution
**Prerequisites**
- Docker

#### Setup Instructions
1. Clone the repository:

    ```sh
    git clone https://github.com/sumit-kushwah/unique-request-counter-app.git
    cd basic-solution
    ```

2. Build the Docker image:

    ```sh
    docker build -t unique-request-basic .
    ```

3. Run the Docker container:

    ```sh
    docker run -p 8080:8080 unique-request-basic
    ```

4. The API will be available at `http://localhost:8080/api/verve/accept`.

### Extended Solution
**Prerequisites**
- Docker
- Docker Compose

#### Setup Instructions
1. Clone the repository:

    ```sh
    git clone https://github.com/sumit-kushwah/unique-request-counter-app.git
    cd extended-solution
    ```

2. Start the services using Docker Compose:

    ```sh
    docker-compose up --build
    ```

3. The API will be available at `http://localhost:8080/api/verve/accept`.

### Environment Variables
For the extended solution, the following environment variables are used:
- `REDIS_ADDR`: Address of the Redis server (default: `redis-server:6379`)
- `KAFKA_ADDR`: Address of the Kafka broker (default: `kafka-broker:9092`)
- `KAFKA_TOPIC`: Kafka topic name for publishing unique request counts (default: `UNIQUE_REQUEST_COUNT_BY_MINUTE`)