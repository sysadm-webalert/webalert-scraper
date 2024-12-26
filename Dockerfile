FROM golang:1.23 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY main.go .
RUN CGO_ENABLED=0 GOOS=linux go build -o webalert-webscraper

FROM alpine:3.21.0

RUN apk add --no-cache \
    ca-certificates \
    chromium \
    freetype \
    harfbuzz \
    nss \
    ttf-freefont

RUN addgroup -S webalertgroup && adduser -S scraper -G webalertgroup
WORKDIR /app
COPY --from=builder /app/webalert-webscraper /app/
RUN chown -R scraper:webalertgroup /app && chmod +x /app/webalert-webscraper
USER scraper

CMD ["/app/webalert-webscraper"]
