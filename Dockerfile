FROM golang:1.16-alpine

RUN apk --no-cache add curl unzip jq

RUN curl https://releases.hashicorp.com/vault/1.8.4/vault_1.8.4_linux_amd64.zip -o vault.zip && \
    unzip vault.zip && \
    mv vault /usr/bin

COPY . /app

WORKDIR /app

RUN go build -o bin/hello-vault

EXPOSE 8080

CMD /app/setup/entrypoint.sh