FROM golang:1.23-alpine AS build
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod tidy || true

COPY . .

RUN go mod download

RUN go build -o rent

FROM alpine:3.18
COPY --from=build /app/rent /rent
ENV PORT=8000
EXPOSE 8000
ENTRYPOINT ["/rent"]