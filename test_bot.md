# Discord Bot Testing Guide

## Setup Steps

### 1. Create Discord Application & Bot
1. Go to [Discord Developer Portal](https://discord.com/developers/applications)
2. Click "New Application" and give it a name
3. Go to "Bot" section in the left sidebar
4. Click "Add Bot"
5. Copy the bot token (keep it secret!)

### 2. Set Environment Variable

#### Windows

**Command Prompt (Temporary - Current Session Only):**
```cmd
set DISCORD_BOT_TOKEN=your_actual_bot_token_here
```

**PowerShell (Temporary - Current Session Only):**
```powershell
$env:DISCORD_BOT_TOKEN="your_actual_bot_token_here"
```

**Permanent System Environment Variable (GUI):**
1. Press `Win + R`, type `sysdm.cpl`, press Enter
2. Click "Environment Variables" button
3. Under "User variables" click "New"
4. Variable name: `DISCORD_BOT_TOKEN`
5. Variable value: `your_actual_bot_token_here`
6. Click OK, OK, OK
7. **Restart your terminal/VS Code**

**Permanent System Environment Variable (Command Line):**
```cmd
setx DISCORD_BOT_TOKEN "your_actual_bot_token_here"
```
*Note: Restart terminal after using `setx`*

#### Linux/Mac

**Temporary (Current Session Only):**
```bash
export DISCORD_BOT_TOKEN=your_actual_bot_token_here
```

**Permanent (Add to ~/.bashrc or ~/.zshrc):**
```bash
echo 'export DISCORD_BOT_TOKEN=your_actual_bot_token_here' >> ~/.bashrc
source ~/.bashrc
```

**For Zsh users:**
```bash
echo 'export DISCORD_BOT_TOKEN=your_actual_bot_token_here' >> ~/.zshrc
source ~/.zshrc
```

### 3. Invite Bot to Your Server
1. In Discord Developer Portal, go to OAuth2 > URL Generator
2. Select scopes: `bot`
3. Select bot permissions (IMPORTANT):
   - **Send Messages** ✅
   - **Read Message History** ✅
   - **View Channels** ✅
   - **Read Messages/View Channels** ✅
   - **Enable message content intent** ✅
   - Use Slash Commands (optional)
4. Copy the generated URL and open it in browser
5. Select your test server and authorize

**Important**: Make sure the bot role has permissions in the specific channel you're testing in!

## Running the Bot

```bash
# Make sure you're in the project directory
cd d:/projects/bg-cars-discord-bot

# Run the bot directly
go run main.go

# Or build and run the executable
go build main.go
./main.exe

# Or simply build without specifying main.go (uses current directory)
go build
./bg-cars-discord-bot.exe
```

**Note**: The bot is now completely self-contained with no external dependencies beyond the standard Go modules.

You should see:
```
Bot is ready! Logged in as: YourBotName#1234
Bot is now running. Press CTRL-C to exit.
```

## Testing Commands

Once the bot is running, test these commands in your Discord server:

### Basic Tests
1. **`!ping`** - Should respond with "Pong! 🏓"
2. **`!help`** - Should show available commands
3. **`!test`** - Should confirm bot permissions are working
4. **`!brands`** - Should show available car brands and models

### Car Search Tests
1. **`!cars`** - Search all cars (max 2 pages)
2. **`!cars bmw`** - Search all BMW cars
3. **`!cars bmw 5series`** - Search BMW 5 Series specifically
>**NOT TESTED**
4. **`!cars audi 3`** - Search Audi cars with max 3 pages
5. **`!cars vw model 5`** - Search VW cars with specific model and page limit
>**NOT TESTED**
### Expected Behavior
- Bot responds immediately with search parameters
- Car results are sent as **rich Discord embeds** in the same channel
- Each car shows:
  - **Car image** (if available from the website)
  - **Title** as clickable embed title linking to full listing
  - **Price** as a clean, formatted field (EUR/BGN properly displayed)
  - **Listing ID** as an additional field
  - **Blue color scheme** for professional appearance
- Limited to 10 results per search (to avoid Discord spam)
- Each car gets its own beautiful embed with image
- Summary and completion messages are included
- Fallback to text messages if embeds fail
- **Smart price parsing** - properly formatted prices like "15,000 BGN" or "7,500 EUR"
- **Currency prioritization** - prefers BGN prices when multiple currencies available
- **Regex-based extraction** - handles complex price formats and conversions

### What to Look For
- Bot appears online in Discord
- Bot responds to commands immediately
- Console shows message logs like:
  ```
  Message received: !ping from YourUsername
  ```