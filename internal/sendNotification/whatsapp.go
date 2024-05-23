package whatsapp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

// SendNotification sends a notification to a whatsapp number

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

func SendNotification(message string) error {

	token, token_exists := os.LookupEnv("WHATSAPP_TOKEN")
	number, number_exists := os.LookupEnv("WHATSAPP_NUMBER")

	if !token_exists {
		return fmt.Errorf("WHATSAPP_TOKEN not found")
	}

	if !number_exists {
		return fmt.Errorf("WHATSAPP_NUMBER not found")
	}

	url := "https://graph.facebook.com/v18.0/202325376305196/messages"

	newMessage := Message{
		MessagingProduct: "whatsapp",
		PreviewURL:       false,
		RecipientType:    "individual",
		To:               number,
		Type:             "text",
		Text:             TextType{Body: message},
	}

	// parse the message to json

	msgBody, err := json.Marshal(newMessage)

	if err != nil {
		fmt.Println("Error parsing message to json", err)
		return err
	}

	fmt.Println("Sending message", msgBody)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(msgBody))

	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		fmt.Println("Error sending message", string(body))
		return fmt.Errorf("error sending message")
	}

	fmt.Println("Message sent successfully")

	return nil

}
