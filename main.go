package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"slices"
	"strings"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/joho/godotenv"
	"golang.org/x/net/html"
	"gopkg.in/gomail.v2"
)

const (
	ToDoTopicId        = 3
	ReadingListTopicId = 229
)

var logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))

func init() {
	if err := godotenv.Load(); err != nil {
		logger.Error("No .env file found")
	}
}

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
	var title string
	var err error

	if isURL(message.Text) {
		title, err = fetchTitle(message.Text)
		logger.Info("link", "title", title)
		if title == "" || err != nil {
			title = "Посмотреть"
		}
		fmt.Println(title)
	} else {
		title = "Прочитать"
	}

	sendEmail(title, message.Text)
	sendReaction(ctx, b, message, "👍")
}

func isURL(str string) bool {
	u, err := url.Parse(str)
	return err == nil && u.Scheme != "" && u.Host != ""
}

func fetchTitle(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	z := html.NewTokenizer(resp.Body)
	for {
		tt := z.Next()

		switch tt {
		case html.ErrorToken:
			return "", errors.New("Заголовок не найден")
		case html.StartTagToken:
			t := z.Token()
			if t.Data == "title" {
				z.Next()
				return strings.TrimSpace(z.Token().Data), nil
			}
		}
	}
}

func sendEmail(subject, text string) {
	smtpHost := "smtp.yandex.ru"
	smtpPort := 587

	from := os.Getenv("SENDER_EMAIL")
	to := os.Getenv("THINGS_EMAIL")

	m := gomail.NewMessage()
	m.SetHeader("From", from)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", text)

	pass := os.Getenv("EMAIL_SERVER_PASSWORD")
	d := gomail.NewDialer(smtpHost, smtpPort, from, pass)
	if err := d.DialAndSend(m); err != nil {
		logger.Error("Ошибка отправки:", "error", err.Error())
		return
	}
	logger.Info("Письмо успешно отправлено!")
}

func handler(ctx context.Context, b *bot.Bot, update *models.Update) {
	handledTopicIds := []int{ToDoTopicId, ReadingListTopicId}
	topicId := update.Message.MessageThreadID

	found := slices.Contains(handledTopicIds, topicId)
	if !found || update.Message.Text == "" {
		return
	}

	logger.Info("Received message", "topicId", topicId, "message", update.Message.Text)

	switch topicId {
	case ToDoTopicId:
		handleToDoMessage(ctx, b, update.Message)
	case ReadingListTopicId:
		handleReadingListMessage(ctx, b, update.Message)
	}
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

func DeleteMessage(ctx context.Context, b *bot.Bot, message *models.Message, reaction string) {
	b.DeleteMessage(ctx, &bot.DeleteMessageParams{ChatID: message.Chat.ID, MessageID: message.ID})
}
