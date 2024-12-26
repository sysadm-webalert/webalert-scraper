FROM golang:1.23 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o webalert-webscraper

FROM alpine:latest

RUN apk add --no-cache \
    chromium \
    nss \
    freetype \
    harfbuzz \
    ca-certificates \
    ttf-freefont

ENV ROD_BROWSER_BIN=/usr/bin/chromium-browser

COPY --from=builder /app/webalert-webscraper .

RUN chmod +x webalert-webscraper

CMD ["./web_check_app"]
