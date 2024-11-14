package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"slices"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"gopkg.in/gomail.v2"
)

const (
	ToDoTopicId        = 3
	ReadingListTopicId = 7
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	opts := []bot.Option{
		bot.WithDefaultHandler(handler),
	}

	token := os.Getenv("TELEGRAM_BOT_API_KEY")
	b, err := bot.New(token, opts...)
	if err != nil {
		panic(err)
	}

	b.Start(ctx)
}

func handleToDoMessage(ctx context.Context, b *bot.Bot, message *models.Message) {
	sendEmail(message.Text, "")
	sendReaction(ctx, b, message, "👍")
}

func handleReadingListMessage(ctx context.Context, b *bot.Bot, message *models.Message) {
}

func sendEmail(subject, text string) {
	from := os.Getenv("SENDER_EMAIL")

	// Данные получателя
	to := os.Getenv("THINGS_EMAIL")

	// SMTP-сервер и порт Gmail
	smtpHost := "smtp.yandex.ru"
	smtpPort := 587

	m := gomail.NewMessage()
	m.SetHeader("From", from)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", text)

	pass := os.Getenv("EMAIL_SERVER_PASSWORD")
	d := gomail.NewDialer(smtpHost, smtpPort, from, pass)
	if err := d.DialAndSend(m); err != nil {
		fmt.Println("Ошибка отправки:", err)
		return
	}
	fmt.Println("Письмо успешно отправлено!")
}

func handler(ctx context.Context, b *bot.Bot, update *models.Update) {
	handledTopicIds := []int{ToDoTopicId, ReadingListTopicId}
	topicId := update.Message.MessageThreadID
	found := slices.Contains(handledTopicIds, topicId)
	if !found {
		return
	}

	switch topicId {
	case ToDoTopicId:
		handleToDoMessage(ctx, b, update.Message)
	case ReadingListTopicId:
		handleReadingListMessage(ctx, b, update.Message)
	}

	fmt.Println("Received message " + update.Message.Text)
}

func sendReaction(ctx context.Context, b *bot.Bot, message *models.Message, reaction string) {
	b.SetMessageReaction(ctx, &bot.SetMessageReactionParams{
		ChatID:    message.Chat.ID,
		MessageID: message.ID,
		Reaction: []models.ReactionType{
			{
				Type:              models.ReactionTypeTypeEmoji,
				ReactionTypeEmoji: &models.ReactionTypeEmoji{Emoji: reaction},
			},
		},
	})
}
