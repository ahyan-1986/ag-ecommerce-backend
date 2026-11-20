package whatsapp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ahyan-1986/ag-ecommerce-backend/internal/domain"
)

type metaClient struct{}

func New() domain.WhatsAppClient {
	return &metaClient{}
}

type templateMessage struct {
	MessagingProduct string         `json:"messaging_product"`
	To               string         `json:"to"`
	Type             string         `json:"type"`
	Template         templateBody   `json:"template"`
}

type templateBody struct {
	Name       string            `json:"name"`
	Language   templateLanguage  `json:"language"`
	Components []templateComponent `json:"components,omitempty"`
}

type templateLanguage struct {
	Code string `json:"code"`
}

type templateComponent struct {
	Type       string            `json:"type"`
	Parameters []templateParam   `json:"parameters,omitempty"`
}

type templateParam struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

func (c *metaClient) SendTemplate(phoneNumberID, accessToken, toPhone, templateName, language string, params []string) error {
	url := fmt.Sprintf("https://graph.facebook.com/v21.0/%s/messages", phoneNumberID)

	components := []templateComponent{}
	if len(params) > 0 {
		pars := make([]templateParam, len(params))
		for i, p := range params {
			pars[i] = templateParam{Type: "text", Text: p}
		}
		components = append(components, templateComponent{
			Type:       "body",
			Parameters: pars,
		})
	}

	msg := templateMessage{
		MessagingProduct: "whatsapp",
		To:               toPhone,
		Type:             "template",
		Template: templateBody{
			Name:       templateName,
			Language:   templateLanguage{Code: language},
			Components: components,
		},
	}

	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("meta API error: status %d", resp.StatusCode)
	}
	return nil
}
