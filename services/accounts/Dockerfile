# Build stage
FROM golang:1.25-alpine AS builder
WORKDIR /src
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/accounts ./cmd/accounts

# Runtime stage
FROM gcr.io/distroless/static:nonroot
WORKDIR /app
COPY --from=builder /out/accounts /app/accounts
ENTRYPOINT ["/app/accounts"]
