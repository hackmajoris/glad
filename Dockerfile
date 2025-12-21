# Multi-stage build for AWS Lambda Go function
# Stage 1: Build the Go application
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the Lambda function
# CGO_ENABLED=0 for static binary
# -ldflags="-s -w" to reduce binary size
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w" \
    -o /build/bootstrap \
    ./cmd/app

# Stage 2: Create the Lambda runtime image
FROM public.ecr.aws/lambda/provided:al2023

# Copy the binary from builder
COPY --from=builder /build/bootstrap ${LAMBDA_RUNTIME_DIR}/bootstrap

# Set the CMD to your handler (could also be done as a parameter override outside)
CMD [ "bootstrap" ]