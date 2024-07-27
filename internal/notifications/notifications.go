package notifications

import (
	"auto-update/internal/database"
	"auto-update/internal/database/models"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"sync"
)

type Notifications interface {
	SendWhatsappMessages(message string, userId int64) error
	SendDiscordWebhookMessages(message string, userId int64) error
	SendAllNotifications(message string, userId int64) error
}

type NotificationService struct {
	db database.Service
}

type TextType struct {
	Body string `json:"body"`
}

type DiscordEmbed struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Color       int    `json:"color"`
}

type DiscordMessage struct {
	Content string         `json:"content"`
	Embeds  []DiscordEmbed `json:"embeds"`
}

type WhatsAppMessage struct {
	MessagingProduct string   `json:"messaging_product"`
	PreviewURL       bool     `json:"preview_url"`
	RecipientType    string   `json:"recipient_type"`
	To               string   `json:"to"`
	Type             string   `json:"type"`
	Text             TextType `json:"text"`
}

type MessageType string

const (
	MessageTypeWhatsapp MessageType = "whatsapp"
	MessageTypeDiscord  MessageType = "discord"
)

type SendMessage struct {
	Message string
	Url     string
	Token   string
	Number  string
	Color   string
	Type    MessageType
}

func getIntDiscordEmbedColors(color string) int {
	switch color {
	case "red":
		return 15548997
	case "yellow":
		return 16776960
	case "green":
		return 5763719

	default:
		return 0
	}

}

func NewNotificationService() *NotificationService {
	return &NotificationService{
		db: database.GetService(),
	}
}

func sendMessageRequest(msg SendMessage) error {

	var msgBody []byte
	var err error
	switch msg.Type {
	case "whatsapp":
		newMessage := WhatsAppMessage{
			MessagingProduct: "whatsapp",
			PreviewURL:       false,
			RecipientType:    "individual",
			To:               msg.Number,
			Type:             "text",
			Text:             TextType{Body: msg.Message},
		}

		msgBody, err = json.Marshal(newMessage)

		if err != nil {
			slog.Error("Error parsing message to json", "error", err)
			return err
		}

	case "discord":
		newMessage := DiscordMessage{
			Embeds: []DiscordEmbed{
				{
					Title:       "Update",
					Description: msg.Message,
					Color:       getIntDiscordEmbedColors(msg.Color),
				},
			},
		}

		msgBody, err = json.Marshal(newMessage)

		if err != nil {
			slog.Error("Error parsing message to json", "error", err)
			return err
		}

	}

	req, err := http.NewRequest("POST", msg.Url, bytes.NewBuffer(msgBody))

	if err != nil {
		slog.Error("error creating request", "error", err)
		return err
	}

	if msg.Type == "whatsapp" {
		req.Header.Set("Authorization", "Bearer "+msg.Token)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		slog.Error("error creating response", "error", err)
		return err
	}

	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		slog.Error("Error sending message", string(body), err)
		return errors.New("error sending message")
	}

	slog.Info("Message sent successfully")
	return nil
}

func (ns *NotificationService) SendWhatsappMessage(message string, userId int64) error {

	notificationConfigs, err := ns.db.GetUserNotificationByType(userId, "whatsapp")

	fmt.Println(notificationConfigs, "configs")
	if err != nil {
		slog.Error("error getting whatsapp notification configs", "error", err)
		return err
	}

	if len(notificationConfigs) == 0 {
		slog.Error("whatsapp notifications config not found", "error", err)
		return errors.New("whatsapp notifications config not found")
	}

	token, token_exists := os.LookupEnv("WHATSAPP_TOKEN")

	if !token_exists {
		return errors.New("WHATSAPP_TOKEN not found")
	}

	errChan := make(chan error, len(notificationConfigs))
	var wg sync.WaitGroup

	for _, config := range notificationConfigs {
		wg.Add(1)
		go func(cfg models.NotificationConfig) {
			defer wg.Done()
			if err := sendMessageRequest(SendMessage{
				Type:    MessageTypeWhatsapp,
				Message: message,
				Token:   token,
				Number:  cfg.Number,
				Url:     "https://graph.facebook.com/v18.0/202325376305196/messages",
			}); err != nil {
				errChan <- err
			}
		}(config)
	}

	wg.Wait()
	close(errChan)

	for err := range errChan {
		slog.Error("Errors in", "error", err)
	}

	slog.Info("Notifications sended")
	return nil
}

func (ns *NotificationService) SendDiscordWebhookMessages(message string, userId int64, color string) error {
	notificationConfigs, err := ns.db.GetUserNotificationByType(userId, "discord")

	if err != nil {
		slog.Error("error getting discord notification configs", "error", err)
		return err
	}

	if len(notificationConfigs) == 0 {
		slog.Info("no discord notifications config found")
		return errors.New("discord notifications config not found")
	}

	errChan := make(chan error, len(notificationConfigs))
	var wg sync.WaitGroup

	for _, config := range notificationConfigs {
		wg.Add(1)
		go func(cfg models.NotificationConfig) {
			defer wg.Done()
			if err := sendMessageRequest(SendMessage{
				Type:    MessageTypeDiscord,
				Message: message,
				Url:     cfg.Url,
				Color:   color,
			}); err != nil {
				errChan <- err
			}
		}(config)
	}

	slog.Info("Discord notifications sended")
	return nil
}

func (ns *NotificationService) SendAllNotifications(message string, userId int64, color string) error {

	err := ns.SendWhatsappMessage(message, userId)

	if err != nil {
		slog.Error("error sending whatsapps messages", "error", err)
		return err
	}

	err = ns.SendDiscordWebhookMessages(message, userId, color)

	if err != nil {
		slog.Error("error sending discord messages", "error", err)
		return err
	}

	return nil
}
