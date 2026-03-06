package dbxsetup

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSendBootstrapRequest(t *testing.T) {
	payload := map[string]interface{}{
		"useFoundationOSBinaryCache":  true,
		"useFoundationPupBinaryCache": false,
	}

	t.Run("returns nil on successful bootstrap response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/system/bootstrap" {
				t.Fatalf("unexpected path: %s", r.URL.Path)
			}
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		err := sendBootstrapRequest(server.Client(), server.URL+"/system/bootstrap", payload)
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
	})

	t.Run("returns error on non-200 bootstrap response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("bootstrap failed"))
		}))
		defer server.Close()

		err := sendBootstrapRequest(server.Client(), server.URL+"/system/bootstrap", payload)
		if err == nil {
			t.Fatal("expected non-nil error")
		}
		if !strings.Contains(err.Error(), "failed to bootstrap system") {
			t.Fatalf("expected bootstrap status error, got %v", err)
		}
	})
}
