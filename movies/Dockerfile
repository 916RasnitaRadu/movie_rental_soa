FROM golang:1.23-alpine AS build
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod tidy || true

COPY . .

RUN go mod download

RUN go build -o movies

FROM alpine:3.18
COPY --from=build /app/movies /movies
ENV PORT=8002
EXPOSE 8002
ENTRYPOINT [ "/movies" ]