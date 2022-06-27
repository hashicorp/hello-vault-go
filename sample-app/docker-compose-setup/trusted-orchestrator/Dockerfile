FROM alpine:latest

# Add some basic command-line tools to make scripting the calls a bit easier
# ref: https://stedolan.github.io/jq/
# ref: https://curl.se/
RUN apk add --no-cache jq curl

COPY entrypoint.sh ./entrypoint.sh

ENTRYPOINT [ "./entrypoint.sh" ]

HEALTHCHECK \
    --start-period=1s \
    --interval=1s \
    --timeout=1s \
    --retries=30 \
        CMD [ "/bin/sh", "-c", "[ -f /tmp/secret ]" ]
