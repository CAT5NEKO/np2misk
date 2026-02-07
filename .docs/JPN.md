# np2misk

Spotify の Now Playing を Misskey に投稿するボット。

YudeさんのMastodon用のナウプレをMisskeyで使えるようにしています。

## Setup

`.env.example` を参考にして、必要な値を `.env` に設定してください。

※`SPOTIFY_REFRESH_TOKEN` について、np2misk をリモートサーバー等で稼働させる場合の注意点:

このソフトウェアでは、ローカル環境において `refresh_token` を取得するよう想定されています。

Spotify Web API アプリケーションのコールバック先を `http://127.0.0.1:3496/callback` に設定し、一旦ローカル環境で
np2misk のバイナリを動かして `refresh_token` を取得し、その値を `.env` に設定してください。

この際に、ローカル環境の np2misk においては、`SPOTIFY_REFRESH_TOKEN` 以外の値が設定された `.env`
が必要です。

## Usage

ビルド

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
