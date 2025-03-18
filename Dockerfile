# ================
# Étape 1 : Dev
# ================
FROM golang:1.24 AS development

RUN apt-get update && apt-get install -y net-tools iproute2

ENV GO111MODULE=on
WORKDIR /app

# Copy mod files + download deps
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Install air for live reload
RUN go install github.com/air-verse/air@latest

CMD ["air", "-c", ".air.toml"]

# ================
# Étape 2 : Build
# ================
FROM golang:1.24 AS builder

WORKDIR /app
COPY go.mod ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /go/bin/app -buildvcs=false

# ================
# Étape 3 : Production
# ================
FROM alpine:3.17 AS production
WORKDIR /app
COPY --from=builder /go/bin/app /app/app
COPY .env .env
COPY .env.prod .env.prod
EXPOSE 8080
CMD ["/app/app"]
