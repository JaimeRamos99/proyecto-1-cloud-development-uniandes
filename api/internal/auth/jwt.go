package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TokenManager encapsulates JWT signing and verification.
type TokenManager struct {
	// Secret is the HMAC secret for signing tokens. In production, load from configuration or a secret manager.
	Secret []byte
	// Issuer identifies this service. Used in token claims.
	Issuer string
}

// RegisteredClaims returns standard JWT registered claims with configured issuer and sensible defaults.
func (t TokenManager) RegisteredClaims(subject string, ttl time.Duration) jwt.RegisteredClaims {
	return jwt.RegisteredClaims{
		Issuer:    t.Issuer,
		Subject:   subject,
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
	}
}

// CreateToken signs a JWT with the provided custom claims map merged with registered claims.
func (t TokenManager) CreateToken(subject string, ttl time.Duration, customClaims map[string]any) (string, error) {
	claims := jwt.MapClaims{}
	for k, v := range customClaims {
		claims[k] = v
	}
	// Merge registered claims
	reg := t.RegisteredClaims(subject, ttl)
	claims["iss"] = reg.Issuer
	claims["sub"] = reg.Subject
	claims["iat"] = reg.IssuedAt.Unix()
	claims["exp"] = reg.ExpiresAt.Unix()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(t.Secret)
}

// VerifyToken parses and validates a JWT and returns its claims if valid.
func (t TokenManager) VerifyToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return t.Secret, nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, errors.New("invalid token")
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}
	return claims, nil
}
