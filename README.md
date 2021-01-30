<h1 align="center">📊 Wakapi</h1>

<p align="center">
  <img src="https://badges.fw-web.space/github/license/muety/wakapi">
  <a href="https://saythanks.io/to/n1try"><img src="https://badges.fw-web.space/badge/SayThanks.io-%E2%98%BC-1EAEDB.svg"></a>
  <a href="https://liberapay.com/muety/"><img src="https://badges.fw-web.space/liberapay/receives/muety.svg?logo=liberapay"></a>
  <img src="https://badges.fw-web.space/endpoint?url=https://wakapi.dev/api/compat/shields/v1/n1try/interval:any/project:wakapi&color=blue&label=wakapi">
</p>

<p align="center">
  <a href="https://goreportcard.com/report/github.com/muety/wakapi"><img src="https://goreportcard.com/badge/github.com/muety/wakapi"></a>
  <img src="https://badges.fw-web.space/github/languages/code-size/muety/wakapi">
  <a href="https://sonarcloud.io/dashboard?id=muety_wakapi"><img src="https://sonarcloud.io/api/project_badges/measure?project=muety_wakapi&metric=sqale_index"></a>
  <a href="https://sonarcloud.io/dashboard?id=muety_wakapi"><img src="https://sonarcloud.io/api/project_badges/measure?project=muety_wakapi&metric=ncloc"></a>
</p>

<h3 align="center">A minimalist, self-hosted WakaTime-compatible backend for coding statistics.</h3>

<div align="center">
  <h3>
    <a href="https://wakapi.dev">Website</a>
    <span> | </span>
    <a href="#-features">Features</a>
    <span> | </span>
    <a href="#-how-to-use">How to use</a>
    <span> | </span>
    <a href="https://github.com/muety/wakapi/issues">Issues</a>
    <span> | </span>
    <a href="https://github.com/muety">Contact</a>
  </h3>
</div>

<p align="center">
  <img src="https://anchr.io/i/bxQ69.png" width="500px">
</p>

