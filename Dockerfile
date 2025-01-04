FROM golang:1.23-alpine AS base

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -ldflags="-s -w" -o main


FROM alpine
COPY --from=base /app/main ./
EXPOSE 8000
CMD ./main