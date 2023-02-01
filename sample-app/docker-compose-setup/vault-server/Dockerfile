# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

FROM vault:1.11.0

# Add jq to make scripting the calls a bit easier
# ref: https://stedolan.github.io/jq/
RUN apk add --no-cache bash jq

# Add our preformatted policies
COPY dev-policy.hcl                   /vault/config/dev-policy.hcl
COPY trusted-orchestrator-policy.hcl  /vault/config/trusted-orchestrator-policy.hcl

COPY entrypoint.sh                    /vault/entrypoint.sh

EXPOSE 8200

ENTRYPOINT [ "/vault/entrypoint.sh" ]

HEALTHCHECK \
    --start-period=5s \
    --interval=1s \
    --timeout=1s \
    --retries=30 \
        CMD [ "/bin/sh", "-c", "[ -f /tmp/healthy ]" ]
