package kamis

import (
	"git.kanosolution.net/kano/kaos"
)

type NeedAccessOptions struct {
	Permission          string
	RequiredAccessLevel int
	CheckFunction       func(ctx *kaos.Context, permission string, accessLevel int) error
}

func NeedAccess(needAccess NeedAccessOptions) func(*kaos.Context) (bool, error) {
	return func(ctx *kaos.Context) (bool, error) {
		if needAccess.CheckFunction == nil {
			//return false, errors.New("need-access middleware is used but no implementation yet")
			return true, nil
		}

		return true, needAccess.CheckFunction(ctx, needAccess.Permission, needAccess.RequiredAccessLevel)
	}
}
