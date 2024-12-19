
# 🗞️ CoinTelegraph RSS News Bot 🤖

> Because manually refreshing CoinTelegraph is so 2023... Let the bot do the clicking for you!

## 🌟 What's This Bot's Deal?

This Telegram bot is like that friend who's obsessed with crypto news and has to tell you about EVERYTHING right away - but in a good way! It fetches the latest news from CoinTelegraph's RSS feed and delivers it straight to your Telegram, so you can:

- 📰 Get news the moment they're published (well, within a minute)
- 🏃‍♂️ Never miss another "Bitcoin hits new ATH" moment
- 🛋️ Stay informed while staying lazy
- 📊 Keep up with the market without losing your sanity

## 🚀 Quick Start

### Prerequisites

- Go 1.21+ (because we're modern like that)
- A Telegram Bot Token (ask @BotFather nicely)
- Basic knowledge of what crypto is (kidding, not required)

### 🔧 Installation

1\. Clone this beauty:

```bash

git clone https://github.com/LywwKkA-aD/gocointelegraphrssparser.git

cd gocointelegraphrssparser

```

2\. Set up your environment (the bot needs to know its secrets):

```bash

cp .env.example .env

# Edit .env with your favorite text editor

# Add your Telegram bot token

```

3\. Build and run:

```bash

go build -o bot cmd/bot/main.go

./bot

```

## 🎮 Bot Commands

- `/start` - Subscribe to news (and begin your journey to enlightenment)
- `/stop` - Unsubscribe (but why would you want to? 😢)
- `/help` - Show available commands (in case you forget these three whole commands)

## 🏗️ Project Structure

```

📁 crypto-news-bot/

├── 📂 cmd/bot/           # Where the magic begins

├── 📂 internal/          # The secret sauce

│   ├── 📂 bot/          # Bot's brain

│   ├── 📂 config/       # Bot's memory

│   ├── 📂 models/       # Bot's knowledge

│   ├── 📂 repository/   # Bot's filing cabinet

│   └── 📂 service/      # Bot's muscles

└── 📂 pkg/              # Bot's toolbox

```

## 🤔 Features

- 🚨 Real-time news delivery (okay, within a minute)
- #️⃣ Automatic hashtag generation from categories
- 🔄 Clean message formatting
- 💤 No duplicate news (we hate spam as much as you do)
- 🎯 Only sends the latest news on first run (no spam bombing)

## 🤝 Contributing

Found a bug? Want to add a feature? Have a better joke for this README? Pull requests are welcome!

1\. Fork it

2\. Create your feature branch (`git checkout -b feature/amazing-feature`)

3\. Commit your changes (`git commit -m 'Add some amazing feature'`)

4\. Push to the branch (`git push origin feature/amazing-feature`)

5\. Open a Pull Request

## 📝 License

This project is licensed under the MIT License - see the LICENSE file for details.

## ⚠️ Disclaimer

This bot is not responsible for:

- Your FOMO
- Your trading decisions
- The current market conditions
- The weather
- Your life choices

## 🙏 Acknowledgments

- CoinTelegraph for having an RSS feed
- Telegram for existing
- Coffee for making this possible
- You for reading this far! 🌟

---

Made with ❤️ and probably too much caffeine
