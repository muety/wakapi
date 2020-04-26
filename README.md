# 📈 wakapi
**A minimalist, self-hosted WakaTime-compatible backend for coding statistics**

![Wakapi screenshot](https://anchr.io/i/bxQ69.png)

[![Buy me a coffee](https://www.buymeacoffee.com/assets/img/custom_images/orange_img.png)](https://buymeacoff.ee/n1try)

## Prerequisites
### Server
* Go >= 1.13 (with `$GOPATH` properly set)
* An SQL database (MySQL, Postgres, Sqlite)

### Client
* [WakaTime plugin](https://wakatime.com/plugins) for your editor / IDE

## Usage
* Create an empty database 
* Enable Go module support: `export GO111MODULE=on`
* Get code: `go get github.com/muety/wakapi`
* Go to project root: `cd "$GOPATH/src/github.com/muety/wakapi"`
* Copy `.env.example` to `.env` and set database credentials
* Set target port in `config.ini`
* Build executable: `go build`
* Run server: `./wakapi`
* Edit your local `~/.wakatime.cfg` file
  * `api_url = https://your.server:someport/api/heartbeat`
  * `api_key = the_api_key_printed_to_the_console_after_starting_the_server`
* Open [http://localhost:3000](http://localhost:3000) in your browser

**As an alternative** to building from source or using `go get` you can also download one of the existing [pre-compiled binaries](https://github.com/muety/wakapi/releases).

### Run with Docker
* Edit `docker-compose.yml` file and change passwords for the DB
* Start the application `docker-compose up -d`
* To get the api key look in the logs `docker-compose logs | grep "API key"`
* The application should now be running on `localhost:3000`

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
