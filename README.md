# 🚗 Bulgarian Cars Discord Bot

A Discord bot that searches for cars on cars.bg and displays them with rich embeds and images.

## 🚀 Quick Start

### 1. Setup
```bash
# Clone the repository
git clone <your-repo-url>
cd bg-cars-discord-bot

# Install dependencies
go mod tidy
```

### 2. Configure Discord Bot
1. Create a Discord application at https://discord.com/developers/applications
2. Create a bot and copy the token
3. Create a `.env` file:
```env
DISCORD_BOT_TOKEN=your_bot_token_here
```

### 3. Run the Bot
```bash
go run main.go
```

## 🤖 Bot Commands

| Command | Description | Example |
|---------|-------------|---------|
| `!cars` | Search all cars | `!cars` |
| `!cars BMW` | Search by brand | `!cars BMW` |
| `!cars BMW X5` | Search brand + model | `!cars BMW X5` |
| `!cars BMW X5 5` | Search with page limit | `!cars BMW X5 5` |
| `!help` | Show help message | `!help` |
| `!ping` | Test bot response | `!ping` |

## 📁 Project Structure

```
bg-cars-discord-bot/
├── main.go                 # 🎯 Entry point (35 lines)
├── pkg/
│   ├── bot/               # 🤖 Discord bot management
│   │   └── bot.go         # Connection, events, commands
│   ├── commands/          # ⚡ Bot commands
│   │   └── cars.go        # Car search command
│   ├── discord/           # 💬 Discord utilities
│   │   └── embeds.go      # Rich message formatting
│   └── scraper/           # 🕷️ Web scraping
│       └── scraper.go     # Cars.bg scraper
├── .env                   # 🔐 Bot token (create this)
└── go.mod                 # 📦 Dependencies
```

## 🛠️ How It Works

### Simple Flow
1. **User types** `!cars BMW X5` in Discord
2. **Bot receives** the message
3. **Scraper searches** cars.bg website
4. **Bot sends back** car listings with images

### Code Flow
```
Discord Message → bot.go → cars.go → scraper.go → embeds.go → Discord Response
```

## 🔧 Adding New Features

### Add a New Command
1. Create function in `pkg/commands/`
2. Add route in `pkg/bot/bot.go` (line 95)
3. Update help message

### Customize Car Display
- Edit `pkg/discord/embeds.go`
- Modify `CreateCarEmbed()` function

### Change Search Logic
- Edit `pkg/scraper/scraper.go`
- Modify `SearchCars()` function

## 🐛 Troubleshooting

**Bot not responding?**
- Check your bot token in `.env`
- Make sure bot has message permissions
- Check console for error messages

**Build errors?**
```bash
go mod tidy
go mod vendor
go build
```

**No cars found?**
- Try broader search terms
- Check if cars.bg is accessible
- Increase page limit: `!cars BMW 5`

## 📝 Development

**Build the project:**
```bash
go build
```

**Run with live reload:**
```bash
# Install air first: go install github.com/cosmtrek/air@latest
air
```

**Test individual packages:**
```bash
go test ./pkg/...
```

## 🎯 Features

- ✅ Rich Discord embeds with car images
- ✅ Bulgarian price formatting (BGN/EUR)
- ✅ Async search (doesn't block Discord)
- ✅ Error handling and fallbacks
- ✅ Clean, modular code structure
- ✅ Easy to extend with new commands

## 📄 License

This project is licensed under the GNU General Public License v3.0 - see the [LICENSE](LICENSE) file for details.