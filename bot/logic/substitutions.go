package logic

import (
	"fmt"
	"github.com/TicketsBot/worker/bot/command/registry"
	"github.com/rxdn/gdl/objects/member"
	"github.com/rxdn/gdl/objects/user"
	"strings"
)

type SubstitutionFunc func(user user.User, member member.Member) string

type Substitutor struct {
	Placeholder string
	NeedsUser   bool
	NeedsMember bool
	F           SubstitutionFunc
}

func NewSubstitutor(placeholder string, needsUser, needsMember bool, f SubstitutionFunc) Substitutor {
	return Substitutor{
		Placeholder: placeholder,
		NeedsUser:   needsUser,
		NeedsMember: needsMember,
		F:           f,
	}
}

func doSubstitutions(ctx registry.CommandContext, s string, userId uint64, substitutors []Substitutor) (string, error) {
	var needsUser, needsMember bool

	// Determine which objects we need to fetch
	for _, substitutor := range substitutors {
		if substitutor.NeedsUser {
			needsUser = true
		}

		if substitutor.NeedsMember {
			needsMember = true
		}

		if needsUser && needsMember {
			break
		}
	}

	// Retrieve user and member if necessary
	var user user.User
	var member member.Member

	var err error
	if needsUser {
		if ctx.UserId() == userId {
			user, err = ctx.User()
		} else {
			user, err = ctx.Worker().GetUser(userId)
		}
	}

	if err != nil {
		return "", err
	}

	if needsMember {
		if ctx.UserId() == userId {
			member, err = ctx.Member()
		} else {
			member, err = ctx.Worker().GetGuildMember(ctx.GuildId(), userId)
		}
	}

	if err != nil {
		return "", err
	}

	for _, substitutor := range substitutors {
		placeholder := fmt.Sprintf("%%%s%%", substitutor.Placeholder)

		if strings.Contains(s, placeholder) {
			s = strings.ReplaceAll(s, placeholder, substitutor.F(user, member))
		}
	}

	return s, nil
}
