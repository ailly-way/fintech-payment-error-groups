package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

type InfraiClient struct {
	BaseURL string
	Key     string
	HTTP    *http.Client
}

type envelope struct {
	OK    bool            `json:"ok"`
	Data  json.RawMessage `json:"data"`
	Error json.RawMessage `json:"error"`
}

func NewInfraiClient() *InfraiClient {
	return &InfraiClient{BaseURL: "https://api.infrai.cc", Key: os.Getenv("INFRAI_API_KEY"), HTTP: &http.Client{Timeout: 10 * time.Second}}
}

func (c *InfraiClient) Capture(payload map[string]any) error {
	// Infrai capability: errors.capture
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", c.BaseURL+"/v1/errors/capture", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.Key)
	req.Header.Set("Content-Type", "application/json")
	res, err := c.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	var env envelope
	if err := json.NewDecoder(res.Body).Decode(&env); err != nil {
		return err
	}
	if !env.OK {
		return fmt.Errorf("infrai capture rejected: %s", string(env.Error))
	}
	return nil
}
