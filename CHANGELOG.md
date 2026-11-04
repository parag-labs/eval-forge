# Changelog

All notable changes to this project are documented here. Format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/); this project uses [semantic versioning](https://semver.org/).

## [Unreleased]

### Changed

- Toolchains: Python 3.15, Java 25 (LTS), Go 1.27, Rust edition 2024 (rust-version 1.99), Node 26 LTS; .NET stays on .NET 10 (LTS) with the SDK pinned in `csharp/global.json`.
- CI: GitHub Actions bumped to current majors and now reads versions from `.python-version`, `csharp/global.json`, `go/go.mod` and `ts/.nvmrc`.
- `ts/package-lock.json` now resolves packages from the public npm registry.
- Libraries: TypeScript 7, Vitest 5, `@types/node` 26 (`tsconfig` now targets ES2024 and lists `node` in `types`); xUnit v3 on Microsoft.Testing.Platform (opted in via `csharp/global.json`) and YamlDotNet 18; JUnit 6, SnakeYAML 2.7 and Surefire 3.6; pytest 9 and ruff 0.17.

### Notes

- `gopkg.in/yaml.v3` (Go) and `serde_yaml` (Rust) are already at their final releases; both are unmaintained upstream but were left in place to keep this upgrade behavior-preserving.

## [0.1.0] - 2026-08-09

First public release. eval-driven CI gate: YAML eval sets, pluggable scorers, and a build that fails on regression.

[0.1.0]: https://github.com/parag-labs/eval-forge/releases/tag/v0.1.0
