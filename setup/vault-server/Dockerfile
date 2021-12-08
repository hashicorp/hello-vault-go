FROM vault:latest
RUN apk add --no-cache jq curl

COPY dev-policy.hcl /vault/config/

COPY entrypoint.sh /vault/entrypoint.sh
RUN chmod +x /vault/entrypoint.sh

EXPOSE 8200
CMD /vault/entrypoint.sh
