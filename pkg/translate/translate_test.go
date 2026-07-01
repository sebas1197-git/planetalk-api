package translate

import (
	"context"
	"testing"
)

func TestStubPassthrough(t *testing.T) {
	res, err := StubTranslator{}.Translate(context.Background(), "hello", "es")
	if err != nil {
		t.Fatal(err)
	}
	if res.Text != "hello" {
		t.Errorf("stub should pass text through, got %q", res.Text)
	}
}

func TestNewReturnsStubWithoutKey(t *testing.T) {
	if _, ok := New("").(StubTranslator); !ok {
		t.Error("New(\"\") should return a StubTranslator")
	}
}

func TestNewReturnsGoogleWithKey(t *testing.T) {
	if _, ok := New("some-key").(*GoogleTranslator); !ok {
		t.Error("New(key) should return a *GoogleTranslator")
	}
}
