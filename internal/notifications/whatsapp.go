package whatsapp

import (
	"auto-update/internal/database"
	"auto-update/internal/database/models"
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"os"
	"sync"
)

// SendNotification sends a notification to a whatsapp number
type Notifications interface {
	SendWhatsappMessages(message string, userId int64) error
}

type NotificationService struct {
	db database.Service
}

type TextType struct {
	Body string `json:"body"`
}

type Message struct {
	MessagingProduct string   `json:"messaging_product"`
	PreviewURL       bool     `json:"preview_url"`
	RecipientType    string   `json:"recipient_type"`
	To               string   `json:"to"`
	Type             string   `json:"type"`
	Text             TextType `json:"text"`
}

func NewNotificationService() *NotificationService {
	return &NotificationService{
		db: database.GetService(),
	}
}

func sendMessage(number string, message string, token string) error {
	newMessage := Message{
		MessagingProduct: "whatsapp",
		PreviewURL:       false,
		RecipientType:    "individual",
		To:               number,
		Type:             "text",
		Text:             TextType{Body: message},
	}

	msgBody, err := json.Marshal(newMessage)

	if err != nil {
		slog.Error("Error parsing message to json", "error", err)
		return err
	}
	url := "https://graph.facebook.com/v18.0/202325376305196/messages"

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(msgBody))

	if err != nil {
		slog.Error("error creating request", "error", err)
		return err
	}

	req.Header.Set("Authorization", "Bearer "+token)
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
			if err := sendMessage(cfg.Number, message, token); err != nil {
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
