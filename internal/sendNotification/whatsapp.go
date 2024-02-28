package whatsapp

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
)

var token = os.Getenv("WAB_TOKEN")

// SendNotification sends a notification to a whatsapp number

func SendNotification(message string) error {

	fmt.Println("token", token)
	url := "https://graph.facebook.com/v18.0/202325376305196/messages"

	/*  messaging_product: 'whatsapp',
	    preview_url: false,
	    recipient_type: 'individual',
	    to: contact,
	    type: 'text',
	    text: {
	      body: msg,
	    }, */

	msgBody := fmt.Sprintf(`{messaging_product: 'whatsapp', preview_url: false, recipient_type: 'individual', to: %s, type: 'text', text: {body: '%s'}}`, "554299488471", message)

	fmt.Println("Sending message", msgBody)
	data := []byte(msgBody)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))

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
	}

	fmt.Println("Message sent successfully", string(body))

	return nil

}
