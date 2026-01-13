FROM golang:1.23-alpine AS build
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod tidy || true

COPY . .

RUN go mod download

RUN go build -o notification

FROM alpine:3.18
COPY --from=build /app/notification /notification
ENV PORT=8004
EXPOSE 8004
ENTRYPOINT ["/notification"]