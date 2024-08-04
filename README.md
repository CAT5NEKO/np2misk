# np2misk

![image](https://github.com/user-attachments/assets/5934c04e-ec3e-4971-85c3-5ab9646d9609)


[日本語ドキュメント](./.docs/JPN.md)

A bot for posting Spotify's Now Playing status to Misskey.

This project adapts Yude's Mastodon Now Playing bot for use with Misskey.

## Setup

Please refer to `.env.sample` and set the required values in `.env`.

**Note on `SPOTIFY_REFRESH_TOKEN`:** When running np2misk on a remote server, please be aware:\

This software is designed to obtain the `refresh_token` in a local environment.

Set the callback URL of your Spotify Web API application to `http://localhost:3000`, run the np2misk binary locally to obtain the `refresh_token`, and set this value in `.env`.

For local np2misk, you will need an `.env` file with values other than `SPOTIFY_REFRESH_TOKEN` configured.

## Usage

First, build the binary using `go build`.

Due to the JSON format of posts, the process may terminate if the song is not read correctly. To ensure continuous operation, combine the following script with a systemd configuration.

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

In this example, it is assumed that the script is located in the same directory, but you can place the script file in any location.
