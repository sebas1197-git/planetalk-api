// Package translate turns text from one language into another.
//
// Same pattern as the SMS sender: an interface with two implementations.
//   - StubTranslator: passes text through unchanged (dev / no API key).
//   - GoogleTranslator: calls Google Cloud Translation v2 (REST + API key).
//
// Chat code depends on the interface, so it never changes when we swap them.
package translate

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Result is a translated string plus the detected source language.
type Result struct {
	Text       string
	SourceLang string
}

type Translator interface {
	Translate(ctx context.Context, text, target string) (Result, error)
}

// ---- Dev implementation -------------------------------------------------

// StubTranslator returns the text unchanged. Lets chat work without an API key.
type StubTranslator struct{}

func (StubTranslator) Translate(_ context.Context, text, _ string) (Result, error) {
	return Result{Text: text, SourceLang: ""}, nil
}

// ---- Production implementation ------------------------------------------

// GoogleTranslator uses Google Cloud Translation v2 via a simple REST call.
type GoogleTranslator struct {
	apiKey string
	client *http.Client
}

func (g *GoogleTranslator) Translate(ctx context.Context, text, target string) (Result, error) {
	payload, _ := json.Marshal(map[string]string{"q": text, "target": target, "format": "text"})
	url := "https://translation.googleapis.com/language/translate/v2?key=" + g.apiKey

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return Result{}, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := g.client.Do(req)
	if err != nil {
		return Result{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return Result{}, fmt.Errorf("translate api status %d", resp.StatusCode)
	}

	var out struct {
		Data struct {
			Translations []struct {
				TranslatedText         string `json:"translatedText"`
				DetectedSourceLanguage string `json:"detectedSourceLanguage"`
			} `json:"translations"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return Result{}, err
	}
	if len(out.Data.Translations) == 0 {
		return Result{Text: text}, nil
	}
	t := out.Data.Translations[0]
	return Result{Text: t.TranslatedText, SourceLang: t.DetectedSourceLanguage}, nil
}

// New returns a GoogleTranslator if an API key is set, otherwise the stub.
func New(apiKey string) Translator {
	if apiKey == "" {
		return StubTranslator{}
	}
	return &GoogleTranslator{apiKey: apiKey, client: &http.Client{Timeout: 5 * time.Second}}
}
