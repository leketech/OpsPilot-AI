FROM golang:1.26.4-alpine AS build

WORKDIR /app

COPY go.mod go.sum* ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /server ./cmd/server

FROM alpine:3.21

WORKDIR /app

COPY --from=build /server /app/server

EXPOSE 8080

CMD ["/app/server"]
