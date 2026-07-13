# MkDocs + Material — GitHub Pages Documentation Platform

**Date:** 2026-07-13
**Author:** Planning Agent
**Status:** Approved Design (pending user sign-off on written spec)

## Overview

Stand up an official documentation site for the Shakespeare Interpreter using MkDocs
with the Material theme, deployed to GitHub Pages. This replaces the sparse flat
`docs/` files with a navigable, searchable, mobile-friendly wiki.

## Directory Structure

MkDocs `docs_dir` is `docs/` (repo root `mkdocs.yml`). Existing `docs/superpowers/`
is excluded from the published site via `exclude_docs`.

```
docs/
├── index.md                        # Home / Introduction
├── getting-started/
│   ├── installation.md             # Build from source, Docker, binaries
│   └── usage.md                    # CLI quickstart: run, tokens, ast, repl
├── spl/
│   ├── specification.md            # ← was docs/SPL_SPECIFICATION.md
│   └── errors.md                   # ← was docs/ERROR_TAXONOMY.md
├── architecture/
│   ├── overview.md                 # Pipeline: lex → parse → analyze → execute
│   ├── lexer.md                    # Token types, scan strategy
│   ├── parser.md                   # Recursive descent, AST nodes, dictionary
│   ├── semantic.md                 # Symbol table, stage manager, M-codes
│   └── runtime.md                  # Environment, trampoline execution, I/O
├── cli/
│   ├── commands.md                 # run, tokens, ast, repl, version, about
│   └── repl.md                     # Replay-based buffer model, auto-declaration
├── contributing.md                 # Dev workflow, PR checklist, `task check`
├── about.md                        # Credits, license, related projects
└── assets/
    ├── stylesheets/
    │   └── extra.css               # Placeholder for future custom CSS
    └── images/                     # Pipeline diagrams, screenshots
```

`docs/superpowers/` stays on disk but is excluded from the MkDocs build.
`.gitkeep` at `docs/.gitkeep` is also excluded.

## Configuration (`mkdocs.yml`)

Material theme with slate dark-mode palette, tabs + sections navigation, code-copy
button, and GitHub repo integration.

```yaml
site_name: Shakespeare Interpreter
site_description: A Go-based interpreter for the Shakespeare Programming Language
site_url: https://lorenzobandini.github.io/shakespeare-interpreter-go
repo_url: https://github.com/lorenzobandini/shakespeare-interpreter-go
repo_name: lorenzobandini/shakespeare-interpreter-go
edit_uri: edit/main/docs/

theme:
  name: material
  palette:
    scheme: slate
    primary: indigo
    accent: deep orange
  features:
    - navigation.tabs
    - navigation.sections
    - toc.integrate
    - content.code.copy

plugins:
  - search
  - minify:
      minify_html: true

markdown_extensions:
  - admonition
  - pymdownx.superfences
  - pymdownx.tabbed:
      alternate_style: true
  - pymdownx.highlight
  - pymdownx.inlinehilite
  - toc:
      permalink: true

nav:
  - Home: index.md
  - Getting Started:
    - Installation: getting-started/installation.md
    - Usage: getting-started/usage.md
  - SPL Language:
    - Specification: spl/specification.md
    - Error Taxonomy: spl/errors.md
  - Architecture:
    - Pipeline Overview: architecture/overview.md
    - Lexer: architecture/lexer.md
    - Parser: architecture/parser.md
    - Semantic Analysis: architecture/semantic.md
    - Runtime: architecture/runtime.md
  - CLI Reference:
    - Commands: cli/commands.md
    - REPL: cli/repl.md
  - Contributing: contributing.md
  - About: about.md

extra:
  social:
    - icon: fontawesome/brands/github
      link: https://github.com/lorenzobandini/shakespeare-interpreter-go

exclude_docs:
  - superpowers/
  - .gitkeep
```

## Python Dependencies (`requirements-docs.txt`)

A single-file dependency manifest kept at the repo root for CI and local use:

```
mkdocs-material~=9.6
mkdocs-minify-plugin~=0.8
```

## CI/CD Workflow (`.github/workflows/docs.yml`)

Triggered on pushes/PRs to `main` touching `docs/**`, `mkdocs.yml`, or
`requirements-docs.txt`. On PRs: dry-run `mkdocs build --strict`. On `main` push:
build + deploy via `mkdocs gh-deploy --force` (no external actions).

```yaml
name: Docs
on:
  push:
    branches: [main]
    paths:
      - 'docs/**'
      - 'mkdocs.yml'
      - 'requirements-docs.txt'
  pull_request:
    branches: [main]
    paths:
      - 'docs/**'
      - 'mkdocs.yml'
      - 'requirements-docs.txt'
jobs:
  build-and-deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-python@v5
        with:
          python-version: '3.x'
      - name: Install dependencies
        run: pip install -r requirements-docs.txt
      - name: Build (strict)
        run: mkdocs build --strict
      - name: Deploy to gh-pages
        if: github.ref == 'refs/heads/main'
        run: mkdocs gh-deploy --force
```

## Verification Steps

1. **Local preview**: `pip install -r requirements-docs.txt && mkdocs serve`
   → confirms live reload at `http://127.0.0.1:8000/`.
2. **Strict build**: `mkdocs build --strict` → catches dead links and missing refs.
3. **Nav completeness**: every `.md` under `docs/` (excl. `superpowers/`) appears in
   `mkdocs.yml` nav tree.
4. **Content parity**: `spl/specification.md` and `spl/errors.md` are faithful
   copies of the originals (no content drift).
5. **CI dry-run**: push a doc-only PR, confirm the Docs workflow runs
   `mkdocs build --strict` without deploying.
6. **Deploy smoke-test**: after merge to `main`, visit the published site and
   confirm all pages render, search works, and GitHub edit links resolve.

## Implementation Order

1. Write `requirements-docs.txt` (3 lines).
2. Write `mkdocs.yml` with nav, theme, plugins, exclude_docs.
3. Create `docs/` subdirectory structure (8 dirs: `getting-started`, `spl`,
   `architecture`, `cli`, `assets/stylesheets`, `assets/images`).
4. Move `SPL_SPECIFICATION.md` → `docs/spl/specification.md`.
5. Move `ERROR_TAXONOMY.md` → `docs/spl/errors.md`.
6. Write content pages: `index.md`, `getting-started/*`, `architecture/*`,
   `cli/*`, `contributing.md`, `about.md`.
7. Create `assets/stylesheets/extra.css` (empty).
8. Write `.github/workflows/docs.yml`.
9. Verify locally (steps 1–4 above).
10. Commit and push.
