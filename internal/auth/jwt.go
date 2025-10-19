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

const (
	keyPrefix      = "auth:key:"
	keySetKey      = "auth:keyset"
	tokenPrefix    = "auth:token:"
	jwksPrefix     = "auth:jwks"
	keyExpiryDays  = 30 // NOTE: adjust as needed
	rsaKeyBits     = 2048
	tokenCacheTime = time.Minute * 60 // NOTE: cache tokens for 1 hour
)

var (
	// NOTE: later on think of an alternative to allow for high
	// throughput request handling
	keyMutex sync.RWMutex
)

type Key struct {
	PrivateKey *rsa.PrivateKey
	PublicKey  *rsa.PublicKey
	Kid        string
	CreatedAt  time.Time
}

func InitializeKeys(cache *redis.Client) error {
	return loadOrGenerateKeys(cache)
}

func loadOrGenerateKeys(cache *redis.Client) error {
	keyMutex.Lock()
	defer keyMutex.Unlock()
	keysJSON, err := cache.Get(context.Background(), keySetKey).Result()
	if err == nil {
		var storedKeys map[string]*Key
		if err := json.Unmarshal([]byte(keysJSON), &storedKeys); err == nil {
			return nil
		}
	}

	utils.Logger.Info("NO KEYS IN REDIS SO GENERATING AN RSA KEY PAIR")
	key, err := generateKey()
	if err != nil {
		return fmt.Errorf("failed to generate key: %v", err)
	}

	keysMap := map[string]*Key{key.Kid: key}
	keysMapInBytes, err := json.Marshal(keysMap)
	if err != nil {
		return fmt.Errorf("failed to marshal keys: %v", err)
	}

	err = cache.Set(
		context.Background(),
		keySetKey,
		keysMapInBytes,
		time.Hour*24*keyExpiryDays,
	).Err()
	if err != nil {
		return fmt.Errorf("failed to store keys in Redis: %v", err)
	}

	return updateJWKSet(keysMap, cache)
}

func generateKey() (*Key, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, rsaKeyBits)
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

func updateJWKSet(keys map[string]*Key, cache *redis.Client) error {
	keySet := jwk.NewSet()
	for _, key := range keys {
		jwkKey, err := jwk.New(key.PublicKey)
		if err != nil {
			return fmt.Errorf("failed to create JWK: %v", err)
		}
		if err := jwkKey.Set(jwk.KeyIDKey, key.Kid); err != nil {
			return fmt.Errorf("failed to set key ID: %v", err)
		}
		keySet.Add(jwkKey)
	}

	jwksJSON, err := json.Marshal(keySet)
	if err != nil {

		return fmt.Errorf("failed to marshal JWKS: %v", err)
	}

	err = cache.Set(
		context.Background(),
		jwksPrefix,
		jwksJSON,
		time.Hour*24*keyExpiryDays,
	).Err()
	if err != nil {
		return fmt.Errorf("failed to store JWKS in Redis: %v", err)
	}

	return nil
}

func CreateJWT(userID uuid.UUID, cache *redis.Client) (string, error) {
	keyMutex.RLock()
	defer keyMutex.RUnlock()

	keys, err := getKeys(cache)
	if err != nil {
		return "", fmt.Errorf("failed to get keys: %v", err)
	}

	if len(keys) == 0 {
		return "", fmt.Errorf("no keys available")
	}

	// Use the most recent key
	var latestKey *Key
	var latestTime time.Time
	for _, k := range keys {
		if k.CreatedAt.After(latestTime) {
			latestKey = k
			latestTime = k.CreatedAt
		}
	}

	expiration := time.Second * time.Duration(config.TokenExpiry)

	claims := jwt.MapClaims{
		"iss": "github.com/shammianand/go-auth",
		"sub": userID.String(),
		"exp": time.Now().Add(expiration).Unix(),
		"iat": time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = latestKey.Kid

	tokenString, err := token.SignedString(latestKey.PrivateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %v", err)
	}

	err = cache.Set(
		context.Background(),
		fmt.Sprintf("%s%s", tokenPrefix, userID.String()),
		tokenString,
		expiration,
	).Err()
	if err != nil {
		return "", fmt.Errorf("failed to store token in Redis: %v", err)
	}

	return tokenString, nil
}

func RefreshToken(cache *redis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		oldTokenString := getTokenFromRequest(r)
		oldToken, err := validateToken(oldTokenString, cache)
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

		storedToken, err := cache.Get(context.Background(), fmt.Sprintf("%s%s", tokenPrefix, userID)).Result()
		if err != nil || storedToken != oldTokenString {
			utils.WriteError(w, http.StatusUnauthorized, fmt.Errorf("token not found or invalid"))
			return
		}

		userUUID, _ := uuid.Parse(userID)
		newTokenString, err := CreateJWT(userUUID, cache)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("failed to create new token"))
			return
		}

		cache.Del(context.Background(), fmt.Sprintf("%s%s", tokenPrefix, userID))

		utils.WriteJSON(w, http.StatusOK, map[string]string{"token": newTokenString})
	}
}

func WithJWTAuth(handlerFunc http.HandlerFunc, cache *redis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenString := getTokenFromRequest(r)
		token, err := validateToken(tokenString, cache)
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

		storedToken, err := cache.Get(context.Background(), fmt.Sprintf("%s%s", tokenPrefix, userID)).Result()
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

func validateToken(tokenString string, cache *redis.Client) (*jwt.Token, error) {
	keyMutex.RLock()
	defer keyMutex.RUnlock()

	keys, err := getKeys(cache)
	if err != nil {
		return nil, fmt.Errorf("failed to get keys: %v", err)
	}

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

func JWKSHandler(cache *redis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		jwksJSON, err := cache.Get(context.Background(), "auth:jwks").Result()
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(jwksJSON))
	}
}

func getKeys(cache *redis.Client) (map[string]*Key, error) {
	keysJSON, err := cache.Get(context.Background(), keySetKey).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get keys from Redis: %v", err)
	}

	var keys map[string]*Key
	if err := json.Unmarshal([]byte(keysJSON), &keys); err != nil {
		return nil, fmt.Errorf("failed to unmarshal keys: %v", err)
	}

	return keys, nil
}

// GetPublicKeyFromCache retrieves a public key by kid from cache
func GetPublicKeyFromCache(cache *redis.Client, kid string) (*rsa.PublicKey, error) {
	keys, err := getKeys(cache)
	if err != nil {
		return nil, fmt.Errorf("failed to get keys: %v", err)
	}

	key, found := keys[kid]
	if !found {
		return nil, fmt.Errorf("key %s not found", kid)
	}

	return key.PublicKey, nil
}
