# Advanced Setup
## Optional: Client-side proxy
Most Wakatime plugins work in a way that, for every heartbeat to send, the plugin calls your local [wakatime-cli](https://github.com/wakatime/wakatime) (a small Python program that is automatically installed when installing a Wakatime plugin) with a few command-line arguments, which is then run as a new process. Inside that process, a heartbeat request is forged and sent to the backend API – Wakapi in this case.

While this is convenient for plugin developers, as they do not have to deal with sending HTTP requests, etc., it comes with a minor drawback. Because the CLI process shuts down after each request, its TCP connection is closed as well. Accordingly, **TCP connections cannot be re-used** and every single heartbeat request is inevitably preceded by the `SYN` + `SYN ACK` + `ACK` sequence for establishing a new TCP connection as well as a handshake for establishing a new TLS session.

While this certainly does not hurt, it is still a bit of overhead. You can avoid that by setting up a local reverse proxy on your machine, that keeps running as a daemon and can therefore keep a continuous connection.

In this example, [Caddy](https://caddyserver.com) is used as an easy-to-set-up webserver / reverse proxy.

1. [Install Caddy](https://caddyserver.com/)
    * When installing manually, don't forget to set up a systemd service to start Caddy on system startup
1. Create a Caddyfile
    ```
    # /etc/caddy/Caddyfile
   
    http://localhost:8070 {
        reverse_proxy * {
            to https://wakapi.dev  # <-- substitute your own Wakapi host here
            header_up Host {http.reverse_proxy.upstream.host}
            header_down -Server
        }
    }
    ```
1. Restart Caddy
1. Verify that you can access [`http://localhost:8070/api/health`](http://localhost:8070/api/health)
1. Update `~/.wakatime.cfg`
    * Set `api_url = http://localhost:8070/api/heartbeat`
1. Done