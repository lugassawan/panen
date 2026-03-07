package applog

import (
	"bytes"
	"errors"
	"log/slog"
	"strings"
	"testing"
)

func TestShortCaller(t *testing.T) {
	tests := []struct {
		name string
		full string
		want string
	}{
		{
			name: "trims module prefix",
			full: "github.com/lugassawan/panen/backend/usecase.(*PortfolioService).syncHoldingPeak",
			want: "usecase.(*PortfolioService).syncHoldingPeak",
		},
		{
			name: "trims infra prefix",
			full: "github.com/lugassawan/panen/backend/infra/db.(*SQLite).Query",
			want: "infra/db.(*SQLite).Query",
		},
		{
			name: "returns full if no prefix match",
			full: "fmt.Println",
			want: "fmt.Println",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shortCaller(tt.full)
			if got != tt.want {
				t.Errorf("shortCaller(%q) = %q, want %q", tt.full, got, tt.want)
			}
		})
	}
}

func setTestLogger(t *testing.T) *bytes.Buffer {
	t.Helper()
	original := slog.Default()
	t.Cleanup(func() { slog.SetDefault(original) })
	var buf bytes.Buffer
	slog.SetDefault(slog.New(slog.NewTextHandler(&buf, nil)))
	return &buf
}

func TestInfo(t *testing.T) {
	buf := setTestLogger(t)

	Info("test info", Fields{"key": "val"})

	out := buf.String()
	if !strings.Contains(out, "test info") {
		t.Errorf("expected log to contain message, got %q", out)
	}
	if !strings.Contains(out, "caller=infra/applog.TestInfo") {
		t.Errorf("expected caller to point to TestInfo, got %q", out)
	}
	if strings.Contains(out, "err=") {
		t.Errorf("info log should not contain err key, got %q", out)
	}
}

func TestWarn(t *testing.T) {
	buf := setTestLogger(t)

	Warn("test warn", errors.New("boom"), Fields{"key": "val"})

	out := buf.String()
	if !strings.Contains(out, "test warn") {
		t.Errorf("expected log to contain message, got %q", out)
	}
	if !strings.Contains(out, "caller=infra/applog.TestWarn") {
		t.Errorf("expected caller to point to TestWarn, got %q", out)
	}
	if !strings.Contains(out, "err=boom") {
		t.Errorf("expected log to contain err=boom, got %q", out)
	}
}

func TestError(t *testing.T) {
	buf := setTestLogger(t)

	Error("test error", errors.New("fail"), nil)

	out := buf.String()
	if !strings.Contains(out, "test error") {
		t.Errorf("expected log to contain message, got %q", out)
	}
	if !strings.Contains(out, "caller=infra/applog.TestError") {
		t.Errorf("expected caller to point to TestError, got %q", out)
	}
	if !strings.Contains(out, "err=fail") {
		t.Errorf("expected log to contain err=fail, got %q", out)
	}
}
