# Go NTP Client

This repository contains a Golang-based NTP v4 client implementing https://datatracker.ietf.org/doc/html/rfc5905#section-7.5

## Build

To build the binary, use `make build-linux-amd64` or any other supported target:
- `build-linux-amd64`
- `build-linux-arm64`
- `build-darwin-amd64`
- `build-darwin-arm64`

This would result in `ntpc` binary being created.

## Usage

Configuration is done via environment variables as follows:
|Name|Default|Notes|
|----|-------|-----|
|`NTPC_REMOTE_HOST`|`"time.google.com"`|Hostname of the remote NTP server|
|`NTPC_REMOTE_PORT`|`123`|Port number of the remote NTP server|
|`NTPC_POLL_INTERVAL`|`5`|Number of seconds between polls to the server|
|`NTPC_SYSTIME_UPDATE_ENABLED`|`false`|Enable system time update(*careful with that one*)|


To display the current clock offset relative to the server and the adjusted time, run:
```sh
$ ntpc connect
```

To continuously update the system time until a desired clock offset is reached, run:
```sh
$ ntpc update-system
```
