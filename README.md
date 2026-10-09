# Bulgarian Cars Discord Bot

A small Discord bot for searching current car listings on [cars.bg](https://www.cars.bg/) and posting useful results as rich embeds.

## What it does

- Searches all listings or a supported brand/model.
- Follows cars.bg pagination and stops when a page is empty.
- Extracts listing IDs, titles, prices, links, and images from current card markup, with fallbacks for older markup variants.
- Normalizes EUR, BGN, `€`, `лв`, and common thousands separators.
- Removes duplicate listings and rejects failed HTTP responses instead of silently returning partial results.
- Limits each search to 10 pages and posts at most 10 results to keep channels readable.
- Runs searches with a two-minute timeout and allows at most four searches at once.
- Uses a text fallback when Discord cannot send an embed.

## Quick start

Requirements: Go 1.23 or newer and a Discord application with a bot user.

1. Create a bot in the [Discord Developer Portal](https://discord.com/developers/applications).
2. Enable the **Message Content Intent** under the bot's privileged gateway intents.
3. Copy `.env.example` to `.env` and set the token:

   ```env
   DISCORD_BOT_TOKEN=your_bot_token_here
   ```

4. Invite the bot with permission to view channels, read messages, send messages, and embed links.
5. Start it:

   ```bash
   go mod download
   go run .
   ```

The token is read from the environment. A local `.env` file is supported for development and is ignored by Git.

## Commands

| Command | Description |
| --- | --- |
| `!cars` | Search all brands and models. |
| `!cars BMW` | Search one brand. |
| `!cars BMW X5` | Search a brand and model. |
| `!cars BMW 5` | Search a brand across five pages. |
| `!cars BMW X5 5` | Search a brand/model across five pages. |
| `!help` | Show command help. |
| `!ping` | Check that the bot is responding. |

Pages must be a whole number from 1 to 10. Model filters are applied for model IDs known by the scraper; an unknown model safely falls back to the selected brand rather than inventing a filter.

The scraper currently knows these cars.bg brand filters: BMW, Audi, VW/Volkswagen, Mercedes-Benz, Toyota, Mitsubishi, Honda, Ford, Opel, Renault, Mazda, Citroën, Peugeot, Nissan, Škoda/Skoda, Fiat, Hyundai, Kia, Volvo, and Suzuki. An unsupported non-empty brand is rejected so it cannot accidentally return all cars.

## Development

```bash
gofmt -w $(rg --files -g '*.go')
go test ./...
go vet ./...
```

The repository includes fixture-based parser tests and unit tests for command parsing, embeds, and bot lifecycle behavior. Tests do not call Discord or cars.bg.

## Project layout

```text
.
├── main.go                 # Configuration and process entry point
├── pkg/
│   ├── bot/                # Discord session, intents, routing, shutdown
│   ├── commands/           # !cars parsing and bounded search jobs
│   ├── discord/            # Embed and text rendering
│   └── scraper/            # cars.bg URL construction, HTTP, and HTML parsing
├── .env.example
└── .github/workflows/ci.yml
```

## Scraping notes

The scraper targets public cars.bg listing pages and expects normal HTML responses. It recognizes the site's `/offer/<id>` links, card/data attributes, price fields, images, and relative URLs. The site can change its markup at any time, so parser fixtures should be updated when the upstream structure changes. Keep request volume reasonable and do not commit scraped data or credentials.

## License

GPL-3.0-only. See [LICENSE](LICENSE).
