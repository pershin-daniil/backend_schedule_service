package rest

import (
	"context"
	"crypto/rsa"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v4"
	"github.com/pershin-daniil/TimeSlots/pkg/models"
)

func (s *Server) jwtAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		headerParts := strings.Split(authHeader, " ")
		if len(headerParts) != 2 {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		if headerParts[0] != "Bearer" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		claims, err := parseToken(headerParts[1], s.publicKey)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		r = r.WithContext(context.WithValue(r.Context(), "claims", claims))
		next.ServeHTTP(w, r)
	})
}

func (s *Server) getClaims(ctx context.Context) models.Claims {
	claims, ok := ctx.Value("claims").(models.Claims)
	if !ok {
		return models.Claims{}
	}
	return claims
}

func parseToken(accessToken string, key *rsa.PublicKey) (*models.Claims, error) {
	token, err := jwt.ParseWithClaims(accessToken, &models.Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("invalid signing method: %v", token.Header["alg"])
		}
		return key, nil
	})
	if err != nil {
		return nil, fmt.Errorf("err parsing token: %w", err)
	}
	claims, ok := token.Claims.(*models.Claims)
	if !ok {
		return nil, fmt.Errorf("invalid claims")
	}
	return claims, nil
}