## Table of Contents
* [User Survey](#-user-survey)
* [Features](#-features)
* [How to use](#-how-to-use)
* [Configuration Options](#-configuration-options)
* [API Endpoints](#-api-endpoints)
* [Prometheus Export](#-prometheus-export)
* [Best Practices](#-best-practices)
* [Developer Notes](#-developer-notes)
* [Support](#-support)
* [FAQs](#-faqs)

## 📬 **User Survey**
I'd love to get some community feedback from active Wakapi users. If you want, please participate in the recent [user survey](https://github.com/muety/wakapi/issues/82). Thanks a lot!

## 🚀 Features
* ✅ 100 % free and open-source
* ✅ Built by developers for developers
* ✅ Statistics for projects, languages, editors, hosts and operating systems
* ✅ Badges
* ✅ REST API
* ✅ Partially compatible with WakaTime
* ✅ WakaTime relay to use both
* ✅ Support for [Prometheus](https://github.com/muety/wakapi#%EF%B8%8F-prometheus-export) exports
* ✅ Self-hosted

## ⌨️ How to use?
There are different options for how to use Wakapi, ranging from out hosted cloud service to self-hosting it. Regardless of which option choose, you will always have to do the [client setup](#-client-setup) in addition. 

### ☁️ Option 1: Use [wakapi.dev](https://wakapi.dev)
If you want to you out free, hosted cloud service, all you need to do is create an account and the set up your client-side tooling (see below).

However, we do not guarantee data persistence, so you might potentially lose your data if the service is taken down some day ❕

### 🐳 Option 2: Use Docker
```bash
# Create a persistent volume
$ docker volume create wakapi-data

# Run the container
$ docker run -d \
  -p 3000:3000 \
  -e "WAKAPI_PASSWORD_SALT=$(cat /dev/urandom | tr -dc 'a-zA-Z0-9' | fold -w ${1:-32} | head -n 1)" \
  -v wakapi-data:/data \
  --name wakapi n1try/wakapi
```

**Note:** By default, SQLite is used as a database. To run Wakapi in Docker with MySQL or Postgres, see [Dockerfile](https://github.com/muety/wakapi/blob/master/Dockerfile) and [config.default.yml](https://github.com/muety/wakapi/blob/master/config.default.yml) for further options.

### 📦 Option 3: Run a release
```bash
# Download the release and unpack it
$ wget https://github.com/muety/wakapi/releases/download/1.20.2/wakapi_linux_amd64.zip
$ unzip wakapi_linux_amd64.zip

# Optionally adapt config to your needs
$ vi config.yml

# Run it
$ ./wakapi
```

### 🧑‍💻 Option 4: Run from source
#### Prerequisites
* Go >= 1.13 (with `$GOPATH` properly set)
* gcc (to compile [go-sqlite3](https://github.com/mattn/go-sqlite3))
    * Fedora / RHEL: `dnf install @development-tools`
    * Ubuntu / Debian: `apt install build-essential`
    * Windows: See [here](https://github.com/mattn/go-sqlite3/issues/214#issuecomment-253216476)

#### Compile & Run
```bash
# Adapt config to your needs
$ cp config.default.yml config.yml
$ vi config.yml

# Install packaging tool
$ export GO111MODULE=on
$ go get github.com/markbates/pkger/cmd/pkger

# Build the executable
$ go generate
$ go build -o wakapi

# Run it
$ ./wakapi
```

**Note:** By default, the application is running in dev mode. However, it is recommended to set `ENV=production` for enhanced performance and security. To still be able to log in when using production mode, you either have to run Wakapi behind a reverse proxy, that enables for HTTPS encryption (see [best practices](#best-practices)) or set `security.insecure_cookies = true` in `config.yml`.

### 💻 Client Setup
Wakapi relies on the open-source [WakaTime](https://github.com/wakatime/wakatime) client tools. In order to collect statistics to Wakapi, you need to set them up.

1. **Set up WakaTime** for your specific IDE or editor. Please refer to the respective [plugin guide](https://wakatime.com/plugins)
2. **Editing your local `~/.wakatime.cfg`** file as follows

```
# Your Wakapi server URL or 'https://wakapi.dev/api/heartbeat' when using the cloud server
api_url = http://localhost:3000/api/heartbeat

# Your Wakapi API key (get it from the web interface after having created an account)
api_key = 406fe41f-6d69-4183-a4cc-121e0c524c2b
```

Optionally, you can set up a [client-side proxy](docs/advanced_setup.md) in addition.

## 🔧 Configuration Options
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
| `db.dialect`                | `WAKAPI_DB_TYPE`            | `sqlite3`      | Database type (one of `sqlite3`, `mysql`, `postgres`, `cockroach`)                     |
| `db.max_conn`               | `WAKAPI_DB_MAX_CONNECTIONS` | `2`            | Maximum number of database connections                              |
| `db.ssl`                    | `WAKAPI_DB_SSL`             | `false`        | Whether to use TLS encryption for database connection (Postgres and CockroachDB only) |

### Supported databases
Wakapi uses [GORM](https://gorm.io) as an ORM. As a consequence, a set of different relational databases is supported.
* [SQLite](https://sqlite.org/) (_default, easy setup_)
* [MySQL](https://hub.docker.com/_/mysql) (_recommended, because most extensively tested_)
* [MariaDB](https://hub.docker.com/_/mariadb) (_open-source MySQL alternative_)
* [Postgres](https://hub.docker.com/_/postgres) (_open-source as well_)
* [CockroachDB](https://www.cockroachlabs.com/docs/stable/install-cockroachdb-linux.html) (_cloud-native, distributed, Postgres-compatible API_)

### Client-side proxy (`optional`)
See the [advanced setup instructions](docs/advanced_setup.md).

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

## 👍 Best Practices
It is recommended to use wakapi behind a **reverse proxy**, like [Caddy](https://caddyserver.com) or _nginx_ to enable **TLS encryption** (HTTPS).
However, if you want to expose your wakapi instance to the public anyway, you need to set `server.listen_ipv4` to `0.0.0.0` in `config.yml`

## 🤓 Developer Notes
### Running tests
```bash
CGO_FLAGS="-g -O2 -Wno-return-local-addr" go test -json -coverprofile=coverage/coverage.out ./... -run ./...
```

## 🙏 Support
If you like this project, please consider supporting it 🙂. You can donate either through [buying me a coffee](https://buymeacoff.ee/n1try) or becoming a GitHub sponsor. Every little donation is highly appreciated and boosts the developers' motivation to keep improving Wakapi!

## ❔ FAQs
Since Wakapi heavily relies on the concepts provided by WakaTime, [their FAQs](https://wakatime.com/faq) apply to Wakapi for large parts as well. You might find answers there.

<details>
<summary><b>What data is sent to Wakapi?</b></summary>

<ul>
  <li>File names</li>
  <li>Project names</li>
  <li>Editor names</li>
  <li>You computer's host name</li>
  <li>Timestamps for every action you take in your editor</li>
  <li>...</li>
</ul>

See the related [WakaTime FAQ section](https://wakatime.com/faq#data-collected) for details.

If you host Wakapi yourself, you have control over all your data. However, if you use our webservice and are concerned about privacy, you can also [exclude or obfuscate](https://wakatime.com/faq#exclude-paths) certain file- or project names.  
</details>

<details>
<summary><b>What happens if I'm offline?</b></summary>

All data is cached locally on your machine and sent in batches once you're online again.
</details>

<details>
<summary><b>How did Wakapi come about?</b></summary>

Wakapi was started when I was a student, who wanted to track detailed statistics about my coding time. Although I'm a big fan of WakaTime I didn't want to pay <a href="https://wakatime.com/pricing)">9 $ a month</a> back then. Luckily, most parts of WakaTime are open source!  
</details>

<details>
<summary><b>How does Wakapi compare to WakaTime?</b></summary>

Wakapi is a small subset of WakaTime and has a lot less features. Cool WakaTime features, that are missing Wakapi, include:

<ul>
  <li>Leaderboards</li>
  <li><a href="https://wakatime.com/share/embed">Embeddable Charts</a></li>
  <li>Personal Goals</li>
  <li>Team- / Organization Support</li>
  <li>Integrations (with GitLab, etc.)</li>
  <li>Richer API</li>
</ul>

WakaTime is worth the price. However, if you only want basic statistics and keep sovereignty over your data, you might want to go with Wakapi.
</details>

<details>
<summary><b>How are durations calculated?</b></summary>

Inferring a measure for your coding time from heartbeats works a bit different than in WakaTime. While WakaTime has <a href="https://wakatime.com/faq#timeout">timeout intervals</a>, Wakapi essentially just pads every heartbeat, that occurs after a longer pause, with 2 extra minutes.

Here is an example (circles are heartbeats):

```
|---o---o--------------o---o---|
|   |10s|      3m      |10s|   |

```

It is unclear how to handle the three minutes in between. Did the developer do a 3-minute break or were just no heartbeats being sent, e.g. because the developer was starring at the screen find a solution, but not actually typing code.

<ul>
  <li><b>WakaTime</b> (with 5 min timeout): 3 min 20 sec
  <li><b>WakaTime</b> (with 2 min timeout): 20 sec
  <li><b>Wakapi:</b> 10 sec + 2 min + 10 sec = 2 min 20 sec</li>
</ul>

Wakapi adds a "padding" of two minutes before the third heartbeat. This is why total times will slightly vary between Wakapi and WakaTime.
</details>

## 🙏 Thanks
I highly appreciate the efforts of [@alanhamlett](https://github.com/alanhamlett) and the WakaTime team and am thankful for their software being open source. 

## 📓 License
GPL-v3 @ [Ferdinand Mütsch](https://muetsch.io)
