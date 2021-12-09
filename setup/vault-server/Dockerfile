FROM vault:latest

# add jq to make scripting the calls a bit easier
# ref: https://stedolan.github.io/jq/
RUN apk add --no-cache jq

# add our config
COPY dev-config.hcl /vault/config/dev-config.hcl

# add our preformatted policies
COPY dev-policy.hcl /vault/config/dev-policy.hcl
COPY trusted-orchestrator-policy.hcl /vault/config/trusted-orchestrator-policy.hcl

# add our trusted cert
COPY certs/certificate.crt /vault/certificate.crt
COPY certs/private.pem /vault/private.pem

COPY entrypoint.sh /vault/entrypoint.sh

EXPOSE 8200
EXPOSE 8400
CMD /vault/entrypoint.sh
