# Changelog

All notable changes to this project are documented here. Format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/); this project uses [semantic versioning](https://semver.org/).

## [Unreleased]

### Changed

- Toolchains: Python 3.15, Java 25 (LTS), Go 1.27, Rust edition 2024 (rust-version 1.99), Node 26 LTS; .NET stays on .NET 10 (LTS) with the SDK pinned in `csharp/global.json`.
- CI: GitHub Actions bumped to current majors and now reads versions from `.python-version`, `csharp/global.json`, `go/go.mod` and `ts/.nvmrc`.
- `ts/package-lock.json` now resolves packages from the public npm registry.

## [0.1.0] - 2026-08-09

First public release. eval-driven CI gate: YAML eval sets, pluggable scorers, and a build that fails on regression.

[0.1.0]: https://github.com/parag-labs/eval-forge/releases/tag/v0.1.0
