# hello-vault-dotnet

This is a sample application that demonstrates how to authenticate to and retrieve secrets from HashiCorp's [Vault](https://www.vaultproject.io/).

## Prerequisites

1. [Docker](https://docs.docker.com/get-docker/) to easily run the application in the same environment regardless of your local operating system.
2. [Docker-Compose](https://docs.docker.com/compose/install/) to easily set up all the components of the demo (the application's web server, the Vault server, the database, etc.) all at once.
3. [Vault]() CLI to easily interact with the Vault server once it's up and running.

## How To Run

To run this application, clone this repo, `cd` into it, then run `/.start.sh`. This will start up all the components of the application using Docker-Compose.

The application will be listening on port 8080. Confirm that it's working correctly by doing `curl http://localhost:8080/`. If you get back "Hello Vault! :)", the application was built and run successfully.

## API Endpoints

The application has a variety of REST API endpoints that allow you to explore some of Vault's features.

(TODO: table explaining each API endpoint and what it's demonstrating here)