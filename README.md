# wakapi

## Usage
* Create an empty MySQL database
* Clone repository
* Copy `.env.example` to `.env` and set database credentials
* Install dependencies: `go get -d ./...`
* Set target port in `config.ini`
* Build executable: `go build`
* Run server: `./wakapi`
* Edit your local `~/.wakatime.cfg` file and add `api_url = https://your.server:someport/api/heartbeat`

**First run** (create user account): When running the server for the very first time, the database gets populated. Afterwards you have to create yourself a user account. Until proper user sign up and login is implemented, this is done via SQL, like this.
* `mysql -u yourusername -p -H your.hostname`
* `USE yourdatabasename;`
* `INSERT INTO users (id, api_key) VALUES ('your_cool_nickname', '728f084c-85e0-41de-aa2a-b6cc871200c1');` (the latter value is your api key from `~/.wakatime.cfg`)

## Todo
* Persisted summaries / aggregations (for performance)
* User sign up and log in
* Additional endpoints for retrieving statistics data
* Dockerize
* Unit tests

## Important note
**This is not an alternative to using WakaTime.** It is just a custom, non-commercial, self-hosted application to collect coding statistics using the already existing editor plugins provided by the WakaTime community. It was created for personal use only and with the purpose of keeping the sovereignity of your own data. However, if you like the official product, **please support the authors and buy an official WakaTime subscription!**

## License
GPL-v3 @ [Ferdinand Mütsch](https://muetsch.io)