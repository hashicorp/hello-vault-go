FROM vault:latest
RUN apk add --no-cache jq curl

COPY dev-policy.hcl /vault/config

COPY entrypoint.sh /vault/entrypoint.sh
COPY generate-secret.sh /vault/generate-secret.sh
RUN chmod +x /vault/entrypoint.sh && chmod +x /vault/generate-secret.sh

EXPOSE 8200
CMD /vault/entrypoint.sh