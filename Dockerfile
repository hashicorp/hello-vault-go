FROM golang:1.16-alpine
COPY . /app
WORKDIR /app
RUN go build -o hello-vault

FROM alpine:latest
COPY --from=0 /app/hello-vault .
COPY setup/wait-for-vault.sh .
RUN chmod +x wait-for-vault.sh
EXPOSE 8080
