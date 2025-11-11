package mail

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

const resendAPIEndpoint = "https://api.resend.com/emails"

type ResendRequest struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	Html    string   `json:"html"`
	// ReplyTo je důležité, aby se dalo odpovědět přímo na email zákazníka
	ReplyTo string `json:"reply_to,omitempty"`
}

func SendMailResend(name string, email string, message string) error {

	apiKey := os.Getenv("RESEND_API_KEY")

	recipientEmail := os.Getenv("RECIPIENT_EMAIL")

	if apiKey == "" || recipientEmail == "" {
		return fmt.Errorf("chybí proměnná prostředí RESEND_API_KEY nebo RECIPIENT_EMAIL")
	}

	requestBody := ResendRequest{
		// Důležité: MUSÍ být z ověřené domény (odesílatel, který se ukáže klientovi)
		From:    "Message from <onboarding@koridev.com>",
		To:      []string{recipientEmail}, // Odesílá se na váš email
		Subject: fmt.Sprintf("New Message : %s", name),
		Html:    fmt.Sprintf("<h2>New Message</h2><p><strong>Name:</strong> %s</p><p><strong>Email:</strong> %s</p><p><strong>Message:</strong></p><p>%s</p>", name, email, message),
		ReplyTo: email, // Klíčové pro možnost odpovědi přímo v poště
	}

	///Pars do JSON formatu
	jsonBody, err := json.Marshal(requestBody)

	if err != nil {
		return fmt.Errorf("chyba při serializaci JSON: %w", err)
	}

	//Vytvoření a konfigurace HTTP požadavku
	req, err := http.NewRequest("POST", resendAPIEndpoint, bytes.NewBuffer(jsonBody))

	if err != nil {
		return fmt.Errorf("chyba při tvorbě požadavku: %w", err)
	}
	////Nastavi hlavicky pro HTTP request
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	///Odeslání požadavku
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("chyba sítě při volání Resend API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		// Logování chyby z Resend API pro diagnostiku
		var resendError map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&resendError)
		return fmt.Errorf("chyba Resend API, status: %d, detail: %v", resp.StatusCode, resendError)
	}

	return nil
}
