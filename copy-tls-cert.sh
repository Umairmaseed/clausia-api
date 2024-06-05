#!/bin/bash

# Remove the .cryptopath folder if it exists
echo "Removing .cryptopath folder..."
if [ -d ".cryptopath" ]; then
  rm -rf .cryptopath
  echo ".cryptopath folder removed."
else
  echo ".cryptopath folder does not exist."
fi

# Remove the tls-cert.pem file from the config directory if it exists
echo "Removing tls-cert.pem from config directory..."
if [ -f "./config/tlscert.pem" ]; then
  rm ./config/tlscert.pem
  echo "tls-cert.pem removed from config directory."
else
  echo "tls-cert.pem does not exist in the config directory."
fi

# Wait for the ca.goprocess.org container to be up and running
echo "Waiting for ca.goprocess.org container to be up..."
while ! docker ps | grep -q "ca.goprocess.org"; do
  sleep 1
done

# Wait for the fabric-ca server to generate the tls-cert.pem
echo "Waiting for tls-cert.pem to be generated..."
while ! docker exec ca.goprocess.org [ -f /etc/hyperledger/fabric-ca-server/tls-cert.pem ]; do
  sleep 1
done

# Copy the tls-cert.pem from the ca.goprocess.org container to the local config directory
echo "Copying tls-cert.pem to the local config directory..."
docker cp ca.goprocess.org:/etc/hyperledger/fabric-ca-server/tls-cert.pem ./config/tlscert.pem

echo "tls-cert.pem has been copied successfully."
