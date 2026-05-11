package sender

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Sender struct {
	apiURL string
	token  string
	client *http.Client
}

func New(apiURL, token string) *Sender {
	return &Sender{
		apiURL: apiURL,
		token:  token,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

type payload struct {
	Token   string      `json:"token"`
	Metrics interface{} `json:"metrics"`
}

func (s *Sender) Send(metrics interface{}) error {
	body, err := json.Marshal(payload{Token: s.token, Metrics: metrics})
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, s.apiURL+"/api/v1/ingest", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.token)

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("server returned %d", resp.StatusCode)
	}
	return nil
}
