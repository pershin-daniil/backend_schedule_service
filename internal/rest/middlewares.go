package rest

import (
	"context"
	"crypto/rsa"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v4"
	"github.com/pershin-daniil/TimeSlots/pkg/models"
)

type ctxClaimsType string

const ctxClaimsStr ctxClaimsType = "claims"

var ErrUnauthorised = errors.New("unauthorized")

func (s *Server) jwtAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			s.writeResponse(w, http.StatusUnauthorized, ErrUnauthorised)
			return
		}
		headerParts := strings.Split(authHeader, " ")
		if len(headerParts) != 2 {
			s.writeResponse(w, http.StatusUnauthorized, ErrUnauthorised)
			return
		}
		if headerParts[0] != "Bearer" {
			s.writeResponse(w, http.StatusUnauthorized, ErrUnauthorised)
			return
		}
		claims, err := parseToken(headerParts[1], s.publicKey)
		if err != nil {
			s.writeResponse(w, http.StatusUnauthorized, ErrUnauthorised)
			return
		}
		r = r.WithContext(context.WithValue(r.Context(), ctxClaimsStr, claims))
		next.ServeHTTP(w, r)
	})
}

func (s *Server) getClaims(ctx context.Context) *models.Claims {
	claims, ok := ctx.Value(ctxClaimsStr).(*models.Claims)
	if !ok {
		return nil
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
