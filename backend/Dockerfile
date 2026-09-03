FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o api ./cmd/api
RUN CGO_ENABLED=0 GOOS=linux go build -o backfill-images ./cmd/backfill-images
RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app
COPY --from=builder /app/api .
COPY --from=builder /app/backfill-images .
COPY --from=builder /app/migrations ./migrations
COPY --from=builder /go/bin/migrate .
EXPOSE 8080
CMD ["./api"]
