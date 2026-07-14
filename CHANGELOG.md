# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/),
and this project adheres to [Semantic Versioning](https://semver.org/).

## [Unreleased]

### Added

- Cross-platform build tasks: `build:win`, `build:linux`, `build:mac`
- CHANGELOG.md

### Changed

- Renamed project-wide from `shpl` to `spl`: binary, CLI, directory, test fixtures, docs, Docker, WASM
- Version date now falls back to `git log` commit date when `date -u` is unavailable
- README now includes project logo and cross-platform examples

### Removed

- Future-phase items (LSP server, performance profiling, language extensions) removed from PROGRESS.md

## [0.1.0] - 2026-07-14

### Added

- Full SPL interpreter pipeline: Lexer → Parser → Semantic Analyzer → Runtime
- Cobra CLI with 6 subcommands: `run`, `tokens`, `ast`, `repl`, `version`, `about`
- Interactive REPL with phase-gated skeleton injection, auto-declaration, and stdin replay
- WASM playground for browser-based SPL execution
- Docker build with multi-stage CI support
- MkDocs documentation site with Material theme
- GitHub Actions CI/CD (quality gates + docs deployment)
- Comprehensive test suite: lexer (golden snapshots), parser (golden JSON + error fixtures), semantic (16 fixtures, 97.9% statement coverage), runtime (7 golden fixtures), CLI integration
