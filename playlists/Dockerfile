FROM golang:1.23.3-alpine3.20 AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN go build -o ./bin/run_app ./cmd/main/main.go

FROM alpine:3.20 AS runner
WORKDIR /app

COPY --from=builder /app/bin ./
RUN adduser -DH usr && chown -R usr: /app && chmod -R 700 /app

USER usr
 
CMD [ "./run_app" ]