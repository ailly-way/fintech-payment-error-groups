package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAssessPayment(t *testing.T) {
	cases := []struct {
		name   string
		amount int64
		want   string
	}{
		{"normal payment", 2500, "approved"}, {"large payment", 150000, "review"}, {"zero amount", 0, "rejected"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := assessPayment(PaymentEvent{Amount: tc.amount}).Status; got != tc.want {
				t.Fatalf("got %s, want %s", got, tc.want)
			}
		})
	}
}

func TestPaymentHandlerRejectsAndCaptures(t *testing.T) {
	captured := false
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		captured = r.Method == "POST" && r.URL.Path == "/v1/errors/capture"
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"ok":true,"data":{},"error":null,"metadata":{}}`))
	}))
	defer api.Close()
	client := &InfraiClient{BaseURL: api.URL, HTTP: api.Client()}
	req := httptest.NewRequest("POST", "/payments", strings.NewReader(`{"id":"p1","amount_cents":0,"country":"US"}`))
	rec := httptest.NewRecorder()
	paymentHandler(client)(rec, req)
	if rec.Code != 200 || !captured {
		t.Fatalf("status=%d captured=%v", rec.Code, captured)
	}
}
