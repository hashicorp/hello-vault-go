# hello-vault-go

This is a sample application that demonstrates how to authenticate to and retrieve secrets from HashiCorp's [Vault](https://www.vaultproject.io/).

## Prerequisites

1. [Docker](https://docs.docker.com/get-docker/) to easily run the application in an identical environment regardless of your local operating system.
2. [Docker-Compose](https://docs.docker.com/compose/install/) to easily set up all the components of the demo (the application's web server, the Vault server, the database, etc.) all at once.
3. [Vault](https://www.vaultproject.io/downloads) CLI to easily interact with the Vault server once it's up and running.

## How To Run

**WARNING** This Vault server is configured to run in "dev" mode, an insecure setting that allows for easy testing.

Never use dev mode in production! Please see [these guidelines](https://learn.hashicorp.com/tutorials/vault/production-hardening) for making your Vault server production-ready.

1. Clone this repo and cd into it.
2. Run `/.start.sh`. This will setup the database and Vault server, and then build and run the application.

If you want to run these actions as individual steps, you can run `scripts/setup.sh`, `scripts/build.sh`, and `scripts/run.sh` respectively.

3. The application is now listening on port 8080. In another terminal tab, `curl http://localhost:8080/` to confirm that it's up and running. You should see "Hello Vault! :)"

## API Endpoints 
... [[explanation of each API endpoint and what it's demonstrating goes here]]