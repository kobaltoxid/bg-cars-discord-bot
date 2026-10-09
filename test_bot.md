# Bot test runbook

## Automated checks

Run these from the repository root:

```bash
gofmt -w $(rg --files -g '*.go')
go test ./...
go vet ./...
```

The automated tests use HTML fixtures and do not require a Discord token or a live cars.bg request.

## Discord smoke test

1. Set `DISCORD_BOT_TOKEN` in `.env` or the shell environment.
2. Enable the **Message Content Intent** in the Discord Developer Portal and grant the bot View Channel, Send Messages, and Embed Links permissions.
3. Start the bot with `go run .`.
4. In a channel where the bot can view and send messages, run:

   - `!ping` — expects `Pong`.
   - `!help` — shows usage and examples.
   - `!cars BMW 1` — runs a one-page BMW search.
   - `!cars BMW X5 1` — runs a one-page model search.
   - `!cars BMW 0` — returns validation help without starting a search.
   - `!cars UnknownBrand` — returns an unsupported-brand error without starting a search.

5. Confirm that results include a clickable listing link, a price, and an image when cars.bg provides one. Listings without an embeddable image should still render as text or as an embed without an image.
6. Press Ctrl-C and confirm the process exits cleanly.

Searches are bounded to four concurrent jobs, ten pages, ten posted results, and a two-minute timeout. A cars.bg HTTP error should be reported in Discord rather than appearing as a successful empty search.
