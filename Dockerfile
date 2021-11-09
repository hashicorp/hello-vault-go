FROM golang:1.16-alpine AS GO_BUILD
COPY . /app
WORKDIR /app
RUN go build -o /go/bin/hello-vault

FROM scratch
WORKDIR /app
COPY --from=GO_BUILD /go/bin/hello-vault/ ./
EXPOSE 8080
CMD ./hello-vault