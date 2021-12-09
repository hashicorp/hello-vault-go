FROM vault:latest

# add jq to make scripting the calls a bit easier
# ref: https://stedolan.github.io/jq/
RUN apk add --no-cache jq

# add our preformatted policies
COPY dev-policy.hcl /vault/config/dev-policy.hcl
COPY trusted-orchestrator-policy.hcl /vault/config/trusted-orchestrator-policy.hcl

COPY entrypoint.sh /vault/entrypoint.sh

EXPOSE 8200
CMD /vault/entrypoint.sh
