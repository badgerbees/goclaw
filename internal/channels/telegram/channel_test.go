package telegram

import (
	"testing"

)

func TestResolveTelegramAPI(t *testing.T) {
	tests := []struct {
		name       string
		apiServer  string
		wantBase   string
		wantCustom bool
	}{
		{
			name:       "official default",
			apiServer:  "",
			wantBase:   "https://api.telegram.org",
			wantCustom: false,
		},
		{
			name:       "custom server without slash",
			apiServer:  "http://localhost:8081",
			wantBase:   "http://localhost:8081",
			wantCustom: true,
		},
		{
			name:       "custom server with slash",
			apiServer:  "http://localhost:8081/",
			wantBase:   "http://localhost:8081",
			wantCustom: true,
		},
		{
			name:       "missing scheme (defaults to http)",
			apiServer:  "localhost:8081",
			wantBase:   "http://localhost:8081",
			wantCustom: true,
		},
		{
			name:       "pure slash (fallback to default)",
			apiServer:  "/",
			wantBase:   "https://api.telegram.org",
			wantCustom: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotBase, gotCustom := resolveTelegramAPI(tt.apiServer)
			if gotBase != tt.wantBase {
				t.Errorf("resolveTelegramAPI(%q) gotBase = %q, want %q", tt.apiServer, gotBase, tt.wantBase)
			}
			if gotCustom != tt.wantCustom {
				t.Errorf("resolveTelegramAPI(%q) gotCustom = %v, want %v", tt.apiServer, gotCustom, tt.wantCustom)
			}
		})
	}
}

func TestChannel_FieldResolution(t *testing.T) {
	apiBase, isCustom := resolveTelegramAPI("myproxy.com/")
	c := &Channel{
		apiBase:     apiBase,
		isCustomAPI: isCustom,
	}

	if c.apiBase != "http://myproxy.com" {
		t.Errorf("expected apiBase http://myproxy.com, got %q", c.apiBase)
	}
	if !c.isCustomAPI {
		t.Errorf("expected isCustomAPI true")
	}
}
