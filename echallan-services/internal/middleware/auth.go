package middleware

import (
    "crypto/rsa"
    "encoding/base64"
    "encoding/json"
    "fmt"
    "math/big"
    "net/http"
    "strings"
    "sync"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/golang-jwt/jwt/v5"
)

type KeycloakConfig struct {
    URL      string
    Realm    string
    ClientID string
}

type jwksCache struct {
    keys      map[string]*rsa.PublicKey
    mu        sync.RWMutex
    expiresAt time.Time
}

var cache = &jwksCache{keys: make(map[string]*rsa.PublicKey)}

func KeycloakAuthMiddleware(cfg KeycloakConfig) gin.HandlerFunc {
    return func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Missing or invalid Authorization header"})
            return
        }
        tokenString := strings.TrimPrefix(authHeader, "Bearer ")

        token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
            if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
                return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
            }
            kid, ok := token.Header["kid"].(string)
            if !ok {
                return nil, fmt.Errorf("missing kid in token header")
            }
            return getPublicKey(cfg, kid)
        }, jwt.WithValidMethods([]string{"RS256"}))

        if err != nil || !token.Valid {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
            return
        }

        claims, ok := token.Claims.(jwt.MapClaims)
        if !ok {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
            return
        }

        // Extract and set user info in context
        c.Set("userID", claims["sub"])
        if realmAccess, ok := claims["realm_access"].(map[string]interface{}); ok {
            if roles, ok := realmAccess["roles"].([]interface{}); ok {
                var roleStrings []string
                for _, r := range roles {
                    if s, ok := r.(string); ok {
                        roleStrings = append(roleStrings, s)
                    }
                }
                c.Set("userRoles", roleStrings)
            }
        }

        c.Next()
    }
}

func RequireRole(role string) gin.HandlerFunc {
    return func(c *gin.Context) {
        roles, exists := c.Get("userRoles")
        if !exists {
            c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "No roles found"})
            return
        }
        roleList, ok := roles.([]string)
        if !ok {
            c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Invalid roles"})
            return
        }
        for _, r := range roleList {
            if r == role {
                c.Next()
                return
            }
        }
        c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": fmt.Sprintf("Required role '%s' not found", role)})
    }
}

func getPublicKey(cfg KeycloakConfig, kid string) (*rsa.PublicKey, error) {
    cache.mu.RLock()
    if key, ok := cache.keys[kid]; ok && time.Now().Before(cache.expiresAt) {
        cache.mu.RUnlock()
        return key, nil
    }
    cache.mu.RUnlock()

    jwksURL := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/certs", cfg.URL, cfg.Realm)
    resp, err := http.Get(jwksURL)
    if err != nil {
        return nil, fmt.Errorf("failed to fetch JWKS: %w", err)
    }
    defer resp.Body.Close()

    var jwks struct {
        Keys []struct {
            Kid string `json:"kid"`
            N   string `json:"n"`
            E   string `json:"e"`
        } `json:"keys"`
    }
    if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
        return nil, fmt.Errorf("failed to decode JWKS: %w", err)
    }

    cache.mu.Lock()
    defer cache.mu.Unlock()
    cache.expiresAt = time.Now().Add(1 * time.Hour)

    for _, k := range jwks.Keys {
        nBytes, err := base64.RawURLEncoding.DecodeString(k.N)
        if err != nil { continue }
        eBytes, err := base64.RawURLEncoding.DecodeString(k.E)
        if err != nil { continue }
        e := 0
        for _, b := range eBytes { e = e*256 + int(b) }
        pubKey := &rsa.PublicKey{N: new(big.Int).SetBytes(nBytes), E: e}
        cache.keys[k.Kid] = pubKey
    }

    if key, ok := cache.keys[kid]; ok {
        return key, nil
    }
    return nil, fmt.Errorf("key with kid '%s' not found in JWKS", kid)
}
