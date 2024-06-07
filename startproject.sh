#!/bin/bash

# Start the Docker containers
docker-compose up -d

# Run the script to copy the tls-cert.pem
./copy-tls-cert.sh

#  Start go-sign container 
cd go-sign
docker-compose -f docker-compose.yaml up -d --build