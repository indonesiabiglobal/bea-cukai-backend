package routes

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

// Setiap route selain yang di-allowlist wajib ditolak 401 tanpa token.
// Route baru yang lupa dipasang middleware.Authentication() akan gagal di sini.
func TestAllRoutesRequireAuthExceptPublic(t *testing.T) {
	gin.SetMode(gin.TestMode)
	app := NewRoute(nil, nil) // middleware menolak sebelum handler menyentuh DB

	public := map[string]bool{
		"POST /auth/login": true,
		"OPTIONS /*any":    true,
	}

	for _, r := range app.Routes() {
		key := r.Method + " " + r.Path
		w := httptest.NewRecorder()
		app.ServeHTTP(w, httptest.NewRequest(r.Method, r.Path, nil))
		// body dicek agar 401 benar-benar dari middleware, bukan dari handler
		blocked := w.Code == http.StatusUnauthorized && strings.Contains(w.Body.String(), "Invalid token")
		if public[key] && blocked {
			t.Errorf("%s: route publik ikut terkunci auth", key)
		}
		if !public[key] && !blocked {
			t.Errorf("%s: want 401 Invalid token tanpa token, got %d", key, w.Code)
		}
	}
}
