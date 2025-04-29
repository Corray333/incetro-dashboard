package gsheets

import (
	"log/slog"
	"os"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/sheets/v4"
)

type Client struct {
	svc *sheets.Service
}

func NewSheetsClient() *Client {
	b, err := os.ReadFile("../secrets/credentials.json")
	if err != nil {
		slog.Error("Unable to read client secret file", "error", err)
		panic(err)
	}

	config, err := google.ConfigFromJSON(b, sheets.SpreadsheetsScope)
	if err != nil {
		slog.Error("Unable to parse client secret file to config", "error", err)
		panic(err)
	}
	client := GetClient(config)

	svc, err := sheets.New(client)
	if err != nil {
		slog.Error("Unable to retrieve Sheets Client", "error", err)
		panic(err)
	}
	return &Client{svc: svc}
}

func (s *Client) Svc() *sheets.Service {
	return s.svc
}
