FROM alpine:latest

# add some basic libraries to make scripting the calls a bit easier
# ref: https://stedolan.github.io/jq/
RUN apk add --no-cache jq curl

COPY entrypoint.sh ./entrypoint.sh

CMD ./entrypoint.sh