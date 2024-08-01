package auth

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/redis/go-redis/v9"
	"github.com/shammianand/go-auth/internal/config"
	"github.com/shammianand/go-auth/internal/utils"
)

var (
	keys     map[string]*Key
	keySet   jwk.Set
	keyMutex sync.RWMutex
)

type Key struct {
	PrivateKey *rsa.PrivateKey
	PublicKey  *rsa.PublicKey
	Kid        string
	CreatedAt  time.Time
}

func generateKey() (*Key, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("failed to generate RSA key: %v", err)
	}

	kid := fmt.Sprintf("key-%d", time.Now().Unix())

	return &Key{
		PrivateKey: privateKey,
		PublicKey:  &privateKey.PublicKey,
		Kid:        kid,
		CreatedAt:  time.Now(),
	}, nil
}

func InitializeKeys() error {
	keyMutex.Lock()
	defer keyMutex.Unlock()

	keys = make(map[string]*Key)
	keySet = jwk.NewSet()

	key, err := generateKey()
	if err != nil {
		return fmt.Errorf("failed to generate key: %v", err)
	}

	keys[key.Kid] = key

	jwkKey, err := jwk.New(key.PublicKey)
	if err != nil {
		return fmt.Errorf("failed to create JWK: %v", err)
	}

	if err := jwkKey.Set(jwk.KeyIDKey, key.Kid); err != nil {
		return fmt.Errorf("failed to set key ID: %v", err)
	}

	keySet.Add(jwkKey)

	return nil
}

func CreateJWT(userID uuid.UUID, cache *redis.Client) (string, error) {
	keyMutex.RLock()
	defer keyMutex.RUnlock()

	if len(keys) == 0 {
		return "", fmt.Errorf("no keys available")
	}

	// Use the first key in the map (we only have one for now)
	var key *Key
	for _, k := range keys {
		key = k
		break
	}

	expiration := time.Second * time.Duration(config.TokenExpiry)

	claims := jwt.MapClaims{
		"iss": "github.com/shammianand/go-auth",
		"sub": userID.String(),
		"exp": time.Now().Add(expiration).Unix(),
		"iat": time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = key.Kid

	tokenString, err := token.SignedString(key.PrivateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %v", err)
	}

	// Store token in Redis
	err = cache.Set(context.Background(), fmt.Sprintf("token:%s", userID.String()), tokenString, expiration).Err()
	if err != nil {
		return "", fmt.Errorf("failed to store token in Redis: %v", err)
	}

	return tokenString, nil
}

func RefreshToken(cache *redis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		oldTokenString := getTokenFromRequest(r)
		oldToken, err := validateToken(oldTokenString)
		if err != nil {
			utils.WriteError(w, http.StatusUnauthorized, fmt.Errorf("invalid token"))
			return
		}

		claims, ok := oldToken.Claims.(jwt.MapClaims)
		if !ok {
			utils.WriteError(w, http.StatusUnauthorized, fmt.Errorf("invalid token claims"))
			return
		}

		userID, ok := claims["sub"].(string)
		if !ok {
			utils.WriteError(w, http.StatusUnauthorized, fmt.Errorf("user ID not found in token"))
			return
		}

		// Check if old token exists in Redis
		storedToken, err := cache.Get(context.Background(), fmt.Sprintf("token:%s", userID)).Result()
		if err != nil || storedToken != oldTokenString {
			utils.WriteError(w, http.StatusUnauthorized, fmt.Errorf("token not found or invalid"))
			return
		}

		// Create new token
		userUUID, _ := uuid.Parse(userID)
		newTokenString, err := CreateJWT(userUUID, cache)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("failed to create new token"))
			return
		}

		// Remove old token from Redis
		cache.Del(context.Background(), fmt.Sprintf("token:%s", userID))

		utils.WriteJSON(w, http.StatusOK, map[string]string{"token": newTokenString})
	}
}

// FIXME: add a storage deb to look up the user id
func WithJWTAuth(handlerFunc http.HandlerFunc, cache *redis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenString := getTokenFromRequest(r)
		token, err := validateToken(tokenString)
		if err != nil {
			log.Printf("failed to validate token: %v", err)
			permissionDenied(w)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok || !token.Valid {
			log.Printf("invalid token claims")
			permissionDenied(w)
			return
		}

		userID, ok := claims["sub"].(string)
		if !ok {
			log.Printf("user ID not found in token")
			permissionDenied(w)
			return
		}

		// Check if token exists in Redis
		storedToken, err := cache.Get(context.Background(), fmt.Sprintf("token:%s", userID)).Result()
		if err != nil || storedToken != tokenString {
			log.Printf("token not found in Redis or mismatch")
			permissionDenied(w)
			return
		}

		ctx := context.WithValue(r.Context(), "userID", userID)
		handlerFunc(w, r.WithContext(ctx))
	}
}

func getTokenFromRequest(r *http.Request) string {
	bearerToken := r.Header.Get("Authorization")
	if len(bearerToken) > 7 && bearerToken[:7] == "Bearer " {
		return bearerToken[7:]
	}
	return ""
}

func validateToken(tokenString string) (*jwt.Token, error) {
	keyMutex.RLock()
	defer keyMutex.RUnlock()

	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		kid, ok := t.Header["kid"].(string)
		if !ok {
			return nil, fmt.Errorf("kid header not found")
		}
		key, found := keys[kid]
		if !found {
			return nil, fmt.Errorf("key %v not found", kid)
		}
		return key.PublicKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %v", err)
	}

	return token, nil
}

func permissionDenied(w http.ResponseWriter) {
	utils.WriteError(w, http.StatusForbidden, fmt.Errorf("permission denied"))
}

func GetUserIdFromContext(ctx context.Context) string {
	userID, ok := ctx.Value("userID").(string)
	if !ok {
		return ""
	}
	return userID
}

func JWKSHandler(w http.ResponseWriter, r *http.Request) {
	keyMutex.RLock()
	defer keyMutex.RUnlock()

	jwks, err := json.Marshal(keySet)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jwks)
}
