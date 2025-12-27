FROM golang:1.25.5-alpine3.23 AS builder
WORKDIR /build
RUN apk add --no-cache git ca-certificates
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -trimpath -o server ./cmd/api/main.go

FROM scratch
WORKDIR /app
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /build/server .
ENV GODEBUG=madvdontneed=1
ENV GOMEMLIMIT=64MiB
ENV GOGC=50
EXPOSE 3000
CMD ["./server"]