package kamis

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"git.kanosolution.net/kano/kaos"
	"github.com/golang-jwt/jwt/v5"
	"github.com/sebarcode/codekit"
	"github.com/sebarcode/siam"
)

type localAuthJwt struct {
	jwt.RegisteredClaims
	Data codekit.M
}

type ValidateFn func(string, *siam.Session) error

type JWTSetupOptions struct {
	Secret           string
	GetSessionMethod string
	GetSessionTopic  string
	ValidateFunction ValidateFn
	EnrichFunction   func(*kaos.Context, *siam.Session)
}

func JWT(opts JWTSetupOptions) func(ctx *kaos.Context, parm interface{}) (bool, error) {
	headerName := "Authorization"

	return func(ctx *kaos.Context, parm interface{}) (bool, error) {
		if opts.Secret == "" {
			return false, errors.New("secret is blank")
		}

		req := ctx.Data().Get("http_request", new(http.Request)).(*http.Request)
		if token := req.Header.Get(headerName); token != "" {
			if strings.HasPrefix(token, "Bearer ") {
				token = strings.ReplaceAll(token, "Bearer ", "")
			}

			bc := localAuthJwt{}
			tkn, e := jwt.ParseWithClaims(token, &bc, func(tkn *jwt.Token) (interface{}, error) {
				return []byte(opts.Secret), nil
			})

			if e != nil {
				return true, nil
			}

			if !tkn.Valid {
				return true, nil
			}

			var expiryAt int64
			if bc.RegisteredClaims.ExpiresAt != nil {
				expiryAt = bc.RegisteredClaims.ExpiresAt.Time.UnixMilli()
			}
			timeNow := time.Now().UnixMilli()
			if expiryAt != 0 && expiryAt < timeNow {
				return false, errors.New("credentials token is expired")
			}

			ctx.Data().Set("jwt_token", token)
			ctx.Data().Set("jwt_data", bc.Data)

			sess := new(siam.Session)
			switch opts.GetSessionMethod {
			case "NATS":
				ev, _ := ctx.DefaultEvent()
				if ev == nil {
					return false, errors.New("invalid event hub")
				}

				getSessionTopic := opts.GetSessionTopic
				if getSessionTopic == "" {
					return false, errors.New("invalid topic")
				}
				if e = ev.Publish(getSessionTopic, codekit.M{}.Set("ID", bc.RegisteredClaims.ID), sess, nil); e != nil {
					ctx.Log().Warningf("get session fail: topic %s | id %s | msg %s", getSessionTopic, bc.RegisteredClaims.ID, e.Error())
					return true, nil
				}

			default:
				fn := opts.ValidateFunction
				if fn != nil {
					e = fn(bc.RegisteredClaims.ID, sess)
					if e != nil {
						return true, nil
					}
				}
			}

			if opts.EnrichFunction != nil {
				opts.EnrichFunction(ctx, sess)
			}

			ctx.Data().Set("jwt_session_id", sess.SessionID)
			ctx.Data().Set("jwt_reference_id", sess.ReferenceID)
			ctx.Data().Set("jwt_session_data", sess.Data)
		}
		return true, nil
	}
}

func NeedJWT() func(ctx *kaos.Context, parm interface{}) (bool, error) {
	return func(ctx *kaos.Context, parm interface{}) (bool, error) {
		jwtRefID := ctx.Data().Get("jwt_reference_id", "").(string)
		if jwtRefID == "" {
			return false, errors.New("invalid access token")
		}
		return true, nil
	}
}
