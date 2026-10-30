package helper

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func ctxWithAuth(header string) *gin.Context {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("GET", "/", nil)
	if header != "" {
		c.Request.Header.Set("Authorization", header)
	}
	return c
}

func TestVerifyToken(t *testing.T) {
	old := SECRETKEY
	t.Cleanup(func() { SECRETKEY = old })
	SECRETKEY = "test-secret"

	valid, err := GenerateToken("1", "admin")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := VerifyToken(ctxWithAuth("Bearer " + valid)); err != nil {
		t.Errorf("token valid ditolak: %v", err)
	}

	// dulu "Bearer"/"Bearerx" bikin panic index out of range
	for _, h := range []string{"", "Bearer", "Bearer ", "Bearerx", "Basic abc", "Bearer garbage"} {
		if _, err := VerifyToken(ctxWithAuth(h)); err == nil {
			t.Errorf("header %q seharusnya ditolak", h)
		}
	}

	SECRETKEY = ""
	if _, err := VerifyToken(ctxWithAuth("Bearer " + valid)); err == nil {
		t.Error("secret kosong seharusnya fail closed")
	}
	if _, err := GenerateToken("1", "admin"); err == nil {
		t.Error("GenerateToken dengan secret kosong seharusnya error")
	}
}
