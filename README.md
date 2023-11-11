# np2misk
Spotify の Now Playing を Misskey に投稿するボット。

YudeさんのMastodon用のナウプレをMisskeyで使えるようにしています。

## Setup
* `.env.sample` を参考にして、必要な値を `.env` に設定してください。

* `SPOTIFY_REFRESH_TOKEN` について、np2mast をリモートサーバー等で稼働させる場合の注意点:\
  このソフトウェアでは、ローカル環境において `refresh_token` を取得するよう想定されています。\
  Spotify Web API アプリケーションのコールバック先を `http://localhost:3000` に設定し、一旦ローカル環境で np2mast のバイナリを動かして `refresh_token` を取得し、その値を `.env` に設定してください。\
  このとき、ローカル環境の np2mast においては、`SPOTIFY_REFRESH_TOKEN` 以外の値が設定された `.env` が必要です。

## License
MIT
Copyright (c) 2022 yude

Copy left 2023 CAT5NEKO (Misskeyに差し替えた部分のみ)