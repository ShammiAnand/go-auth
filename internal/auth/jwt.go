package auth

import (
	"context"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/redis/go-redis/v9"
	"github.com/shammianand/go-auth/internal/config"
	"github.com/shammianand/go-auth/internal/utils"
)

var (
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
	keySet     jwk.Set
	keyMutex   sync.RWMutex
)

func InitializeKeys() error {
	var err error
	privateKey, err = jwt.ParseRSAPrivateKeyFromPEM([]byte("SOME PRIVATE KEY SHOULD COME FROM ENV"))
	if err != nil {
		return fmt.Errorf("failed to parse private key: %v", err)
	}
	publicKey = &privateKey.PublicKey

	key, err := jwk.New(publicKey)
	if err != nil {
		return fmt.Errorf("failed to create JWK: %v", err)
	}

	if err := key.Set(jwk.KeyIDKey, "my-key-id"); err != nil {
		return fmt.Errorf("failed to set key ID: %v", err)
	}

	keySet = jwk.NewSet()
	keySet.Add(key)

	return nil
}

func CreateJWT(userID int) (string, error) {
	keyMutex.RLock()
	defer keyMutex.RUnlock()
	// expiration := time.Second * time.Duration(config.TokenExpiry)

	// TODO: follow RFC 7519
	claims := jwt.MapClaims{
		// "userID":    strconv.Itoa(userID),
		// "expiredAt": time.Now().Add(expiration).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = config.ENV_SECRET_KEY_ID

	tokenString, err := token.SignedString(privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %v", err)
	}

	return tokenString, nil
}

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

		userID, _ := strconv.Atoi(claims["userID"].(string))
		expirationTime := int64(claims["expiredAt"].(float64))

		if time.Now().Unix() > expirationTime {
			log.Println("TOKEN EXPIRED")
			permissionDenied(w)
			return
		}

		// FIXME: this is hacky for now
		// user, err := store.GetUserByID(userID)
		// if err != nil {
		// 	log.Printf("failed to get user: %v", err)
		// 	permissionDenied(w)
		// 	return
		// }

		ctx := context.WithValue(r.Context(), "userID", userID)
		handlerFunc(w, r.WithContext(ctx))
	}
}

func getTokenFromRequest(r *http.Request) string {
	return r.Header.Get("Authorization")
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
		key, found := keySet.LookupKeyID(kid)
		if !found {
			return nil, fmt.Errorf("key %v not found", kid)
		}
		var publicKey interface{}
		if err := key.Raw(&publicKey); err != nil {
			return nil, fmt.Errorf("failed to get raw public key: %v", err)
		}
		return publicKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %v", err)
	}

	return token, nil
}

func permissionDenied(w http.ResponseWriter) {
	utils.WriteError(w, http.StatusForbidden, fmt.Errorf("permission denied"))
}

func GetUserIdFromContext(ctx context.Context) int {
	userID, ok := ctx.Value("userID").(int)
	if !ok {
		return -1
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
