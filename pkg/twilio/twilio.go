// Package twilio sends SMS messages (used to deliver OTP codes).
//
// Design idea: we define a small INTERFACE (Sender) instead of calling Twilio
// directly everywhere. That way:
//   - in dev, we use ConsoleSender, which just prints the code to the logs
//     (no Twilio account or money needed),
//   - in production, we use TwilioSender, which really sends the SMS.
//
// The auth code depends on the interface, so it never changes when we swap them.
package twilio

import (
	"context"
	"log"
)

// Sender is anything that can send an SMS. The auth service depends on THIS,
// not on a concrete Twilio client.
type Sender interface {
	SendSMS(ctx context.Context, to, body string) error
}

// ---- Dev implementation -------------------------------------------------

// ConsoleSender just logs the message instead of sending a real SMS.
// Perfect for development: you read the OTP code straight from the app logs.
type ConsoleSender struct{}

func (ConsoleSender) SendSMS(ctx context.Context, to, body string) error {
	log.Printf("[SMS stub] to=%s body=%q", to, body)
	return nil
}

// ---- Production implementation ------------------------------------------

// TwilioSender sends real SMS via Twilio's API.
type TwilioSender struct {
	AccountSID string
	AuthToken  string
	From       string
	// TODO: hold a *twilio.RestClient here once you add the twilio-go library.
}

// NewTwilioSender builds a real Twilio sender from your credentials.
// TODO: construct the twilio-go REST client and store it on the struct.
func NewTwilioSender(accountSID, authToken, from string) *TwilioSender {
	return &TwilioSender{AccountSID: accountSID, AuthToken: authToken, From: from}
}

func (s *TwilioSender) SendSMS(ctx context.Context, to, body string) error {
	// TODO: use github.com/twilio/twilio-go to send the message:
	//   params := &openapi.CreateMessageParams{}
	//   params.SetTo(to); params.SetFrom(s.From); params.SetBody(body)
	//   _, err := client.Api.CreateMessage(params)
	//   return err
	return nil
}

// New picks the right Sender based on whether Twilio credentials are present.
// Empty credentials (dev) -> ConsoleSender. Filled (prod) -> TwilioSender.
// TODO:
//
//	if accountSID == "" || authToken == "" { return ConsoleSender{} }
//	return NewTwilioSender(accountSID, authToken, from)
func New(accountSID, authToken, from string) Sender {
	if accountSID == "" || authToken == "" {
		return ConsoleSender{} // dev: print code to logs
	}
	return NewTwilioSender(accountSID, authToken, from)
}
