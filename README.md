# gin-playground

A playground for the Gin web framework

## Setup

This project uses a `config.yaml` file for environment variables. To get started quickly, you
can copy and paste the contents of `config-example.yaml` into a new `config.yaml` file.

## Run the project

To get this project up and running, make sure Go is installed on your machine and then run:

    `go run main.go`

Alternatively, you can run `air` if it is installed on your machine.

Once the application is running, visit the [home page](http://localhost:8080).

Gin supports Sentry out of the box and is used in this project. You can add your Sentry DSN to
the `config.yaml` file, or set `SENTRY_KEY` to an empty string.
