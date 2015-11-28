package shell

import (
	"fmt"
	"github.com/djmitche/proj/proj/config"
	"github.com/djmitche/proj/proj/util"
)

type Context struct {
	Shell     string
	Path      []string
	Modifiers []Modifier
}

type Modifier interface {
	Apply(shell Shell) error
}

type modifierFactory func(interface{}) (Modifier, error)

var modifierFactories map[string]modifierFactory = make(map[string]modifierFactory)

func newModifier(raw interface{}) (Modifier, error) {
	modType, args, err := util.SingleKeyMap(raw)
	if err != nil {
		return nil, err
	}
	factory, ok := modifierFactories[modType]
	if !ok {
		return nil, fmt.Errorf("unknown modifier type %s", modType)
	}
	return factory(args)
}

// update a context based on a configuration; this amounts to appending the
// config's context modifiers to the context's modifiers
func (ctx *Context) Update(config *config.Config) error {
	for _, elt := range config.Modifiers {
		mod, err := newModifier(elt)
		if err != nil {
			return err
		}
		ctx.Modifiers = append(ctx.Modifiers, mod)
	}
	return nil
}
