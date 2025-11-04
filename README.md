# go-nland-liveticker

Liveticker for kfv-online.de/home/einsaetze because the RSS feed is not that reliable as wanted.

## About

Scrapes the kfv-online.de/home/einsaetze page and sends the latest news to a Telegram channel.  
Found operations are stored in a SQLite database.

## Docker

Image is available on [Docker Hub](https://hub.docker.com/repository/docker/tbauriedel/go-nland-liveticker/general).

Recommended persistent docker volumes:
- /root/go-nland-liveticker.log
- Database file

Environments:
- LOG_LEVEL: logging level. "debug" or "info" (Default: "debug")
- TELEGRAM_BOT_ID: ID of the telegram bot
- TELEGRAM_CHAT_ID: ID of the telegram chat
- DatabaseDSN: DSN for the sqlite database (Default: "operations.db")

# LICENSE

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.