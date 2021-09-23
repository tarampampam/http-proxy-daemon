# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog][keepachangelog] and this project adheres to [Semantic Versioning][semver].

## UNRELEASED

### Changed

- Go updated from `1.16.3` up to `1.17.1`

## v0.3.0

### Changed

- Golang updated from `1.15` up to `1.16.3`
- Module name changed from `http-proxy-daemon` to `github.com/tarampampam/http-proxy-daemon`
- Docker image based on `scratch` (instead `alpine` image)
- Logging using `uber-go/zap` package
- HTTP route `/` now outputs HTML-page with basic info
- HTTP errors now in plain text (instead `json`)
- HTTP route `/metrics` generates metrics in [prometheus](https://github.com/prometheus) format
- `--proxy-request-timeout` flag (`serve` sub-command) now accepts string value (examples: `5s`, `15s30ms`) instead count of seconds

### Added

- Support for `linux/arm64`, `linux/arm/v6` and `linux/arm/v7` platforms for docker image
- Sub-command `healthcheck` (hidden in CLI help) that makes a simple HTTP request (with user-agent `HealthChecker/internal`) to the `http://127.0.0.1:8080/live` endpoint. Port number can be changed using `--port`, `-p` flag or `LISTEN_PORT` environment variable
- Healthcheck in dockerfile
- Global (available for all sub-commands) flags:
  - `--log-json` for logging using JSON format (`stderr`)
  - `--debug` for debug information for logging messages
  - `--verbose` for verbose output
- HTTP panics logging middleware

### Removed

- Binary file packing using `upx`
- HTTP route `/ping`

## v0.2.0

### Changed

- Golang updated from `1.14` up to `1.15`

## v0.1.1

### Fixed

- Server logs now includes `user-agent` (format changed from `gorilla/handlers.LoggingHandler` to `gorilla/handlers.CombinedLoggingHandler`)

## v0.1.0

### Removed

- TSL supports (options `--tsl-cert` and `--tsl-key`)

### Added

- Sub-command `version` (instead `-V` flag)

### Changed

- For server starting must be used sub-command `serve` (`./app serve --port 8080` instead `./app --port 8080`)
- Docker image uses unprivileged user for application starting

## v0.0.3

### Fixed

- Fatal bug with concurrent memory access

### Changed

- Keep alive disabled

## v0.0.2

### Changed

- Labels for docker image

## v0.0.1

### Added

- Basic features:
  - Metrics endpoint
  - Ping endpoint
  - Allowed routes list (if requested `/`)
  - Proxy route
- CLI allows to use options (can read it from environment variables) like listen address, port, and proxy route prefix
- Source code tests
- TSL supports (options `--tsl-cert` and `--tsl-key`)

[keepachangelog]:https://keepachangelog.com/en/1.0.0/
[semver]:https://semver.org/spec/v2.0.0.html
