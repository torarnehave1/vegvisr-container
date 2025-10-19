# syntax=docker/dockerfile:1

FROM golang:1.24-alpine AS build

# Set destination for COPY
WORKDIR /app

# Download any Go modules
COPY container_src/go.mod ./
RUN go mod download

# Copy container source code
COPY container_src/*.go ./

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o /server

FROM alpine:latest

# Install FFmpeg and other necessary tools
RUN apk add --no-cache ffmpeg ca-certificates

# Copy the built server
COPY --from=build /server /server

# Create a temporary directory for processing files
RUN mkdir -p /tmp/processing

EXPOSE 8080

# Run
CMD ["/server"]
