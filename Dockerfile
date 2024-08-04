# Build stage
FROM golang:1.22.1-alpine AS build
RUN apk --update add ca-certificates git
WORKDIR /src/module
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/main.go

# Run stage
FROM alpine as run
WORKDIR /root/
COPY --from=build /src/module/server ./server
COPY .env ./.env

# Install ca-certificates in the run stage
RUN apk --no-cache add ca-certificates

ENTRYPOINT ["./server"]
