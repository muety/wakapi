<p align="center">
  <img src="static/assets/images/logo-gh.svg" width="350">
</p>

<p align="center">
  <img src="https://badges.fw-web.space/github/license/muety/wakapi">
  <a href="https://liberapay.com/muety/"><img src="https://badges.fw-web.space/liberapay/receives/muety.svg?logo=liberapay"></a>
  <img src="https://badges.fw-web.space/endpoint?url=https://wakapi.dev/api/compat/shields/v1/n1try/interval:any/project:wakapi&color=blue&label=wakapi">
  <img src="https://badges.fw-web.space/github/languages/code-size/muety/wakapi">
</p>

<p align="center">
  <a href="https://goreportcard.com/report/github.com/muety/wakapi"><img src="https://goreportcard.com/badge/github.com/muety/wakapi"></a>
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
    <a href="#%EF%B8%8F-how-to-use">How to use</a>
    <span> | </span>
    <a href="https://github.com/muety/wakapi/issues">Issues</a>
    <span> | </span>
    <a href="https://github.com/muety">Contact</a>
  </h3>
</div>

<p align="center">
  <img src="static/assets/images/screenshot.png" width="500px">
</p>

## Table of Contents
* [User Survey](#-user-survey)
* [Features](#-features)
* [Roadmap](#-roadmap)
* [How to use](#-how-to-use)
* [Configuration Options](#-configuration-options)
* [API Endpoints](#-api-endpoints)
* [Integrations](#-integrations)
* [Best Practices](#-best-practices)
* [Tests](#-tests)
* [Developer Notes](#-developer-notes)
* [Support](#-support)
* [FAQs](#-faqs)

Further instructions can be found in the [Wiki](https://github.com/muety/wakapi/wiki).

## 📬 **User Survey**
I'd love to get some community feedback from active Wakapi users. If you want, please participate in the recent [user survey](https://github.com/muety/wakapi/issues/82). Thanks a lot!

## 🚀 Features
* ✅ 100 % free and open-source
* ✅ Built by developers for developers
* ✅ Statistics for projects, languages, editors, hosts and operating systems
* ✅ Badges
* ✅ Weekly E-Mail Reports
* ✅ REST API
* ✅ Partially compatible with WakaTime
* ✅ WakaTime integration
* ✅ Support for Prometheus exports
* ✅ Lightning fast
* ✅ Self-hosted

## 🚧 Roadmap
Plans for the near future mainly include, besides usual improvements and bug fixes, a UI redesign as well as additional types of charts and statistics (see [#101](https://github.com/muety/wakapi/issues/101), [#80](https://github.com/muety/wakapi/issues/80), [#76](https://github.com/muety/wakapi/issues/76), [#12](https://github.com/muety/wakapi/issues/12)). If you have feature requests or any kind of improvement proposals feel free to open an issue or share them in our [user survey](https://github.com/muety/wakapi/issues/82). 

## ⌨️ How to use?
There are different options for how to use Wakapi, ranging from our hosted cloud service to self-hosting it. Regardless of which option choose, you will always have to do the [client setup](#-client-setup) in addition. 

### ☁️ Option 1: Use [wakapi.dev](https://wakapi.dev)
If you want to you out free, hosted cloud service, all you need to do is create an account and the set up your client-side tooling (see below).

However, we do not guarantee data persistence, so you might potentially lose your data if the service is taken down some day ❕

### 📦 Option 2: Quick-run a Release
```bash
$ curl -L https://wakapi.dev/get | bash
```

### 🐳 Option 3: Use Docker
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

If you want to run Wakapi on **Kubernetes**, there is [wakapi-helm-chart](https://github.com/andreymaznyak/wakapi-helm-chart) for quick and easy deployment.

### 🧑‍💻 Option 4: Compile and run from source
#### Prerequisites
* Go >= 1.16 (with `$GOPATH` properly set)
* gcc (to compile [go-sqlite3](https://github.com/mattn/go-sqlite3))
    * Fedora / RHEL: `dnf install @development-tools`
    * Ubuntu / Debian: `apt install build-essential`
    * Windows: See [here](https://github.com/mattn/go-sqlite3/issues/214#issuecomment-253216476)

#### Compile & Run
```bash
# Build the executable
$ go build -o wakapi

# Adapt config to your needs
$ cp config.default.yml config.yml
$ vi config.yml

# Run it
$ ./wakapi
```

**Note:** Check the comments `config.yml` for best practices regarding security configuration and more.

### 💻 Client Setup
Wakapi relies on the open-source [WakaTime](https://github.com/wakatime/wakatime) client tools. In order to collect statistics to Wakapi, you need to set them up.

1. **Set up WakaTime** for your specific IDE or editor. Please refer to the respective [plugin guide](https://wakatime.com/plugins)
2. **Editing your local `~/.wakatime.cfg`** file as follows

```ini
[settings]

# Your Wakapi server URL or 'https://wakapi.dev/api' when using the cloud server
api_url = http://localhost:3000/api

# Your Wakapi API key (get it from the web interface after having created an account)
api_key = 406fe41f-6d69-4183-a4cc-121e0c524c2b
```

Optionally, you can set up a [client-side proxy](https://github.com/muety/wakapi/wiki/Advanced-Setup:-Client-side-proxy) in addition.

## 🔧 Configuration Options
You can specify configuration options either via a config file (default: `config.yml`, customziable through the `-c` argument) or via environment variables. Here is an overview of all options.

| YAML Key                  | Environment Variable      | Default      | Description                                                         |
|---------------------------|---------------------------|--------------|---------------------------------------------------------------------|
| `env`                       | `ENVIRONMENT`               | `dev`          | Whether to use development- or production settings                  |
| `app.custom_languages`      | -                           | -              | Map from file endings to language names                             |
| `server.port`               | `WAKAPI_PORT`               | `3000`         | Port to listen on                                                   |
| `server.listen_ipv4`        | `WAKAPI_LISTEN_IPV4`        | `127.0.0.1`    | IPv4 network address to listen on (leave blank to disable IPv4)     |
| `server.listen_ipv6`        | `WAKAPI_LISTEN_IPV6`        | `::1`          | IPv6 network address to listen on (leave blank to disable IPv6)     |
| `server.listen_socket`      | `WAKAPI_LISTEN_SOCKET`      | -              | UNIX socket to listen on (leave blank to disable UNIX socket)       |
| `server.tls_cert_path`      | `WAKAPI_TLS_CERT_PATH`      | -              | Path of SSL server certificate (leave blank to not use HTTPS)       |
| `server.tls_key_path`       | `WAKAPI_TLS_KEY_PATH`       | -              | Path of SSL server private key (leave blank to not use HTTPS)       |
| `server.base_path`          | `WAKAPI_BASE_PATH`          | `/`            | Web base path (change when running behind a proxy under a sub-path) |
| `security.password_salt`    | `WAKAPI_PASSWORD_SALT`      | -              | Pepper to use for password hashing                                  |
| `security.insecure_cookies` | `WAKAPI_INSECURE_COOKIES`   | `false`        | Whether or not to allow cookies over HTTP                           |
| `security.cookie_max_age`   | `WAKAPI_COOKIE_MAX_AGE`     | `172800`       | Lifetime of authentication cookies in seconds or `0` to use [Session](https://developer.mozilla.org/en-US/docs/Web/HTTP/Cookies#Define_the_lifetime_of_a_cookie) cookies |
| `security.allow_signup`     | `WAKAPI_ALLOW_SIGNUP`       | `true`         | Whether to enable user registration                                 |
| `security.expose_metrics`   | `WAKAPI_EXPOSE_METRICS`     | `false`        | Whether to expose Prometheus metrics under `/api/metrics`               |
| `db.host`                   | `WAKAPI_DB_HOST`            | -              | Database host                                                       |
| `db.port`                   | `WAKAPI_DB_PORT`            | -              | Database port                                                       |
| `db.user`                   | `WAKAPI_DB_USER`            | -              | Database user                                                       |
| `db.password`               | `WAKAPI_DB_PASSWORD`        | -              | Database password                                                   |
| `db.name`                   | `WAKAPI_DB_NAME`            | `wakapi_db.db` | Database name                                                       |
| `db.dialect`                | `WAKAPI_DB_TYPE`            | `sqlite3`      | Database type (one of `sqlite3`, `mysql`, `postgres`, `cockroach`)  |
| `db.charset`                | `WAKAPI_DB_CHARSET`         | `utf8mb4`      | Database connection charset (for MySQL only)                        |
| `db.max_conn`               | `WAKAPI_DB_MAX_CONNECTIONS` | `2`            | Maximum number of database connections                              |
| `db.ssl`                    | `WAKAPI_DB_SSL`             | `false`        | Whether to use TLS encryption for database connection (Postgres and CockroachDB only) |
| `db.automgirate_fail_silently` | `WAKAPI_DB_AUTOMIGRATE_FAIL_SILENTLY` | `false` | Whether to ignore schema auto-migration failures when starting up |
| `mail.enabled`              | `WAKAPI_MAIL_ENABLED`       | `true`         | Whether to allow Wakapi to send e-mail (e.g. for password resets) |
| `mail.sender`               | `WAKAPI_MAIL_SENDER`        | `noreply@wakapi.dev` | Default sender address for outgoing mails (ignored for MailWhale) |
| `mail.provider`             | `WAKAPI_MAIL_PROVIDER`      | `smtp`         | Implementation to use for sending mails (one of [`smtp`, `mailwhale`]) |
| `mail.smtp.*`               | `WAKAPI_MAIL_SMTP_*`        | `-`            | Various options to configure SMTP. See [default config](config.default.yaml) for details |
| `mail.mailwhale.*`          | `WAKAPI_MAIL_MAILWHALE_*`   | `-`            | Various options to configure [MailWhale](https://mailwhale.dev) sending service. See [default config](config.default.yaml) for details |
| `sentry.dsn`                | `WAKAPI_SENTRY_DSN`         | –              | DSN for to integrate [Sentry](https://sentry.io) for error logging and tracing (leave empty to disable) |
| `sentry.enable_tracing`     | `WAKAPI_SENTRY_TRACING`     | `false`        | Whether to enable Sentry request tracing                           |
| `sentry.sample_rate`        | `WAKAPI_SENTRY_SAMPLE_RATE` | `0.75`         | Probability of tracing a request in Sentry                         |
| `sentry.sample_rate_heartbeats` | `WAKAPI_SENTRY_SAMPLE_RATE_HEARTBEATS` | `0.1` | Probability of tracing a heartbeats request in Sentry         |

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
See our [Swagger API Documentation](https://wakapi.dev/swagger-ui).

### Generating Swagger docs
```bash
$ go get -u github.com/swaggo/swag/cmd/swag
$ swag init -o static/docs
```

## 🤝 Integrations
### Prometheus Export
You can export your Wakapi statistics to Prometheus to view them in a Grafana dashboard or so. Here is how.

```bash
# 1. Start Wakapi with the feature enabled
$ export WAKAPI_EXPOSE_METRICS=true
$ ./wakapi

# 2. Get your API key and hash it
$ echo "<YOUR_API_KEY>" | base64

# 3. Add a Prometheus scrape config to your prometheus.yml (see below)
```

#### Scrape config example
```yml
# prometheus.yml
# (assuming your Wakapi instance listens at localhost, port 3000)

scrape_configs:
  - job_name: 'wakapi'
    scrape_interval: 1m
    metrics_path: '/api/metrics'
    bearer_token: '<YOUR_BASE64_HASHED_TOKEN>'
    static_configs:
      - targets: ['localhost:3000']
```

#### Grafana 
There is also a [nice Grafana dashboard](https://grafana.com/grafana/dashboards/12790), provided by the author of [wakatime_exporter](https://github.com/MacroPower/wakatime_exporter).

![](https://grafana.com/api/dashboards/12790/images/8741/image)

### WakaTime Integration
Wakapi plays well together with [WakaTime](https://wakatime.com). For one thing, you can **forward heartbeats** from Wakapi to WakaTime to effectively use both services simultaneously. In addition, there is the option to **import historic data** from WakaTime for consistency between both services. Both features can be enabled in the _Integrations_ section of your Wakapi instance's settings page.     

### GitHub Readme Stats Integrations
Wakapi also integrates with [GitHub Readme Stats](https://github.com/anuraghazra/github-readme-stats#wakatime-week-stats) to generate fancy cards for you. Here is an example.

![](https://github-readme-stats.vercel.app/api/wakatime?username=n1try&api_domain=wakapi.dev&bg_color=2D3748&title_color=2F855A&icon_color=2F855A&text_color=ffffff&custom_title=Wakapi%20Week%20Stats&layout=compact)

<details>
<summary>Click to view code</summary>

```md
![](https://github-readme-stats.vercel.app/api/wakatime?username={yourusername}&api_domain=wakapi.dev&bg_color=2D3748&title_color=2F855A&icon_color=2F855A&text_color=ffffff&custom_title=Wakapi%20Week%20Stats&layout=compact)
```

</details>
<br>


### Github Readme Metrics Integration
There is a [WakaTime plugin](https://github.com/lowlighter/metrics/tree/master/source/plugins/wakatime) for GitHub [metrics](https://github.com/lowlighter/metrics/) that is also compatible with Wakapi.

Preview:

![](https://raw.githubusercontent.com/lowlighter/lowlighter/master/metrics.plugin.wakatime.svg)

<details>
<summary>Click to view code</summary>

```yml
- uses: lowlighter/metrics@latest
  with:
    # ... other options
    plugin_wakatime: yes
    plugin_wakatime_token: ${{ secrets.WAKATIME_TOKEN }}      # Required
    plugin_wakatime_days: 7                                   # Display last week stats
    plugin_wakatime_sections: time, projects, projects-graphs # Display time and projects sections, along with projects graphs
    plugin_wakatime_limit: 4                                  # Show 4 entries per graph
    plugin_wakatime_url: http://wakapi.dev                  # Wakatime url endpoint
    plugin_wakatime_user: .user.login                         # User

```

</details>
<br>

## 👍 Best Practices
It is recommended to use wakapi behind a **reverse proxy**, like [Caddy](https://caddyserver.com) or _nginx_ to enable **TLS encryption** (HTTPS).
However, if you want to expose your wakapi instance to the public anyway, you need to set `server.listen_ipv4` to `0.0.0.0` in `config.yml`

## 🧪 Tests
### Unit Tests
Unit tests are supposed to test business logic on a fine-grained level. They are implemented as part of the application, using Go's [testing](https://pkg.go.dev/testing?utm_source=godoc) package alongside [stretchr/testify](https://pkg.go.dev/github.com/stretchr/testify).

#### How to run
```bash
$ CGO_FLAGS="-g -O2 -Wno-return-local-addr" go test -json -coverprofile=coverage/coverage.out ./... -run ./...
```

### API Tests
API tests are implemented as black box tests, which interact with a fully-fledged, standalone Wakapi through HTTP requests. They are supposed to check Wakapi's web stack and endpoints, including response codes, headers and data on a syntactical level, rather than checking the actual content that is returned.

Our API (or end-to-end, in some way) tests are implemented as a [Postman](https://www.postman.com/) collection and can be run either from inside Postman, or using [newman](https://www.npmjs.com/package/newman) as a command-line runner.

To get a predictable environment, tests are run against a fresh and clean Wakapi instance with a SQLite database that is populated with nothing but some seed data (see [data.sql](testing/data.sql)). It is usually recommended for software tests to be [safe](https://www.restapitutorial.com/lessons/idempotency.html), stateless and without side effects. In contrary to that paradigm, our API tests strictly require a fixed execution order (which Postman assures) and their assertions may rely on specific previous tests having succeeded.

#### Prerequisites (Linux only)
```bash
# 1. sqlite (cli)
$ sudo apt install sqlite  # Fedora: sudo dnf install sqlite

# 2. screen
$ sudo apt install screen  # Fedora: sudo dnf install screen

# 3. newman
$ npm install -g newman
```

#### How to run (Linux only) 
```bash
$ ./testing/run_api_tests.sh
```

## 🤓 Developer Notes
### Building web assets
To keep things minimal, Wakapi does not contain a `package.json`, `node_modules` or any sort of frontend build step. Instead, all JS and CSS assets are included as static files and checked in to Git. This way we can avoid requiring NodeJS to build Wakapi. However, for [TailwindCSS](https://tailwindcss.com/docs/installation#building-for-production) it makes sense to run it through a "build" step to benefit from purging and significantly reduce it in size. To only require this at the time of development, the compiled asset is checked in to Git as well. Similarly, [Iconify](https://iconify.design/docs/icon-bundles/) bundles are also created at development time and checked in to the repo. 

#### TailwindCSS
```bash
$ tailwindcss-cli build static/assets/vendor/tailwind.css -o static/assets/vendor/tailwind.dist.css
```

#### Iconify
```bash
$ yarn add -D @iconify/json-tools @iconify/json
$ node scripts/bundle_icons.js
```

New icons can be added by editing the `icons` array in [scripts/bundle_icons.js](scripts/bundle_icons.js).

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
