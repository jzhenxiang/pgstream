package router

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"
	"time"
)

// CSRFConfig holds configuration for CSRF protection middleware.
type CSRFConfig struct {
	// Secret is used to sign CSRF tokens. Required.
	Secret string
	// TokenHeader is the header name to read the CSRF token from.
	// Defaults to "X-CSRF-Token".
	TokenHeader string
	// CookieName is the name of the cookie that stores the CSRF token.
	// Defaults to "csrf_token".
	CookieName string
	// SafeMethods lists HTTP methods that skip CSRF validation.
	SafeMethods []string
}

func (c *CSRFConfig) defaults() *CSRFConfig {
	if c.TokenHeader == "" {
		c.TokenHeader = "X-CSRF-Token"
	}
	if c.CookieName == "" {
		c.CookieName = "csrf_token"
	}
	if len(c.SafeMethods) == 0 {
		c.SafeMethods = []string{http.MethodGet, http.MethodHead, http.MethodOptions}
	}
	return c
}

func (c *CSRFConfig) isSafe(method string) bool {
	for _, m := range c.SafeMethods {
		if strings.EqualFold(m, method) {
			return true
		}
	}
	return false
}

// WithCSRF returns middleware that validates CSRF tokens for state-changing requests.
// If cfg is nil or Secret is empty the middleware is a no-op.
func WithCSRF(cfg *CSRFConfig) func(http.Handler) http.Handler {
	if cfg == nil || cfg.Secret == "" {
		return func(next http.Handler) http.Handler { return next }
	}
	cfg = cfg.defaults()

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if cfg.isSafe(r.Method) {
				token := generateCSRFToken(cfg.Secret)
				http.SetCookie(w, &http.Cookie{
					Name:     cfg.CookieName,
					Value:    token,
					HttpOnly: false,
					SameSite: http.SameSiteStrictMode,
					Path:     "/",
				})
				next.ServeHTTP(w, r)
				return
			}

			cookie, err := r.Cookie(cfg.CookieName)
			if err != nil {
				http.Error(w, "csrf cookie missing", http.StatusForbidden)
				return
			}

			headerToken := r.Header.Get(cfg.TokenHeader)
			if !validateCSRFToken(cfg.Secret, cookie.Value, headerToken) {
				http.Error(w, "csrf token invalid", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func generateCSRFToken(secret string) string {
	nonce := make([]byte, 16)
	_, _ = rand.Read(nonce)
	ts := time.Now().Unix()
	payload := hex.EncodeToString(nonce) + ":" + hex.EncodeToString([]byte{byte(ts)})
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(payload))
	return payload + "." + hex.EncodeToString(mac.Sum(nil))
}

func validateCSRFToken(secret, cookie, header string) bool {
	if cookie == "" || header == "" {
		return false
	}
	if cookie != header {
		return false
	}
	parts := strings.SplitN(cookie, ".", 2)
	if len(parts) != 2 {
		return false
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(parts[0]))
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(parts[1]))
}
