# On Picket

> External Attack Surface Management

On Picket is an open source project useful for assessing your external attack
surface.

This project is in pre-alpha - the API could break at any time during this phase. In
this phase the only supported scanner is [nmap].

## Overview

On Picket is primarily an API to interact with several other open source applications used in
threat modelling.

By providing an API it is possible to create your own ways, means and methodologies for scheduling
these tools. No terminal or interactive TTY's are needed, only `curl`. A user interface may be added
in the future.

## Hosted

A hosted version of this project is live at [onpicket.com][url]. It is available for use today
however, until we reach a stable version expect braking changes and possible data loss!

The documentation can be found at [onpicket.com/docs](https://onpicket.com/docs).

To get a glimpse of the previous responses, run this command:

```shell
# tip: pipe to `jq` for better results.
curl --request GET \
  --url 'https://onpicket.com/api/scans?page=1&page_size=1' \
  --header 'Accept: application/json'
```

## Self-Host

This application can be self-hosted easily.

The following services are needed to run On Picket:

- [MongoDB](https://mongodb.com)
- [NATS](https://nats.io)
- [OnPicket][ops] server

The easiest way to get started is to run the `docker-compose.yml` file locally. This has basic
defaults and should not be deployed as is to a production server.

**NOTE**: until the first release, you will need to compile the binary by running `go
build` manually. Likewise for the docker images.

Once the API is stable releases and docker images will be published and more details
provided for a better self-hosted experience.

## Local Development

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

The assets are compiled as needed, meaning when adding new Tailwind classes the assets
will need to be re-compiled.

A shortcut for this is to use [entr](https://github.com/eradman/entr). Here's my snippet, which also
uses [fd](https://github.com/eradman/entr):

```shell
fd . 'assets/templates' | entr -s -c 'task assets'
```
This will re-compile the `js` and `css` files anytime a file within `assets/templates` is
changed. `air` will then hot-reload the application because the files in the directory have changed.

## Contributing

Please contribute by raising issues, opening discussions or reaching out to be directly.

A contribution guide will come in due time.

[nmap]: https://nmap.org
[url]: https://onpicket.com?ref=github.com
[ops]: https://github.com/danielmichaels/onpicket/releases
