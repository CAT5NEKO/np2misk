# np2misk

Spotify の Now Playing を Misskey に投稿するボット。

YudeさんのMastodon用のナウプレをMisskeyで使えるようにしています。

## Setup

`.env.sample` を参考にして、必要な値を `.env` に設定してください。

※`SPOTIFY_REFRESH_TOKEN` について、np2mast をリモートサーバー等で稼働させる場合の注意点:\

このソフトウェアでは、ローカル環境において `refresh_token` を取得するよう想定されています。

Spotify Web API アプリケーションのコールバック先を `http://localhost:3000` に設定し、一旦ローカル環境で
np2mast のバイナリを動かして `refresh_token` を取得し、その値を `.env` に設定してください。

この際に、ローカル環境の np2misk においては、`SPOTIFY_REFRESH_TOKEN` 以外の値が設定された `.env`
が必要です。

## Usage

まず、`go build` でバイナリをビルドしてください。

JSON形式でポストする都合上、曲がうまく読み取れなかった場合に処理が終了してしまう場合があるので、恒久的に動作させたい場合は以下のスクリプトとsystemdの設定を組み合わせてください。

```shell
#!/bin/bash

cd /path/to/your/np2misk/
 
if [ -x ./np2misk ]; then
    ./np2misk
  else
    echo "Error: ./np2misk not found or not executable."
fi

while true; do
    ./np2misk
    if [ $? -eq 0 ]; then
      echo "Execution completed successfully. Exiting."
      exit 0
        else
          echo "Error occurred. Retrying..."
            sleep 1
    fi
done
```

```dotenv
[Unit]
  Description=Run run_main script repeatedly

[Service]
  Type=simple
  ExecStart=/bin/bash /path/to/your/np2misk/run_main.sh

[Install]
  WantedBy=default.target
```

今回のサンプルではスクリプトが同一ディレクトリに存在している想定ですが、スクリプトファイルは任意の場所に置いてください。

## License

MIT
Copyright (c) 2022 yude

Copy left 2023 CAT5NEKO (Misskeyに差し替えた部分のみ)