package shared

import "testing"

func TestResolveBind(t *testing.T) {
	tests := []struct {
		name       string
		port       int
		envPort    string
		envBind    string
		envEnv     string
		wantHost   string
		wantPort   int
		wantPublic bool
		wantOpen   bool
	}{
		{
			name:     "default is loopback",
			port:     8080,
			wantHost: "localhost", wantPort: 8080, wantPublic: false, wantOpen: true,
		},
		{
			name:    "stray PORT does not publish",
			port:    0,
			envPort: "3000",
			// A dev shell exporting PORT must not expose the server.
			wantHost: "localhost", wantPort: 3000, wantPublic: false, wantOpen: true,
		},
		{
			name:    "PORT with UNUM_ENV binds all interfaces",
			envPort: "8080", envEnv: "production",
			wantHost: "0.0.0.0", wantPort: 8080, wantPublic: true, wantOpen: false,
		},
		{
			name:     "UNUM_BIND is explicit",
			port:     9000,
			envBind:  "0.0.0.0",
			wantHost: "0.0.0.0", wantPort: 9000, wantPublic: true, wantOpen: false,
		},
		{
			name:     "UNUM_BIND loopback stays private",
			port:     9000,
			envBind:  "127.0.0.1",
			wantHost: "127.0.0.1", wantPort: 9000, wantPublic: false, wantOpen: true,
		},
		{
			name:     "invalid PORT falls back to the given port",
			port:     7000,
			envPort:  "not-a-number",
			wantHost: "localhost", wantPort: 7000, wantPublic: false, wantOpen: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("PORT", tt.envPort)
			t.Setenv("UNUM_BIND", tt.envBind)
			t.Setenv("UNUM_ENV", tt.envEnv)

			b, err := ResolveBind(tt.port)
			if err != nil {
				t.Fatalf("ResolveBind: %v", err)
			}
			if b.Host != tt.wantHost {
				t.Errorf("Host=%q, want %q", b.Host, tt.wantHost)
			}
			if b.Port != tt.wantPort {
				t.Errorf("Port=%d, want %d", b.Port, tt.wantPort)
			}
			if b.Public != tt.wantPublic {
				t.Errorf("Public=%v, want %v", b.Public, tt.wantPublic)
			}
			if b.AutoOpen != tt.wantOpen {
				t.Errorf("AutoOpen=%v, want %v", b.AutoOpen, tt.wantOpen)
			}
		})
	}
}

func TestResolveBind_AllocatesFreePort(t *testing.T) {
	t.Setenv("PORT", "")
	t.Setenv("UNUM_BIND", "")
	t.Setenv("UNUM_ENV", "")

	b, err := ResolveBind(0)
	if err != nil {
		t.Fatalf("ResolveBind: %v", err)
	}
	if b.Port == 0 {
		t.Error("Port should be allocated when none is given")
	}
}

func TestBindAddrAndURL(t *testing.T) {
	b := Bind{Host: "localhost", Port: 1234}
	if got := b.Addr(); got != "localhost:1234" {
		t.Errorf("Addr()=%q", got)
	}
	if got := b.URL(); got != "http://localhost:1234" {
		t.Errorf("URL()=%q", got)
	}
}

func TestIsLoopbackHost(t *testing.T) {
	tests := map[string]bool{
		"localhost":   true,
		"127.0.0.1":   true,
		"::1":         true,
		"0.0.0.0":     false,
		"192.168.1.5": false,
		"":            false,
	}
	for host, want := range tests {
		if got := isLoopbackHost(host); got != want {
			t.Errorf("isLoopbackHost(%q)=%v, want %v", host, got, want)
		}
	}
}
