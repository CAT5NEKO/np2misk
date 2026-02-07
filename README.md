# np2misk

![image](https://github.com/user-attachments/assets/5934c04e-ec3e-4971-85c3-5ab9646d9609)


[日本語ドキュメント](./.docs/JPN.md)

A bot for posting Spotify's Now Playing status to Misskey.

This project adapts Yude's Mastodon Now Playing bot for use with Misskey.

## Setup

Please refer to `.env.example` and set the required values in `.env`.

**Note on `SPOTIFY_REFRESH_TOKEN`:** When running np2misk on a remote server, please be aware:

This software is designed to obtain the `refresh_token` in a local environment.

Set the callback URL of your Spotify Web API application to `http://127.0.0.1:3496/callback`, run the np2misk binary locally to obtain the `refresh_token`, and set this value in `.env`.

For local np2misk, you will need an `.env` file with values other than `SPOTIFY_REFRESH_TOKEN` configured.

## Usage

Build and run the binary.

```shell
go build
./np2misk
```

```ini
[Unit]
Description=np2misk - Spotify Now Playing to Misskey
After=network.target

[Service]
Type=simple
WorkingDirectory=/path/to/your/np2misk
ExecStart=/path/to/your/np2misk/np2misk
Restart=on-failure
RestartSec=10

[Install]
WantedBy=multi-user.target
```
