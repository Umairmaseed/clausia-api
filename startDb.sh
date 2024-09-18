#!/bin/bash

# Define container names
DB_CONTAINER="goprocess-db"
DOCKER_COMPOSE_FILE="docker-compose-db.yml"
INIT_SCRIPT_PATH="./config/db-rs-init.sh"

# Start Docker Compose services
echo "Starting Docker Compose services..."
docker-compose -f $DOCKER_COMPOSE_FILE up -d --build $DB_CONTAINER

# Wait for MongoDB to start
echo "Waiting for MongoDB to start..."
sleep 10 # Adjust this delay if necessary

# Initialize MongoDB replica set
echo "Initializing MongoDB replica set..."
docker exec $DB_CONTAINER /scripts/rs-init.sh

# Check if initialization was successful
if [ $? -eq 0 ]; then
    echo "DB initialized successfully."
    # Start mongo-express service
    echo "Starting mongo-express..."
    docker-compose -f $DOCKER_COMPOSE_FILE up -d mongo-express
else
    echo "Failed to initialize DB with exit code $?."
fi
