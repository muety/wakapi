---
title: Sending heartbeats
description: A worked example of pushing a coding heartbeat to Wakapi.
---

# Sending heartbeats

A heartbeat records one coding event: a file you edited, in a project, at a time. Send it with `POST /heartbeat`:

```bash
curl -X POST https://your-wakapi-instance/api/heartbeat \
  -H "Authorization: Basic <base64-of-your-api-key>" \
  -H "Content-Type: application/json" \
  -d '{
    "entity": "/home/user/projects/my-app/src/main.go",
    "type": "file",
    "category": "coding",
    "time": 1794216000,
    "project": "my-app",
    "branch": "main",
    "language": "Go",
    "is_write": true
  }'
```

Notes:

- `entity` is the file path, `type` is usually `file`, `category` is usually `coding`, and `time` is a Unix timestamp.
- To send many heartbeats at once, use `POST /heartbeats` with a JSON array of the same objects.
- WakaTime-compatible clients can point at the `/compat/wakatime/v1/users/{user}/heartbeats` endpoints instead — same payload shape, WakaTime-compatible paths. The native `/heartbeat` and the compat endpoints overlap; prefer the native endpoints for new integrations and the compat endpoints only when an existing WakaTime client needs them.
- Check the result with `GET /summary` or `GET /users/{user}/statusbar/today`.
