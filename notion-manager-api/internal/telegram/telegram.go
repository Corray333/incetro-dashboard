package telegram

import (
	"github.com/PaulSonOfLars/gotgbot/v2"
)

type TelegramClient struct {
	bot *gotgbot.Bot
}

func (t *TelegramClient) GetBot() *gotgbot.Bot {
	return t.bot
}

func NewTelegramClient(token string) *TelegramClient {
	bot, err := gotgbot.NewBot(token, nil)
	if err != nil {
		panic("failed to create bot", "error", err)
	}

	tr := &TelegramClient{
		bot: bot,
	}

	return tr
}
