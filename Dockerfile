FROM golang:latest

# Set the working directory inside the container.
WORKDIR /app

# Copy the local package files to the container's workspace.
COPY . .

# Build the Go app
RUN go build -o myapp ./cmd/main.go

# Run the binary.
CMD ["./myapp"]
