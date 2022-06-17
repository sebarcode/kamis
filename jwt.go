package kamis

import (
	"errors"
	"net/http"
	"strings"

	"git.kanosolution.net/kano/kaos"
	"github.com/golang-jwt/jwt"
	"github.com/sebarcode/codekit"
	"github.com/sebarcode/siam"
)

type ValidateFn func(string, *siam.Session) error

type JWTSetupOptions struct {
	GetSessionMethod string
	GetSessionTopic  string
	ValidateFunction ValidateFn
	EnrichFunction   func(*kaos.Context, *siam.Session)
}

func JWT(headerName, secret string, requiredJWT bool, opts JWTSetupOptions) func(ctx *kaos.Context) (bool, error) {
	if headerName == "" {
		headerName = "Authorization"
	}

	return func(ctx *kaos.Context) (bool, error) {
		req := ctx.Data().Get("http-request", new(http.Request)).(*http.Request)
		if token := req.Header.Get(headerName); token != "" {
			if strings.HasPrefix(token, "Bearer ") {
				token = strings.ReplaceAll(token, "Bearer ", "")
			}

			bc := jwt.StandardClaims{}
			tkn, e := jwt.ParseWithClaims(token, &bc, func(tkn *jwt.Token) (interface{}, error) {
				return []byte(secret), nil
			})

			if e != nil {
				return true, nil
			}

			if !tkn.Valid {
				return true, nil
			}

			ctx.Data().Set("jwt_token_id", bc.Id)

			sess := new(siam.Session)
			switch opts.GetSessionMethod {
			case "NATS":
				ev, _ := ctx.DefaultEvent()
				if ev == nil {
					return false, errors.New("invalid pubsub handler")
				}

				getSessionTopic := opts.GetSessionTopic
				if getSessionTopic == "" {
					return false, errors.New("invalid pubsub topic")
				}
				if e = ev.Publish(getSessionTopic, codekit.M{}.Set("ID", bc.Id), sess); e != nil {
					if requiredJWT {
						return false, errors.New("invalid access token")
					}
					return true, nil
				}

			default:
				fn := opts.ValidateFunction
				if fn == nil {
					return false, errors.New("invalid function to validate token")
				}
				e = fn(bc.Id, sess)
				if e != nil {
					return true, nil
				}
			}

			if opts.EnrichFunction != nil {
				opts.EnrichFunction(ctx, sess)
			}

			ctx.Data().Set("jwt_session_id", sess.SessionID)
			ctx.Data().Set("jwt_reference_id", sess.ReferenceID)
			ctx.Data().Set("jwt_data", sess.Data)
		}
		return true, nil
	}
}

func NeedJWT() func(ctx *kaos.Context) (bool, error) {
	return func(ctx *kaos.Context) (bool, error) {
		jwtRefID := ctx.Data().Get("jwt_reference_id", "").(string)
		if jwtRefID == "" {
			return false, errors.New("invalid access token")
		}
		return true, nil
	}
}
