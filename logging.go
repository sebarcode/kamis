package kamis

import (
	"git.kanosolution.net/kano/kaos"
)

func Logging(fn func(ctx *kaos.Context) string) func(ctx *kaos.Context) (bool, error) {
	return func(ctx *kaos.Context) (bool, error) {
		if fn == nil {
			ctx.Log().Infof("Accessing %s", ctx.Data().Get("path", ""))
			return true, nil
		}
		ctx.Log().Infof(fn(ctx))
		return true, nil
	}
}
