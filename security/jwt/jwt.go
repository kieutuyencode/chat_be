package jwt

import (
	"backend/config"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/fx"
)

type Jwt interface {
	GenerateAccessToken(claims claims) (string, error)
	ValidateAccessToken(tokenString string, claims claims) error
}

type defaultJwt struct {
	accessTokenSecretKey []byte
	accessTokenExpiresIn time.Duration
}

type jwtParams struct {
	fx.In
	Env *config.Env
}

func newJwt(p jwtParams) Jwt {
	return &defaultJwt{
		accessTokenSecretKey: []byte(p.Env.JwtAccessTokenSecretKey),
		accessTokenExpiresIn: p.Env.JwtAccessTokenExpiresIn,
	}
}

func (j *defaultJwt) GenerateAccessToken(claims claims) (string, error) {
	return j.generateToken(claims, j.accessTokenSecretKey, j.accessTokenExpiresIn)
}

func (j *defaultJwt) generateToken(claims claims, secretKey []byte, expiresIn time.Duration) (string, error) {
	claims.setExpiresAt(time.Now().Add(expiresIn))

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString(secretKey)
	if err != nil {
		return "", errors.Wrap(err, "SignedString failed")
	}

	return tokenStr, nil
}

func (j *defaultJwt) ValidateAccessToken(tokenString string, claims claims) error {
	return validateToken(tokenString, j.accessTokenSecretKey, claims)
}

func validateToken(tokenString string, secretKey []byte, claims claims) error {
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return secretKey, nil
	})

	if err != nil {
		return errors.Wrap(err, "ParseWithClaims failed")
	}

	if !token.Valid {
		return errors.New("invalid token")
	}

	return nil
}
