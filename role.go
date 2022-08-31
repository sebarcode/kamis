package kamis

import (
	"git.kanosolution.net/kano/kaos"
)

type NeedAccessOptions struct {
	Permission          string
	RequiredAccessLevel int
	CheckFunction       func(ctx *kaos.Context, parm interface{}, permission string, accessLevel int) error
}

func NeedAccess(needAccess NeedAccessOptions) func(*kaos.Context, interface{}) (bool, error) {
	return func(ctx *kaos.Context, parm interface{}) (bool, error) {
		if needAccess.CheckFunction == nil {
			//return false, errors.New("need-access middleware is used but no implementation yet")
			return true, nil
		}

		return true, needAccess.CheckFunction(ctx, parm, needAccess.Permission, needAccess.RequiredAccessLevel)
	}
}
