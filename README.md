# onpicket

> External asset threat modeling

## Server setup

To run the server:

```shell
air
# OR
# task dev
```

## Requirements

- MongoDB
- NATS

## Run locally

Two options:

### Docker and hot-reloading locally (recommended)

The server needs an active NATS and MongoDB server. To create these containers
run `task mongo` and `task nats`. Once they are up run `task dev`.

You can now develop with hot-reloading.

### Docker compose (no hot-reload)

Simply run `docker compose up`

## Assets setup

The CSS and JS requires some manual building occasionally.

A `Makefile` helper exists to do both of the following in a single command.
`task assets` will regenerate new bundles.
