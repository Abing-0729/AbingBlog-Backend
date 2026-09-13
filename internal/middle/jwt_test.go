package middle

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func makeToken(t *testing.T, secret string, method jwt.SigningMethod, expiresAt time.Time) string {
	t.Helper()
	claims := jwt.RegisteredClaims{Subject: "admin", ExpiresAt: jwt.NewNumericDate(expiresAt)}
	token, err := jwt.NewWithClaims(method, claims).SignedString([]byte(secret))
	if err != nil {
		t.Fatal(err)
	}
	return token
}

func runMiddlewareRequest(t *testing.T, authorization string) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/admin", JWTAuth("secret"), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"subject": c.MustGet(JWTSubjectKey)})
	})
	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	if authorization != "" {
		req.Header.Set("Authorization", authorization)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestJWTAuthRejectsMalformedBearerHeader(t *testing.T) {
	for _, header := range []string{"", "Bearer", "BearerX token", "Basic token", "Bearer "} {
		w := runMiddlewareRequest(t, header)
		if w.Code != http.StatusUnauthorized {
			t.Errorf("header %q: expected 401, got %d", header, w.Code)
		}
	}
}

func TestJWTAuthAcceptsHS256AndStoresSubject(t *testing.T) {
	token := makeToken(t, "secret", jwt.SigningMethodHS256, time.Now().Add(time.Minute))
	w := runMiddlewareRequest(t, "Bearer "+token)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	if w.Body.String() != `{"subject":"admin"}` {
		t.Fatalf("unexpected response: %s", w.Body.String())
	}
}

func TestJWTAuthRejectsExpiredToken(t *testing.T) {
	token := makeToken(t, "secret", jwt.SigningMethodHS256, time.Now().Add(-time.Minute))
	w := runMiddlewareRequest(t, "Bearer "+token)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestJWTAuthRejectsNonHS256Token(t *testing.T) {
	token := makeToken(t, "secret", jwt.SigningMethodHS384, time.Now().Add(time.Minute))
	w := runMiddlewareRequest(t, "Bearer "+token)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestJWTAuthRejectsTokenWithoutSubject(t *testing.T) {
	claims := jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute))}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte("secret"))
	if err != nil {
		t.Fatal(err)
	}
	w := runMiddlewareRequest(t, "Bearer "+token)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}
