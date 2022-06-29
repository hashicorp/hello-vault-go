# Two-Vault Setup To Test DNS issues

## Prerequisites

1. [`docker`][docker] to easily run the application in the same environment
   regardless of your local operating system
1. [`docker compose`][docker-compose] to easily set up all the components of the
   demo (the application's web server, the Vault server, the database, etc.) all
   at once
1. [`curl`][curl] to test our endpoints
1. [`jq`][jq] _(optional)_ for prettier `JSON` output

## Try it out

### 1. Bring up the docker-compose environment:

```shell-session
docker compose up -d --build
```

```
[+] Running 11/11
 ⠿ Network sample-app_default                       Created        0.0s
 ⠿ Network sample-app_mynetwork                     Created        0.0s
 ⠿ Volume "sample-app_trusted-orchestrator-volume"  Created        0.0s
 ⠿ Container sample-app-secure-service-1            Healthy       39.3s
 ⠿ Container sample-app-database-1                  Healthy       39.3s
 ⠿ Container sample-app-dns-1                       Healthy       39.3s
 ⠿ Container sample-app-vault-server2-1             Healthy       39.3s
 ⠿ Container sample-app-vault-server1-1             Healthy       39.3s
 ⠿ Container sample-app-trusted-orchestrator-1      Healthy       40.2s
 ⠿ Container sample-app-app-1                       Healthy       42.0s
 ⠿ Container sample-app-app-healthy-1               Started       42.2s

```

The script runs `docker compose up -d --build` and brings up:
1. two vault servers running in `-dev` mode:
  - `vault-server1` (192.168.19.4)
  - `vault-server2` (192.168.19.5)
2. `dns`, unbound DNS server (192.168.19.9)
  - used in place of the default docker compose DNS
  - has `a-records.conf` with entries for all docker-compose containers
  - has `vlt` entry currently pointing to 192.168.19.4 (`vault-server-1`)
3. a test application that uses `vault/api` to authenticate with and send requests to `vlt:8200`
4. a few other containers that are not relevant to this test

### 2. Check the established connections

```shell-session
 $ docker exec -it sample-app-app-1 netstat -n | grep 8200
tcp        0      0 192.168.19.3:52108      192.168.19.4:8200       TIME_WAIT
tcp        0      0 192.168.19.3:52112      192.168.19.4:8200       TIME_WAIT
tcp        0      0 192.168.19.3:52104      192.168.19.4:8200       TIME_WAIT
tcp        0      0 192.168.19.3:52114      192.168.19.4:8200       TIME_WAIT
tcp        0      0 192.168.19.3:52106      192.168.19.4:8200       TIME_WAIT
```

These are the initial auth-related connections established to `vlt` which currently points at `192.168.19.4` (`vault-server1`). They are all in `TIME_WAIT` status (closed, waiting for any late server responses).

After about a minute, all these connections are gone:

```shell-session
$ docker exec -it sample-app-app-1 netstat -n | grep 8200 | wc -l
       0
```

### 3. Make the app send a request to Vault and check the established connections

```shell-session
curl -s -X POST http://localhost:8080/payments | jq
```

```json
{
  "message": "hello world!"
}
```

This makes a request to our `app`, which internally reads a Vault secret at `kv-v2/data/api-key` using `vault/api` library. This establishes another connection:

```shell-session
$  docker exec -it sample-app-app-1 netstat -n | grep 8200
tcp        0      0 192.168.19.3:53834      192.168.19.4:8200       TIME_WAIT
```

If we make additional curl request to our app, we'll see more connections established & closed in `TIME_WAIT` as expected.

To load test it, run the following:

```shell-session
k6 run --vus 100 --duration 60s k6.js
```

### 4. Switch our Unbound DNS to point to the other vault server

Modify [`a-records.conf`](docker-compose-setup/dns/etc/a-records.conf) to
point `vlt` at `192.168.19.5`:

```dns
    # vault-server1
    #local-data: "vlt A 192.168.19.4"

    # vault-server2
    local-data: "vlt A 192.168.19.5"
```

Restart the unbound DNS server:

```shell-session
docker compose restart dns
```

Verify that the routing has been updated on the `app` node:

```shell-session
$ docker exec -it sample-app-app-1 nslookup vlt
Server:         127.0.0.11
Address:        127.0.0.11:53


Name:   vlt
Address: 192.168.19.5
```

### 5. Make the app send another request to Vault and check the connections:

```shell-session
$ curl -s -X POST http://localhost:8080/payments | jq
```

```json
{
  "error": "unable to read secret: error encountered while reading secret at kv-v2/data/api-key: Error making API request.\n\nURL: GET http://vlt:8200/v1/kv-v2/data/api-key\nCode: 500. Errors:\n\n* token mac for token_version:1 hmac:\"\\xa5!qx\\x17\\xcc\\xc9\\xd0\\x0ci\\xe5 6v\\xad/V\\x94\\x17M\\x8f\\x150\\x85sXfi\\xc1\\x1f\\x9e\\xe8\" token:\"\\n\\x1chvs.ibIusIJdnAxhOoa6rVQ863LD\" is incorrect: err %!w(<nil>)"
}
```

We get back an error since the two Vault instances are not running in `DR` replication mode (this is an Enterprise feature). However, the fact that token is rejected tells us that we are hitting the other vault server. Verify:

```shell-session
$  docker exec -it sample-app-app-1 netstat -n | grep 8200
tcp        0      0 192.168.19.3:56208      192.168.19.4:8200       TIME_WAIT
tcp        0      0 192.168.19.3:60596      192.168.19.5:8200       TIME_WAIT   <- new connection
tcp        0      0 192.168.19.3:56200      192.168.19.4:8200       TIME_WAIT
tcp        0      0 192.168.19.3:56190      192.168.19.4:8200       TIME_WAIT
tcp        0      0 192.168.19.3:56212      192.168.19.4:8200       TIME_WAIT
```

After about a minute, all the `192.168.19.4` entries disappear as expected:

```shell-session
$  docker exec -it sample-app-app-1 netstat -n | grep 8200
tcp        0      0 192.168.19.3:60596      192.168.19.5:8200       TIME_WAIT
```

[vault]:                 https://www.vaultproject.io/
[vault-leases]:          https://www.vaultproject.io/docs/concepts/lease
[vault-app-role]:        https://www.vaultproject.io/docs/auth/approle
[vault-token-wrapping]:  https://www.vaultproject.io/docs/concepts/response-wrapping
[vault-kv-v2]:           https://www.vaultproject.io/docs/secrets/kv/kv-v2
[vault-postgresql]:      https://www.vaultproject.io/docs/secrets/databases/postgresql
[docker]:                https://docs.docker.com/get-docker/
[docker-compose]:        https://docs.docker.com/compose/install/
[curl]:                  https://curl.se/
[jq]:                    https://stedolan.github.io/jq/
