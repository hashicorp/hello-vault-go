FROM alpine:latest

# We'll use some basic libraries to make scripting the calls a bit easier
# ref: https://stedolan.github.io/jq/
RUN apk add --no-cache jq curl

# Add our trusted cert
COPY . ./

CMD ./entrypoint.sh