# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog][keepachangelog] and this project adheres to [Semantic Versioning][semver].

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
