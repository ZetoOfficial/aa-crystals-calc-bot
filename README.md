# AA Crystals Calc Bot

Telegram bot for calculating the cheapest pack combination for a requested amount of ШК.

## Run

Enable inline mode for the bot in BotFather, then run:

```bash
BOT_TOKEN=123:token go run ./cmd/bot
```

Optional environment variables:

```bash
USD_RUB_FALLBACK=80
MAX_SHK=100000
CBR_RATE_URL=https://www.cbr.ru/scripts/XML_daily.asp
METRICS_ADDR=127.0.0.1:8080
```

## Commands

Inline mode:

```text
@bot_name куса 100
```

Direct message or group message:

```text
куса 100
```

The bot returns the required crystals, package combination, total USD, total RUB, and extra crystals.

## Observability

The bot logs one trace line per request:

```text
request_trace kind=inline total_ms=12.345 parse_ms=0.010 rate_ms=0.004 calculate_ms=0.100 format_ms=0.020 send_ms=0.000 answer_ms=12.100
```

Metrics are exposed through Go `expvar`:

```bash
curl http://127.0.0.1:8080/debug/vars
```

Set `METRICS_ADDR=off` to disable the metrics HTTP server.
