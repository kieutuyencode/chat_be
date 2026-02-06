package auth

import (
	"backend/apperror"
	"backend/common"
	"backend/config"
	"backend/database/ent"
	"backend/database/ent/user"
	"backend/database/ent/verificationcode"
	"backend/notification/mail"
	"backend/security/jwt"
	"context"
	"strings"
	"time"

	"github.com/cockroachdb/errors"
	"go.uber.org/fx"
)

type Auth struct {
	jwt     jwt.Jwt
	mail    *mail.Mail
	handler apperror.Handler
}

type authParams struct {
	fx.In
	Jwt     jwt.Jwt
	Mail    *mail.Mail
	Handler apperror.Handler
}

func newAuth(p authParams) *Auth {
	return &Auth{
		jwt:     p.Jwt,
		mail:    p.Mail,
		handler: p.Handler,
	}
}

func (a *Auth) SignIn(ctx context.Context, client *ent.Client, p *SignInParams) error {
	p.Email = strings.ToLower(p.Email)

	user, err := client.User.Query().Where(user.EmailEQ(p.Email)).First(ctx)
	if err != nil && !ent.IsNotFound(err) {
		return errors.Wrap(err, "User.Query() failed")
	}

	if user == nil {
		user, err = client.User.Create().SetFullname(p.Email).SetEmail(p.Email).Save(ctx)
		if err != nil {
			return errors.Wrap(err, "User.Create() failed")
		}
	}

	verificationCode, err := a.GenerateVerificationCode(ctx, client, &GenerateVerificationCodeParams{
		UserId:          user.ID,
		ExpiresInMinute: config.VerifySignInExpiresInMinute,
	})
	if err != nil {
		return err
	}

	go func() {
		a.handler(func() error {
			return a.mail.SendSignIn(&mail.SendSignInParams{
				To:              []string{p.Email},
				Code:            verificationCode,
				ExpiresInMinute: config.VerifySignInExpiresInMinute,
			})
		})
	}()

	return nil
}

func (a *Auth) GenerateVerificationCode(ctx context.Context, client *ent.Client, p *GenerateVerificationCodeParams) (string, error) {
	code, err := common.GenerateOTP(6)
	if err != nil {
		return "", err
	}

	expiresAt := time.Now().Add(time.Minute * time.Duration(p.ExpiresInMinute))

	verificationCode, err := client.VerificationCode.Query().Where(verificationcode.UserIdEQ(p.UserId)).First(ctx)
	if err != nil && !ent.IsNotFound(err) {
		return "", errors.Wrap(err, "VerificationCode.Query() failed")
	}

	if verificationCode != nil {
		_, err = verificationCode.Update().SetCode(code).SetExpiresAt(expiresAt).Save(ctx)
		if err != nil {
			return "", errors.Wrap(err, "VerificationCode.Update() failed")
		}
	} else {
		_, err = client.VerificationCode.Create().SetCode(code).SetExpiresAt(expiresAt).SetUserId(p.UserId).Save(ctx)
		if err != nil {
			return "", errors.Wrap(err, "VerificationCode.Create() failed")
		}
	}

	return code, nil
}

func (a *Auth) VerifyCode(ctx context.Context, client *ent.Client, p *VerifyCodeParams) error {
	verificationCode, err := client.VerificationCode.Query().
		Where(verificationcode.UserIdEQ(p.UserId), verificationcode.ExpiresAtGT(time.Now())).
		First(ctx)
	if err != nil {
		return errors.Wrap(err, "VerificationCode.Query() failed")
	}

	if verificationCode == nil {
		return apperror.BadRequest(messageInvalidCode, nil, nil)
	}

	if verificationCode.Code != p.Code {
		return apperror.BadRequest(messageInvalidCode, nil, nil)
	}

	return nil
}

func (a *Auth) VerifySignIn(ctx context.Context, client *ent.Client, p *VerifySignInParams) (string, error) {
	p.Email = strings.ToLower(p.Email)

	user, err := client.User.Query().Where(user.EmailEQ(p.Email)).First(ctx)
	if err != nil && !ent.IsNotFound(err) {
		return "", errors.Wrap(err, "User.Query() failed")
	}
	if user == nil {
		return "", apperror.BadRequest(messageInvalidEmail, nil, nil)
	}

	if err = a.VerifyCode(ctx, client, &VerifyCodeParams{
		UserId: user.ID,
		Code:   p.Code,
	}); err != nil {
		return "", err
	}

	claims := jwt.NewUserClaims(user.ID)
	accessToken, err := a.jwt.GenerateAccessToken(claims)
	if err != nil {
		return "", err
	}

	return accessToken, nil
}

const (
	messageInvalidEmail = "Email not found"
	messageInvalidCode  = "Mã xác nhận không hợp lệ"
)

type SignInParams struct {
	Email string
}

type GenerateVerificationCodeParams struct {
	ExpiresInMinute int
	UserId          int
}

type VerifyCodeParams struct {
	UserId int
	Code   string
}

type VerifySignInParams struct {
	Email string
	Code  string
}
