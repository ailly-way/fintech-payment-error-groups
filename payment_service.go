package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type PaymentEvent struct {
	ID      string `json:"id"`
	Amount  int64  `json:"amount_cents"`
	Country string `json:"country"`
}

type Decision struct {
	Status string `json:"status"`
	Reason string `json:"reason"`
}

func assessPayment(p PaymentEvent) Decision {
	if p.Amount <= 0 {
		return Decision{"rejected", "invalid_amount"}
	}
	if p.Amount > 100000 {
		return Decision{"review", "high_value"}
	}
	return Decision{"approved", "standard"}
}

func paymentHandler(client *InfraiClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var p PaymentEvent
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}
		d := assessPayment(p)
		if d.Status == "rejected" {
			_ = client.Capture(map[string]any{"title": "payment rejected", "message": d.Reason, "level": "warning", "fingerprint": []string{"payment", d.Reason}, "exception": fmt.Sprintf("payment %s", p.ID), "context": map[string]any{"payment_id": p.ID, "amount_cents": p.Amount, "country": p.Country}})
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(d)
	}
}

func runServer() error {
	client := NewInfraiClient()
	mux := http.NewServeMux()
	mux.HandleFunc("/payments", paymentHandler(client))
	server := &http.Server{Addr: ":8080", Handler: mux, ReadHeaderTimeout: 5 * time.Second}
	return server.ListenAndServe()
}
