package middleware

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// AuthMiddleware validates Clerk session JWTs (RS256) using the issuer's JWKS.
//
// Clerk signs session tokens with RS256. The public keys are published at
// {issuer}/.well-known/jwks.json — we fetch them lazily and cache per issuer.
//
// For local development without Clerk configured, set AUTH_DISABLED=true to
// bypass verification (never do this in production).
func AuthMiddleware() gin.HandlerFunc {
	cache := newJWKSCache()
	authDisabled := strings.EqualFold(os.Getenv("AUTH_DISABLED"), "true")

	return func(c *gin.Context) {
		if authDisabled {
			c.Set("user_id", "dev-user")
			c.Set("user_email", "dev@redrose.local")
			c.Next()
			return
		}

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			unauthorized(c, "Authorization header required")
			return
		}
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			unauthorized(c, "Invalid authorization header format")
			return
		}

		claims := jwt.MapClaims{}
		_, err := jwt.ParseWithClaims(parts[1], claims, func(token *jwt.Token) (any, error) {
			if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			iss, _ := claims["iss"].(string)
			kid, _ := token.Header["kid"].(string)
			if iss == "" || kid == "" {
				return nil, fmt.Errorf("token missing iss or kid")
			}
			return cache.key(iss, kid)
		}, jwt.WithValidMethods([]string{"RS256"}))

		if err != nil {
			unauthorized(c, "Invalid or expired token")
			return
		}

		if sub, ok := claims["sub"].(string); ok {
			c.Set("user_id", sub)
		}
		if email, ok := claims["email"].(string); ok {
			c.Set("user_email", email)
		}
		c.Next()
	}
}

func unauthorized(c *gin.Context, msg string) {
	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"success": false, "error": msg})
}

// ── JWKS cache ─────────────────────────────────────────────────────────────

type jwksCache struct {
	mu     sync.RWMutex
	keys   map[string]*rsa.PublicKey // key: issuer|kid
	loaded map[string]time.Time      // issuer -> last fetch
	client *http.Client
}

func newJWKSCache() *jwksCache {
	return &jwksCache{
		keys:   make(map[string]*rsa.PublicKey),
		loaded: make(map[string]time.Time),
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (jc *jwksCache) key(issuer, kid string) (*rsa.PublicKey, error) {
	cacheKey := issuer + "|" + kid

	jc.mu.RLock()
	k := jc.keys[cacheKey]
	last := jc.loaded[issuer]
	jc.mu.RUnlock()
	if k != nil {
		return k, nil
	}

	// Refresh at most once per minute per issuer to avoid hammering the JWKS URL.
	if time.Since(last) < time.Minute && !last.IsZero() {
		return nil, fmt.Errorf("unknown key id %q", kid)
	}
	if err := jc.refresh(issuer); err != nil {
		return nil, err
	}

	jc.mu.RLock()
	k = jc.keys[cacheKey]
	jc.mu.RUnlock()
	if k == nil {
		return nil, fmt.Errorf("unknown key id %q", kid)
	}
	return k, nil
}

type jwk struct {
	Kid string `json:"kid"`
	Kty string `json:"kty"`
	N   string `json:"n"`
	E   string `json:"e"`
}

func (jc *jwksCache) refresh(issuer string) error {
	url := strings.TrimRight(issuer, "/") + "/.well-known/jwks.json"
	resp, err := jc.client.Get(url)
	if err != nil {
		return fmt.Errorf("fetch jwks: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("jwks endpoint returned %d", resp.StatusCode)
	}

	var doc struct {
		Keys []jwk `json:"keys"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&doc); err != nil {
		return fmt.Errorf("decode jwks: %w", err)
	}

	jc.mu.Lock()
	defer jc.mu.Unlock()
	for _, k := range doc.Keys {
		if k.Kty != "RSA" {
			continue
		}
		pub, err := rsaPublicKey(k.N, k.E)
		if err != nil {
			continue
		}
		jc.keys[issuer+"|"+k.Kid] = pub
	}
	jc.loaded[issuer] = time.Now()
	return nil
}

func rsaPublicKey(nStr, eStr string) (*rsa.PublicKey, error) {
	nBytes, err := base64.RawURLEncoding.DecodeString(nStr)
	if err != nil {
		return nil, err
	}
	eBytes, err := base64.RawURLEncoding.DecodeString(eStr)
	if err != nil {
		return nil, err
	}
	e := 0
	for _, b := range eBytes {
		e = e<<8 | int(b)
	}
	return &rsa.PublicKey{N: new(big.Int).SetBytes(nBytes), E: e}, nil
}
