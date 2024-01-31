# Go NTP Client

This repository contains a Golang-based NTP v4 client implementing https://datatracker.ietf.org/doc/html/rfc5905#section-7.5

## Usage

All configuration is done via environment variables as follows:
|Name|Default|Notes|
|----|-------|-----|
|`NTPC_REMOTE_HOST`|`"time.google.com"`|Hostname of the remote NTP server|
|`NTPC_REMOTE_PORT`|`123`|Port number of the remote NTP server|
|`NTPC_POLL_INTERVAL`|`5`|Number of seconds between polls to the server|

To display the current clock offset relative to the server and the adjusted time, run:
```sh
$ go run .
```
