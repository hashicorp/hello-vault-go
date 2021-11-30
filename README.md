# hello-vault-go

This is a sample application that demonstrates how to authenticate to and retrieve secrets from HashiCorp [Vault](https://www.vaultproject.io/).

## Prerequisites

1. [Docker](https://docs.docker.com/get-docker/) to easily run the application in the same environment regardless of your local operating system.
2. [Docker-Compose](https://docs.docker.com/compose/install/) to easily set up all the components of the demo (the application's web server, the Vault server, the database, etc.) all at once.

## How To Run

**WARNING** This Vault server is configured to run in "dev" mode, an insecure setting that allows for easy testing.

1. Clone this repo and `cd` into it.
2. Run `/.run.sh`. This will start up all the components of the application using Docker-Compose. The application will be listening on port 8080. 
3. `curl http://localhost:8080/products` to confirm it's working. If you get back a list of fake products, the application and all of its dependencies were built and run successfully.

## API Endpoints

The application has a variety of REST API endpoints that allow you to explore some of Vault's features. 

Try running these sample `curl` commands to see the result (you can pipe to `jq` to make it prettier), then take a peek at the code to see how the app integrates with Vault to accomplish this.

| Sample curl  | Demonstrated feature |
| ------------- | ------------- |
| `curl http://localhost:8080/products` | Uses Vault's database secrets engine to generate dynamic database credentials, which are then used to make calls to the PostgreSQL database. |
| `curl -X PUT 'http://localhost:8080/admin/keys' \
--header 'Content-Type: application/json' \
--data-raw '{"key": "my-secret-key"}'` | Uses Vault's static secrets engine to store an API key so that we can access another service's restricted API endpoint. |
| `curl -X POST 'http://localhost:8080/payments | Makes a request to another service's restricted API endpoint, using the API key that we just added to Vault's static secrets engine. | 