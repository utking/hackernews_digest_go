package fetcher

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"strconv"
)

// DigestTelegram Telegram data type and its methods
type DigestTelegram struct {
	tgConfig TelegramConfig
}

// Prepare the message to be sent to Telegram
func (telegram *DigestTelegram) prepareMessage(digest *[]DigestItem) string {
	message := ""

	for _, item := range *digest {
		message += item.newsTitle + " - " + item.newsUrl + "\n"
	}

	return message
}

// SendTelegram Prepare and send an Telegram message from the list of the provided news items
func (telegram *DigestTelegram) SendTelegram(digest *[]DigestItem, tgConfig TelegramConfig) {
	bot, err := tgbotapi.NewBotAPI(tgConfig.Token)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true
	log.Printf("Authorized on account %s", bot.Self.UserName)

	chatID, err := strconv.Atoi(tgConfig.ChatId)
	if err != nil {
		log.Panic(err)
	}

	for _, item := range *digest {
		message := fmt.Sprintf("*%s*\n\n[%s](%s)", item.newsTitle, item.newsUrl, item.newsUrl)
		msg := tgbotapi.NewMessage(int64(chatID), message)
		msg.ParseMode = "Markdown"
		_, err = bot.Send(msg)
		if err != nil {
			log.Panic(err)
		}
	}

	log.Printf("Message sent to chat ID %d", chatID)
}
