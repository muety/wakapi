---
title: Wakapi API Reference
description: A navigable Sourcey-generated reference for the Wakapi REST API, built from the project's own OpenAPI spec.
---

# Wakapi API Reference

Wakapi is a minimalist, self-hosted, WakaTime-compatible backend for coding statistics ([wakapi.dev](https://wakapi.dev), [github.com/muety/wakapi](https://github.com/muety/wakapi), MIT licensed).

This site is a generated reference for the Wakapi REST API: 22 operations across summaries, heartbeats, projects, statistics, leaders, badges, and diagnostics. It was built with [Sourcey](https://sourcey.com) from the OpenAPI spec shipped in the Wakapi repository itself (`static/docs/swagger.yaml`), pinned to commit [`c561bbc174d139d2d3f8ee222ca61671cdc3caf1`](https://github.com/muety/wakapi/commit/c561bbc174d139d2d3f8ee222ca61671cdc3caf1) (2026-11-08).

## Why this exists

Wakapi's API is documented today only as a raw Swagger file inside the repository and a Swagger UI served by a running instance. There is no public, linkable API reference a user or integrator can read without running the server first. This site closes that gap: every endpoint, parameter, and schema from the shipped spec, in one navigable place.

## What is covered

- **Summaries** — per-user coding summaries, status bar, ranges, all-time views
- **Heartbeats** — native and WakaTime-compatible heartbeat ingestion (single and bulk)
- **Projects, statistics, leaders** — filtered project lists, stats ranges, leaderboards
- **Badges and diagnostics** — Shields.io-compatible badge data, plugin error reports
- **Health** — application health status

Start with [Authentication](authentication) for the `Authorization` header, then [Sending heartbeats](heartbeats) for a worked `curl` example. The full endpoint list is under **API Reference**.
