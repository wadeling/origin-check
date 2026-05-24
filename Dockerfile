FROM golang:1.25-alpine AS builder
WORKDIR /src
RUN apk add --no-cache git
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /bin/api ./cmd/api
RUN CGO_ENABLED=0 go build -o /bin/worker ./cmd/worker
RUN CGO_ENABLED=0 go build -o /bin/scheduler ./cmd/scheduler

FROM alpine:3.21 AS api
RUN apk add --no-cache ca-certificates
COPY --from=builder /bin/api /bin/api
EXPOSE 8080
CMD ["/bin/api"]

FROM alpine:3.21 AS worker
RUN apk add --no-cache ca-certificates
COPY --from=builder /bin/worker /bin/worker
CMD ["/bin/worker"]

FROM alpine:3.21 AS scheduler
RUN apk add --no-cache ca-certificates
COPY --from=builder /bin/scheduler /bin/scheduler
CMD ["/bin/scheduler"]
