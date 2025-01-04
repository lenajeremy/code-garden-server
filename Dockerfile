FROM golang:1.18-bullseye AS base

RUN apt-get update && apt-get install -y --no-install-recommends git \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

RUN go install github.com/cosmtrek/air@latest

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

EXPOSE 8000

CMD ["air"]
