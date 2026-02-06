package jwt

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type claims interface {
	jwt.Claims
	setExpiresAt(expiresAt time.Time)
}

type UserClaims struct {
	UserId int `json:"userId,omitempty"`
	jwt.RegisteredClaims
}

func NewUserClaims(userId int) *UserClaims {
	return &UserClaims{
		UserId: userId,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:       uuid.New().String(),
			IssuedAt: jwt.NewNumericDate(time.Now()),
		},
	}
}

func (c *UserClaims) setExpiresAt(expiresAt time.Time) {
	c.ExpiresAt = jwt.NewNumericDate(expiresAt)
}
