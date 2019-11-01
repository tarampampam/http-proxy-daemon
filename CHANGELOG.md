# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog][keepachangelog] and this project adheres to [Semantic Versioning][semver].

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
