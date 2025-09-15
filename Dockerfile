FROM golang:1.25-alpine

WORKDIR /app

RUN apk add --no-cache git

COPY . .

RUN go build -o main .

ENV HOST=0.0.0.0

EXPOSE 8080

CMD ["./main"] 