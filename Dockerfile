FROM golang:1.25.1 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY schema/sqlite.sql schema/sqlite.sql
# mattn/go-sqlite3 needs CGO enabled
RUN CGO_ENABLED=1 GOOS=linux go build -C cmd/go-nland-liveticker -a -installsuffix cgo -o /app/go-nland-liveticker
RUN chmod +x /app/go-nland-liveticker

FROM debian:trixie-slim
WORKDIR /root
RUN apt update && apt install ca-certificates -y
COPY --from=builder /app/go-nland-liveticker /root/go-nland-liveticker
COPY --from=builder /app/schema/sqlite.sql /root/schema/sqlite.sql
CMD ["./go-nland-liveticker"]
