# 📈 wakapi
**A minimalistic, self-hosted WakaTime-compatible backend for coding statistics**

[![Buy me a coffee](https://www.buymeacoffee.com/assets/img/custom_images/orange_img.png)](https://buymeacoff.ee/n1try)

## Prerequisites
* Go >= 1.10 (with `$GOPATH` properly set)
* A MySQL database

## Usage
* Create an empty MySQL database
* Get code: `go get github.com/n1try/wakapi`
* Go to project root: `cd "$GOPATH/src/github.com/n1try/wakapi"`
* Install dependencies: `go get -d ./...`
* Copy `.env.example` to `.env` and set database credentials
* Set target port in `config.ini`
* Build executable: `go build`
* Run server: `./wakapi`
* On your development computers, edit your local `~/.wakatime.cfg` file and add `api_url = https://your.server:someport/api/heartbeat`

**First run** (create user account): When running the server for the very first time, the database gets populated. Afterwards you have to create yourself a user account. Until proper user sign up and login is implemented, this is done via SQL, like this.
* `mysql -u yourusername -p -H your.hostname`
* `USE yourdatabasename;`
* `INSERT INTO users (id, api_key) VALUES ('your_cool_nickname', '728f084c-85e0-41de-aa2a-b6cc871200c1');` (the latter value is your api key from `~/.wakatime.cfg`)

## Best Practices
It is recommended to use wakapi behind a **reverse proxy**, like [Caddy](https://caddyserver.com) or _nginx_ to enable **TLS encryption** (HTTPS).
However, if you want to expose your wakapi instance to the public anyway, you need to set `listen = 0.0.0.0` in `config.ini`

## Todo
* Persisted summaries / aggregations (for performance)
* User sign up and log in
* Additional endpoints for retrieving statistics data
* Enhanced UI
  * Loading spinner
  * Responsiveness
* Dockerize
* Unit tests

## Important Note
**This is not an alternative to using WakaTime.** It is just a custom, non-commercial, self-hosted application to collect coding statistics using the already existing editor plugins provided by the WakaTime community. It was created for personal use only and with the purpose of keeping the sovereignity of your own data. However, if you like the official product, **please support the authors and buy an official WakaTime subscription!**

## License
GPL-v3 @ [Ferdinand Mütsch](https://muetsch.io)