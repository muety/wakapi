# Wakapi API Reference (Sourcey)

A navigable API reference for the Wakapi REST API, generated with
[Sourcey](https://sourcey.com) from the OpenAPI spec shipped in this
repository (`static/docs/swagger.yaml`).

- **Live site (once Pages is enabled):** https://muety.github.io/wakapi/
- **Coverage:** 22 operations, 33 schemas, plus a 3-page guide
  (getting started, authentication, sending heartbeats).

## Build locally

```bash
cd docs/api
npm install
npm run build   # output in docs/api/dist
```

## Deploy

`.github/workflows/docs.yml` builds the site on every push to `master`
that touches `docs/api/**` or `static/docs/swagger.yaml`, and deploys it
to GitHub Pages. Before the first deploy, enable Pages in the repository
settings with **Source: GitHub Actions**. No custom domain is configured;
the site serves at `https://muety.github.io/wakapi/`.
