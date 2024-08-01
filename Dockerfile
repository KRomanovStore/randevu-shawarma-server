# Use the official Golang image as the build environment
FROM golang:1.19-alpine AS build

# Set the Current Working Directory inside the container
WORKDIR /app

# Install git as it is needed for downloading Go modules
RUN apk update && apk add --no-cache git

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download all dependencies
RUN go mod download

# Copy the source code to the container
COPY . .

# Build the Go app
RUN go build -o main .

# Use a minimal image as the run environment
FROM alpine:latest

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy the pre-built binary file from the build environment
COPY --from=build /app/main .

# Expose the port on which the app will run
EXPOSE 8090

# Command to run the executable
CMD ["./main"]