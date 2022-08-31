package kamis

import (
	"git.kanosolution.net/kano/kaos"
)

func Logging(fn func(*kaos.Context, interface{}) string) func(ctx *kaos.Context, parm interface{}) (bool, error) {
	return func(ctx *kaos.Context, parm interface{}) (bool, error) {
		if fn == nil {
			ctx.Log().Infof("Accessing %s", ctx.Data().Get("path", ""))
			return true, nil
		}
		ctx.Log().Infof(fn(ctx, parm))
		return true, nil
	}
}
