# 📈 wakapi

[![](http://img.shields.io/liberapay/receives/muety.svg?logo=liberapay&style=flat-square)](https://liberapay.com/muety/)
[![Say thanks](https://img.shields.io/badge/SayThanks.io-%E2%98%BC-1EAEDB.svg?style=flat-square)](https://saythanks.io/to/n1try)
![](https://img.shields.io/github/license/muety/wakapi?style=flat-square)
[![Go Report Card](https://goreportcard.com/badge/github.com/muety/wakapi?style=flat-square)](https://goreportcard.com/report/github.com/muety/wakapi)

[![Buy me a coffee](https://www.buymeacoffee.com/assets/img/custom_images/orange_img.png)](https://buymeacoff.ee/n1try)

---

**A minimalist, self-hosted WakaTime-compatible backend for coding statistics**

![Wakapi screenshot](https://anchr.io/i/bxQ69.png)

If you like this project, please consider supporting it 🙂. You can donate either through [buying me a coffee](https://buymeacoff.ee/n1try) or becoming a GitHub sponor. Every little donation is highly appreciated and boosts the developers' motivation to keep improving Wakapi!

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

### Run with Docker
```
docker run -d -p 3000:3000 --name wakapi n1try/wakapi
```

To get your API key, take a look into the logs `docker logs wakapi | grep "API key"`

In addition, you can specify several environment variables for configuration:
* `-e WAKAPI_DEFAULT_USER_NAME=admin`
* `-e WAKAPI_DEFAULT_USER_PASSWORD=admin`

By default, SQLite is used as a database. To run Wakapi in Docker with MySQL or Postgres, see [Dockerfile](https://github.com/muety/wakapi/blob/master/Dockerfile) and [.env.example](https://github.com/muety/wakapi/blob/master/.env.example) for further options.

## Client Setup
Wakapi relies on the open-source [WakaTime](https://github.com/wakatime/wakatime) client tools. In order to collect statistics to Wakapi, you need to set them up.

1. **Set up WakaTime** for your specific IDE or editor. Please refer to the respective [plugin guide](https://wakatime.com/plugins)
2. Make your local WakaTime client talk to Wakapi by **editing your local `~/.wakatime.cfg`** file as follows
```
api_url = https://your.server:someport/api/heartbeat`
api_key = the_api_key_printed_to_the_console_after_starting_the_server`
```

## Customization

### User Accounts
* When starting wakapi for the first time, a default user _**admin**_ with password _**admin**_ is created. The corresponding API key is printed to the console.
* Additional users, at the moment, can be added only via SQL statements on your database, like this:
    * Connect to your database server: `mysql -u yourusername -p -H your.hostname` (alternatively use GUI tools like _MySQL Workbench_)
    * Select your database: `USE yourdatabasename;`
    * Add the new user: `INSERT INTO users (id, password, api_key) VALUES ('your_nickname', MD5('your_password'), '728f084c-85e0-41de-aa2a-b6cc871200c1');` (the latter value should be a random [UUIDv4](https://tools.ietf.org/html/rfc4122), as can be found in your `~/.wakatime.cfg`)

### Aliases
There is an option to add aliases for project names, editors, operating systems and languages. For instance, if you want to map two projects – `myapp-frontend` and `myapp-backend` – two a common project name – `myapp-web` – in your statistics, you can add project aliases.

At the moment, this can only be done via raw database queries. See [_User Accounts_](#user-accounts) section above on how to do such.
For the above example, you would need to add two aliases, like this:

* `INSERT INTO aliases (type, user_id, key, value) VALUES (0, 'your_username', 'myapp-web', 'myapp-frontend')` (analogously for `myapp-backend`)

#### Types
* Project ~  type **0**
* Language ~  type **1**
* Editor ~ type **2**
* OS ~  type **3**

**NOTE:** In order for the aliases to take effect for non-live statistics, you would either have to wait 24 hours for the cache to be invalidated or restart Wakapi.

## Best Practices
It is recommended to use wakapi behind a **reverse proxy**, like [Caddy](https://caddyserver.com) or _nginx_ to enable **TLS encryption** (HTTPS).
However, if you want to expose your wakapi instance to the public anyway, you need to set `listen = 0.0.0.0` in `config.ini`

## Important Note
**This is not an alternative to using WakaTime.** It is just a custom, non-commercial, self-hosted application to collect coding statistics using the already existing editor plugins provided by the WakaTime community. It was created for personal use only and with the purpose of keeping the sovereignity of your own data. However, if you like the official product, **please support the authors and buy an official WakaTime subscription!**

## License
GPL-v3 @ [Ferdinand Mütsch](https://muetsch.io)
