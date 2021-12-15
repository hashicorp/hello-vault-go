# hello-vault-go

This is a sample application that demonstrates how to authenticate to and
retrieve secrets from HashiCorp [Vault][vault].

## Prerequisites

1. [`docker`][docker] to easily run the application in the same environment
   regardless of your local operating system
1. [`docker compose`][docker-compose] to easily set up all the components of the
   demo (the application's web server, the Vault server, the database, etc.) all
   at once
1. [`curl`][curl] to test our endpoints
1. [`jq`][jq] _(optional)_ for prettier `JSON` output

## Try it out

> **WARNING**: The Vault server used in this setup is configured to run in
> `-dev` mode, an insecure setting that allows for easy testing.

### 1. Bring up the services

This step may take a few minutes to download the necessary dependencies.

```bash
./run.sh
```

```
[+] Running 7/7
 ⠿ Network hello-vault-go_default                          Created        0.1s
 ⠿ Volume "hello-vault-go_trusted-orchestrator-volume"     Created        0.0s
 ⠿ Container hello-vault-go-secure-service-1               Started        0.6s
 ⠿ Container hello-vault-go-database-1                     Started        0.6s
 ⠿ Container hello-vault-go-vault-server-1                 Started        1.3s
 ⠿ Container hello-vault-go-trusted-orchestrator-1         Started        8.6s
 ⠿ Container hello-vault-go-app-1                          Started       10.3s

```

Verify that the services started successfully:

```bash
docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"
```

```
NAMES                                   STATUS                        PORTS
hello-vault-go-app-1                    Up About a minute (healthy)   0.0.0.0:8080->8080/tcp
hello-vault-go-trusted-orchestrator-1   Up About a minute (healthy)
hello-vault-go-vault-server-1           Up About a minute (healthy)   0.0.0.0:8200->8200/tcp
hello-vault-go-secure-service-1         Up About a minute (healthy)   0.0.0.0:1717->80/tcp
hello-vault-go-database-1               Up About a minute (healthy)   0.0.0.0:5432->5432/tcp
```

### 2. Try out `POST /payments` endpoint (static secrets workflow)

`POST /payments` endpoint is a simple example of the static secrets workflow.
Our service will make a request to another service's restricted API endpoint
using an API key value stored in Vault's static secrets engine.

```bash
curl -s -X POST http://localhost:8080/payments | jq
```

```json
{
  "message": "hello world!"
}
```

Check the logs:

```bash
docker logs hello-vault-go-app-1
```

```log
...
2021/12/15 23:15:36 getting secret api key from vault
2021/12/15 23:15:36 getting secret api key from vault: success!
[GIN] 2021/12/15 23:15:36 | 200 |    3.219167ms |    192.168.96.1 | POST     "/payments"
```

### 3. Try out `GET /products` endpoint (dynamic secrets workflow)

`GET /products` endpoint is a simple example of the dynamic secrets workflow.
Our application uses Vault's database secrets engine to generate dynamic
database credentials, which are then used to connect to and retrieve data from a
PostgreSQL database.

```bash
curl -s -X GET http://localhost:8080/products | jq
```

```json
[
  {
    "id": 1,
    "name": "Rustic Webcam"
  },
  {
    "id": 2,
    "name": "Haunted Coloring Book"
  }
]
```

Check the logs:

```bash
docker logs hello-vault-go-app-1
```

```log
2021/12/15 23:17:49 getting temporary database credentials from vault
2021/12/15 23:17:49 getting temporary database credentials from vault: success!
2021/12/15 23:17:49 connecting to "postgres" database @ database:5432 with username "v-approle-dev-read-dBbQdpLrIv8Xyh8nwzSX-1639610269"
2021/12/15 23:17:49 connecting to "postgres" database: success!
...
[GIN] 2021/12/15 - 23:18:22 | 200 |    2.559083ms |    192.168.96.1 | GET      "/products"
```

### 4. Examine the logs for renew logic

One of the complexities of dealing with short-lived secrets is that they must be
renewed periodically. This includes authentication tokens and database
credentials.

Examine the logs for how the Vault auth token is periodically renewed:

```bash
docker logs hello-vault-go-app-1 2>&1 | grep auth
```

```log
2021/12/15 23:17:49 logging in to vault with approle auth; role id: demo-web-app
2021/12/15 23:17:49 logging in to vault with approle auth: success!
2021/12/15 23:17:49 auth token renew / login loop: begin
2021/12/15 23:17:49 auth token renew cycle: started; lease duration: 0s
2021/12/15 23:17:49 auth token: successfully renewed; remaining lease duration: 0s
2021/12/15 23:19:15 auth token: successfully renewed; remaining lease duration: 0s
2021/12/15 23:20:40 auth token: successfully renewed; remaining lease duration: 0s
2021/12/15 23:22:06 auth token: successfully renewed; remaining lease duration: 0s
2021/12/15 23:23:21 auth token: successfully renewed; remaining lease duration: 0s
2021/12/15 23:23:21 auth token renew cycle: the secret can no longer be renewed
2021/12/15 23:23:21 logging in to vault with approle auth; role id: demo-web-app
2021/12/15 23:23:21 logging in to vault with approle auth: success!
2021/12/15 23:23:21 auth token renew cycle: started; lease duration: 0s
2021/12/15 23:23:21 auth token: successfully renewed; remaining lease duration: 0s
2021/12/15 23:24:46 auth token: successfully renewed; remaining lease duration: 0s
```

Examine the logs for database credentials renew / reconnect cycle:

```bash
docker logs hello-vault-go-app-1 2>&1 | grep database
```

```log
2021/12/15 23:17:49 getting temporary database credentials from vault
2021/12/15 23:17:49 getting temporary database credentials from vault: success!
2021/12/15 23:17:49 connecting to "postgres" database @ database:5432 with username "v-approle-dev-read-dBbQdpLrIv8Xyh8nwzSX-1639610269"
2021/12/15 23:17:49 connecting to "postgres" database: success!
2021/12/15 23:17:49 database credentials renew / reconnect loop: begin
2021/12/15 23:17:49 database credentials renew cycle: started; lease duration: 180s
2021/12/15 23:17:49 database credentials: successfully renewed; remaining lease duration: 180s
2021/12/15 23:19:57 database credentials: successfully renewed; remaining lease duration: 180s
2021/12/15 23:22:06 database credentials: successfully renewed; remaining lease duration: 163s
2021/12/15 23:24:14 database credentials renew cycle: the secret can no longer be renewed
2021/12/15 23:24:14 getting temporary database credentials from vault
2021/12/15 23:24:14 getting temporary database credentials from vault: success!
2021/12/15 23:24:14 connecting to "postgres" database @ database:5432 with username "v-approle-dev-read-TG9mg6avrhBO09f1HjEd-1639610654"
2021/12/15 23:24:14 connecting to "postgres" database: success!
2021/12/15 23:24:14 database credentials renew cycle: started; lease duration: 180s
2021/12/15 23:24:14 database credentials: successfully renewed; remaining lease duration: 180s
2021/12/15 23:26:21 database credentials: successfully renewed; remaining lease duration: 180s
2021/12/15 23:28:27 database credentials: successfully renewed; remaining lease duration: 167s
2021/12/15 23:30:37 database credentials renew cycle: the secret can no longer be renewed
```

## Stack Design

### API

| Endpoint             | Description                                                            |
| -------------------- | ---------------------------------------------------------------------- |
| **POST** `/payments` | A simple example of Vault static secrets workflow (see example above)  |
| **GET** `/products`  | A simple example of Vault dynamic secrets workflow (see example above) |

### Docker Compose Architecture

![arch overview](images/arch-overview.svg)

[vault]:           https://www.vaultproject.io/
[docker]:          https://docs.docker.com/get-docker/
[docker-compose]:  https://docs.docker.com/compose/install/
[curl]:            https://curl.se/
[jq]:              https://stedolan.github.io/jq/
