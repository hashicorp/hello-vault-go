FROM golang:1.16-alpine

COPY . /app

WORKDIR /app

RUN go build -o hello-vault

EXPOSE 8080

CMD ./hello-vault