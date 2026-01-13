# we name it "build" so we can copy from it (this stage) later
FROM golang:1.23-alpine AS build
WORKDIR /app

# copy mod and sum to help docker cache dependencies
COPY go.mod go.sum ./
RUN go mod tidy || true

# copy source code
COPY . .

# ensure dependencies are fetched
RUN go mod download

# build executable
RUN go build -o auth

FROM alpine:3.18

# copy the built in binary from the prev stage into the new image
COPY --from=build /app/auth /auth
ENV PORT=8000
EXPOSE 8000
# docker will run the Go exec when the container starts
ENTRYPOINT [ "/auth" ]