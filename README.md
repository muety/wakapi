# 📈 wakapi

[![](http://img.shields.io/liberapay/receives/muety.svg?logo=liberapay&style=flat-square)](https://liberapay.com/muety/)
[![Say thanks](https://img.shields.io/badge/SayThanks.io-%E2%98%BC-1EAEDB.svg?style=flat-square)](https://saythanks.io/to/n1try)
![](https://img.shields.io/github/license/muety/wakapi?style=flat-square)
[![Go Report Card](https://goreportcard.com/badge/github.com/muety/wakapi?style=flat-square)](https://goreportcard.com/report/github.com/muety/wakapi)

[![Buy me a coffee](https://www.buymeacoffee.com/assets/img/custom_images/orange_img.png)](https://buymeacoff.ee/n1try)

---

**A minimalist, self-hosted WakaTime-compatible backend for coding statistics**

![Wakapi screenshot](https://anchr.io/i/bxQ69.png)

If you like this project, please consider supporting it 🙂. You can donate either through [buying me a coffee](https://buymeacoff.ee/n1try) or becoming a GitHub sponsor. Every little donation is highly appreciated and boosts the developers' motivation to keep improving Wakapi!

## Demo 
🔥 **New:** There is hosted [demo version](https://apps.muetsch.io/wakapi) available now. Go check it out! Please use responsibly.

To use the demo version set `api_url = https://apps.muetsch.io/wakapi/api/heartbeat`. However, this hosted instance might be taken down again in the future, so you might potentially lose your data ❕

## Prerequisites
**On the server side:**
* Go >= 1.13 (with `$GOPATH` properly set)
* gcc (to compile [go-sqlite3](https://github.com/mattn/go-sqlite3))
    * Fedora / RHEL: `dnf install @development-tools`
    * Ubuntu / Debian: `apt install build-essential`
    * Windows: See [here](https://github.com/mattn/go-sqlite3/issues/214#issuecomment-253216476)
* _Optional_: A MySQL- or Postgres database

**On your local machine:**
* [WakaTime plugin](https://wakatime.com/plugins) for your editor / IDE

## Server Setup
### Run from source
1. Clone the project
1. Copy `.env.example` to `.env` and set database credentials
1. Adapt `config.ini` to your needs
1. Build executable: `GO111MODULE=on go build`
1. Run server: `./wakapi`

**As an alternative** to building from source you can also grab a pre-built [release](https://github.com/muety/wakapi/releases). Steps 2, 3 and 5 apply analogously.

**Note:** By default, the application is running in dev mode. However, it is recommended to set `ENV=production` in `.env` for enhanced performance and security. To still be able to log in when using production mode, you either have to run Wakapi behind a reverse proxy, that enables for HTTPS encryption (see [best practices](i#best-practices)) or set `insecure_cookies = true` in `config.ini`. 

### Run with Docker
```
docker run -d -p 3000:3000 --name wakapi n1try/wakapi
```

By default, SQLite is used as a database. To run Wakapi in Docker with MySQL or Postgres, see [Dockerfile](https://github.com/muety/wakapi/blob/master/Dockerfile) and [.env.example](https://github.com/muety/wakapi/blob/master/.env.example) for further options.

## Client Setup
Wakapi relies on the open-source [WakaTime](https://github.com/wakatime/wakatime) client tools. In order to collect statistics to Wakapi, you need to set them up.

1. **Set up WakaTime** for your specific IDE or editor. Please refer to the respective [plugin guide](https://wakatime.com/plugins)
2. Make your local WakaTime client talk to Wakapi by **editing your local `~/.wakatime.cfg`** file as follows

```
api_url = https://your.server:someport/api/heartbeat`
api_key = the_api_key_printed_to_the_console_after_starting_the_server`
```

You can view your API Key after logging in to the web interface.

## Customization

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

**NOTE:** In order for the aliases to take effect for non-live statistics, you would either have to wait 24 hours for the cache to be invalidated or restart Wakapi.

## API Endpoints
The following API endpoints are available. A more detailed Swagger documentation is about to come ([#40](https://github.com/muety/wakapi/issues/40)).

* `POST /api/heartbeat`
* `GET /api/summary`
  * `string` parameter `interval`: One of `today`, `day`, `week`, `month`, `year`, `any`
  * `bool` parameter `live`: Whether to compute the summary to present time
* `GET /api/compat/v1/users/current/all_time_since_today` (see [Wakatime API docs](https://wakatime.com/developers#all_time_since_today))
* `GET /api/compat/v1/users/current/summaries` (see [Wakatime API docs](https://wakatime.com/developers#summaries)) (⏳ [coming soon](https://github.com/muety/wakapi/issues/44))
* `GET /api/health`


## Best Practices
It is recommended to use wakapi behind a **reverse proxy**, like [Caddy](https://caddyserver.com) or _nginx_ to enable **TLS encryption** (HTTPS).
However, if you want to expose your wakapi instance to the public anyway, you need to set `listen = 0.0.0.0` in `config.ini`

## Important Note
**This is not an alternative to using WakaTime.** It is just a custom, non-commercial, self-hosted application to collect coding statistics using the already existing editor plugins provided by the WakaTime community. It was created for personal use only and with the purpose of keeping the sovereignity of your own data. However, if you like the official product, **please support the authors and buy an official WakaTime subscription!**

## License
GPL-v3 @ [Ferdinand Mütsch](https://muetsch.io)
