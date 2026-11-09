---
title: Authentication
description: How to authenticate against the Wakapi API with an API key.
---

# Authentication

All authenticated endpoints take an API key in the `Authorization` header. The key is your Wakapi API key, Base64-encoded and prefixed with `Basic` (as documented in the spec itself):

```http
Authorization: Basic <base64-of-your-api-key>
```

Example from the spec:

```http
Authorization: Basic ODY2NDhkNzQtMTljNS00NTJiLWJhMDEtZmIzZWM3MGQ0YzJmCg==
```

Get your API key from your Wakapi instance: log in, open your profile/settings page, and copy the key shown there. Keys are per user; keep yours private and never commit it.

Endpoints under `/compat/wakatime/v1/` follow the WakaTime API shape and accept the same header. The Shields badge endpoint (`/compat/shields/v1/{user}/{interval}/{filter}`) needs the user to have public data access enabled instead of a key.
