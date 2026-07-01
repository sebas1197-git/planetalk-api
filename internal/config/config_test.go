package config

import (
	"testing"
	"time"
)

func TestGetenvFallback(t *testing.T) {
	if got := getenv("PLANETALK_UNSET_XYZ", "def"); got != "def" {
		t.Errorf("fallback = %q, want def", got)
	}
	t.Setenv("PLANETALK_SET_XYZ", "val")
	if got := getenv("PLANETALK_SET_XYZ", "def"); got != "val" {
		t.Errorf("value = %q, want val", got)
	}
}

func TestGetDuration(t *testing.T) {
	if got := getDuration("PLANETALK_DUR_UNSET", time.Minute); got != time.Minute {
		t.Errorf("fallback = %v, want 1m", got)
	}
	t.Setenv("PLANETALK_DUR", "2h")
	if got := getDuration("PLANETALK_DUR", time.Minute); got != 2*time.Hour {
		t.Errorf("parsed = %v, want 2h", got)
	}
	t.Setenv("PLANETALK_DUR_BAD", "not-a-duration")
	if got := getDuration("PLANETALK_DUR_BAD", time.Minute); got != time.Minute {
		t.Errorf("bad value should fall back, got %v", got)
	}
}
