FROM golang:latest as builder

WORKDIR /app

COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy from the current directory to the WORKDIR in the container
COPY . .

# Build the server
RUN go build -o publisher publisher-server.go

# Command to run the executable
CMD ["./publisher"]