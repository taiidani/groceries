FROM golang:1.27-alpine AS builder

ENV CGO_ENABLED=0 GOOS=linux

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY --parents **/*.go internal ./

RUN go build -o /groceries main.go

# Deployable image
FROM alpine:latest

RUN adduser -D appuser
USER appuser

COPY --from=builder /groceries /groceries

EXPOSE 3000
ENTRYPOINT ["/groceries"]
