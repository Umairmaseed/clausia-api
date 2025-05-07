# Use an official Golang runtime as a parent image
FROM golang:1.19-alpine AS build

# Set the working directory to /api
WORKDIR /api

# Copy the current directory contents into the container at /api
COPY . .

RUN go mod download

# Build the Go ccapi
RUN go build -o clausia-api

# Use an official Alpine runtime as a parent image
FROM alpine:latest

ENV PATH="${PATH}:/usr/bin/"

RUN apk update 

RUN apk add --no-cache \
    docker \
    openrc \
    git \ 
    gcc \
    gcompat \
    libc-dev  \
    libc6-compat  \
    libstdc++ && \
    ln -s /lib/libc.so.6 /usr/lib/libresolv.so.2

# Add timezone
RUN apk --no-cache add tzdata
ENV TZ="America/Sao_Paulo"

# Set the working directory to /api
WORKDIR /api

RUN mkdir config
COPY ./config config

# Copy the ccapi binary from the build container to the current directory in the Alpine container
COPY --from=build /api/clausia-api /usr/bin/clausia-api

# Run the ccapi binary
CMD ["clausia-api"]