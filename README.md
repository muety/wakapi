# 📈 wakapi

![](https://badges.fw-web.space/github/license/muety/wakapi)
![GitHub release (latest by date)](https://badges.fw-web.space/github/v/release/muety/wakapi)
![GitHub code size in bytes](https://img.shields.io/github/languages/code-size/muety/wakapi)
![Docker Cloud Build Status](https://badges.fw-web.space/docker/cloud/build/n1try/wakapi)
![GitHub issues](https://img.shields.io/github/issues/muety/wakapi)
![GitHub last commit](https://img.shields.io/github/last-commit/muety/wakapi)
[![Say thanks](https://badges.fw-web.space/badge/SayThanks.io-%E2%98%BC-1EAEDB.svg)](https://saythanks.io/to/n1try)
[![](https://badges.fw-web.space/liberapay/receives/muety.svg?logo=liberapay)](https://liberapay.com/muety/)
![GitHub go.mod Go version](https://badges.fw-web.space/github/go-mod/go-version/muety/wakapi)
[![Go Report Card](https://goreportcard.com/badge/github.com/muety/wakapi)](https://goreportcard.com/report/github.com/muety/wakapi)
![Coding Activity](https://badges.fw-web.space/endpoint?url=https://wakapi.dev/api/compat/shields/v1/n1try/interval:any/project:wakapi&color=blue)
[![Security Rating](https://sonarcloud.io/api/project_badges/measure?project=muety_wakapi&metric=security_rating)](https://sonarcloud.io/dashboard?id=muety_wakapi)
[![Maintainability Rating](https://sonarcloud.io/api/project_badges/measure?project=muety_wakapi&metric=sqale_rating)](https://sonarcloud.io/dashboard?id=muety_wakapi)
[![Technical Debt](https://sonarcloud.io/api/project_badges/measure?project=muety_wakapi&metric=sqale_index)](https://sonarcloud.io/dashboard?id=muety_wakapi)
[![Lines of Code](https://sonarcloud.io/api/project_badges/measure?project=muety_wakapi&metric=ncloc)](https://sonarcloud.io/dashboard?id=muety_wakapi)
---

**A minimalist, self-hosted WakaTime-compatible backend for coding statistics**

![Wakapi screenshot](https://anchr.io/i/bxQ69.png)

## 📬 **User Survey**
I'd love to get some community feedback from active Wakapi users. If you like, please participate in the recent [user survey](https://github.com/muety/wakapi/issues/82). Thanks a lot!

## 👀  Demo
🔥 **New:** Wakapi is available as a hosted service now. Check out **[wakapi.dev](https://wakapi.dev)**. Please use responsibly.

To use the hosted version set `api_url = https://wakapi.dev/api/heartbeat`. However, we do not guarantee data persistence, so you might potentially lose your data if the service is taken down some day ❕

## ⚙️ Prerequisites
**On the server side:**
* Go >= 1.13 (with `$GOPATH` properly set)
* gcc (to compile [go-sqlite3](https://github.com/mattn/go-sqlite3))
    * Fedora / RHEL: `dnf install @development-tools`
    * Ubuntu / Debian: `apt install build-essential`
    * Windows: See [here](https://github.com/mattn/go-sqlite3/issues/214#issuecomment-253216476)
* _Optional_: A MySQL- or Postgres database

**On your local machine:**
* [WakaTime plugin](https://wakatime.com/plugins) for your editor / IDE

## ⌨️ Server Setup
### Run from source
1. Clone the project
1. Copy `config.default.yml` to `config.yml` and adapt it to your needs
1. Build executable: `GO111MODULE=on go build`
1. Run server: `./wakapi`

**As an alternative** to building from source you can also grab a pre-built [release](https://github.com/muety/wakapi/releases). Steps 2, 3 and 5 apply analogously.

**Note:** By default, the application is running in dev mode. However, it is recommended to set `ENV=production` for enhanced performance and security. To still be able to log in when using production mode, you either have to run Wakapi behind a reverse proxy, that enables for HTTPS encryption (see [best practices](#best-practices)) or set `security.insecure_cookies` to `true` in `config.yml`. 

### Run with Docker
```bash
docker run -d -p 3000:3000 -e "WAKAPI_PASSWORD_SALT=$(cat /dev/urandom | tr -dc 'a-zA-Z0-9' | fold -w ${1:-32} | head -n 1)" --name wakapi n1try/wakapi
```

By default, SQLite is used as a database. To run Wakapi in Docker with MySQL or Postgres, see [Dockerfile](https://github.com/muety/wakapi/blob/master/Dockerfile) and [config.default.yml](https://github.com/muety/wakapi/blob/master/config.default.yml) for further options.

### Running tests
```bash
CGO_FLAGS="-g -O2 -Wno-return-local-addr" go test -json -coverprofile=coverage/coverage.out ./... -run ./...
```

## 🔧 Configuration
You can specify configuration options either via a config file (default: `config.yml`, customziable through the `-c` argument) or via environment variables. Here is an overview of all options.

| YAML Key                  | Environment Variable      | Default      | Description                                                         |
|---------------------------|---------------------------|--------------|---------------------------------------------------------------------|
| `env`                       | `ENVIRONMENT`               | `dev`          | Whether to use development- or production settings                  |
| `app.custom_languages`      | -                           | -              | Map from file endings to language names                             |
| `server.port`               | `WAKAPI_PORT`               | `3000`         | Port to listen on                                                   |
| `server.listen_ipv4`        | `WAKAPI_LISTEN_IPV4`        | `127.0.0.1`    | IPv4 network address to listen on (leave blank to disable IPv4)     |
| `server.listen_ipv6`        | `WAKAPI_LISTEN_IPV6`        | `::1`          | IPv6 network address to listen on (leave blank to disable IPv6)     |
| `server.tls_cert_path`      | `WAKAPI_TLS_CERT_PATH`      | -              | Path of SSL server certificate (leave blank to not use HTTPS)       |
| `server.tls_key_path`       | `WAKAPI_TLS_KEY_PATH`       | -              | Path of SSL server private key (leave blank to not use HTTPS)       |
| `server.base_path`          | `WAKAPI_BASE_PATH`          | `/`            | Web base path (change when running behind a proxy under a sub-path) |
| `security.password_salt`    | `WAKAPI_PASSWORD_SALT`      | -              | Pepper to use for password hashing                                  |
| `security.insecure_cookies` | `WAKAPI_INSECURE_COOKIES`   | `false`        | Whether or not to allow cookies over HTTP                           |
| `security.cookie_max_age`   | `WAKAPI_COOKIE_MAX_AGE  `   | `172800`       | Lifetime of authentication cookies in seconds or `0` to use [Session](https://developer.mozilla.org/en-US/docs/Web/HTTP/Cookies#Define_the_lifetime_of_a_cookie) cookies |
| `db.host`                   | `WAKAPI_DB_HOST`            | -              | Database host                                                       |
| `db.port`                   | `WAKAPI_DB_PORT`            | -              | Database port                                                       |
| `db.user`                   | `WAKAPI_DB_USER`            | -              | Database user                                                       |
| `db.password`               | `WAKAPI_DB_PASSWORD`        | -              | Database password                                                   |
| `db.name`                   | `WAKAPI_DB_NAME`            | `wakapi_db.db` | Database name                                                       |
| `db.dialect`                | `WAKAPI_DB_TYPE`            | `sqlite3`      | Database type (one of sqlite3, mysql, postgres)                     |
| `db.max_conn`               | `WAKAPI_DB_MAX_CONNECTIONS` | `2`            | Maximum number of database connections                              |

## 💻 Client Setup
Wakapi relies on the open-source [WakaTime](https://github.com/wakatime/wakatime) client tools. In order to collect statistics to Wakapi, you need to set them up.

1. **Set up WakaTime** for your specific IDE or editor. Please refer to the respective [plugin guide](https://wakatime.com/plugins)
2. Make your local WakaTime client talk to Wakapi by **editing your local `~/.wakatime.cfg`** file as follows

```
api_url = https://your.server:someport/api/heartbeat`
api_key = the_api_key_printed_to_the_console_after_starting_the_server`
```

You can view your API Key after logging in to the web interface.

### Optional: Client-side proxy

See the [advanced setup instructions](docs/advanced_setup.md).  

## 🔵 Customization

### Aliases
There is an option to add aliases for project names, editors, operating systems and languages. For instance, if you want to map two projects – `myapp-frontend` and `myapp-backend` – two a common project name – `myapp-web` – in your statistics, you can add project aliases.

At the moment, this can only be done via raw database queries. For the above example, you would need to add two aliases, like this:

```sql
INSERT INTO aliases (`type`, `user_id`, `key`, `value`) VALUES (0, 'your_username', 'myapp-web', 'myapp-frontend');
```

#### Types
* Project ~  type **0**
* Language ~  type **1**
* Editor ~ type **2**
* OS ~  type **3**
* Machine ~  type **4**

## 🔧 API Endpoints
The following API endpoints are available. A more detailed Swagger documentation is about to come ([#40](https://github.com/muety/wakapi/issues/40)).

* `POST /api/heartbeat`
* `GET /api/summary`
  * `string` parameter `interval`: One of `today`, `day`, `week`, `month`, `year`, `any`
* `GET /api/compat/wakatime/v1/users/current/all_time_since_today` (see [Wakatime API docs](https://wakatime.com/developers#all_time_since_today))
* `GET /api/compat/wakatime/v1/users/current/summaries` (see [Wakatime API docs](https://wakatime.com/developers#summaries))
* `GET /api/health`

## ⤴️ Prometheus Export
If you want to export your Wakapi statistics to Prometheus to view them in a Grafana dashboard or so please refer to an excellent tool called **[wakatime_exporter](https://github.com/MacroPower/wakatime_exporter)**.

[![](https://github-readme-stats.vercel.app/api/pin/?username=MacroPower&repo=wakatime_exporter&show_owner=true)](https://github.com/MacroPower/wakatime_exporter)

It is a standalone webserver that connects to your Wakapi instance and exposes the data as Prometheus metrics. Although originally developed to scrape data from WakaTime, it will mostly for with Wakapi as well, as the APIs are partially compatible.

Simply configure the exporter with `WAKA_SCRAPE_URI` to equal `"https://wakapi.your-server.com/api/compat/wakatime/v1"` and set your API key accordingly.

## 🏷 Badges
We recently introduced support for [Shields.io](https://shields.io) badges (see above). Visit your Wakapi server's settings page to see details.

## 👍 Best Practices
It is recommended to use wakapi behind a **reverse proxy**, like [Caddy](https://caddyserver.com) or _nginx_ to enable **TLS encryption** (HTTPS).
However, if you want to expose your wakapi instance to the public anyway, you need to set `server.listen_ipv4` to `0.0.0.0` in `config.yml`

## 🙏 Support
If you like this project, please consider supporting it 🙂. You can donate either through [buying me a coffee](https://buymeacoff.ee/n1try) or becoming a GitHub sponsor. Every little donation is highly appreciated and boosts the developers' motivation to keep improving Wakapi!

## ⚠️ Important Note
**This is not an alternative to using WakaTime.** It is just a custom, non-commercial, self-hosted application to collect coding statistics using the already existing editor plugins provided by the WakaTime community. It was created for personal use only and with the purpose of keeping the sovereignity of your own data. However, if you like the official product, **please support the authors and buy an official WakaTime subscription!**

## 📓 License
GPL-v3 @ [Ferdinand Mütsch](https://muetsch.io)
