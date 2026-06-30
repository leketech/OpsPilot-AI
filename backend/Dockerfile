# --- Build Stage ---
FROM golang:1.26.4-alpine AS build

WORKDIR /app

# 1. Copy dependency files first (Docker caches this layer, speeding up future builds)
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# 2. Copy the rest of the source code
COPY . .

# 3. Build the binary
# CGO_ENABLED=0 creates a static binary
# -ldflags="-w -s" strips debug information, making the binary significantly smaller
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /server ./cmd/server

# --- Final Stage ---
FROM alpine:3.21

# 4. CRITICAL: Install CA certificates and timezone data
# Without ca-certificates, HTTPS calls to AI APIs will fail!
# tzdata ensures your logrus timestamps are accurate
RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

# Copy the binary from the build stage
COPY --from=build /server /app/server

EXPOSE 8080

CMD ["/app/server"]