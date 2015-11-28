// The shell package handles starting a shell at the end of a proj path.
//
// Shells are started with a "context" which evolves as proj traverses the path.
// In particular, the context accumulates shell modifiers along the way.  Shell
// modifiiers are specified like this in the configuration file:
//
//     shell:
//         env:
//             FOO=foo
//             BAR=bar
//
// Here, `env` is the modifier type, and `{"FOO": "foo", "BAR": bar}` are its
// arguments.  `env.go` provides a good example of a modifier implementation.
//
// Modifiers call methods on a Shell to control the behavior of that shell; this
// allows different shells to interoperate smoothly with modifiers.
//
// When traversing the proj path involves hopping from host to host or some
// other transition that requires spawning a new process, the entire Context
// is encoded by the parent process and decoded in the child.  As such, all
// Modifier implementations must be JSONable.
package shell

import (
	"fmt"
	"github.com/djmitche/proj/proj/config"
	"github.com/djmitche/proj/proj/util"
)

// A context for execution of a shell
type Context struct {
	Shell     string
	Path      []string
	Modifiers []Modifier
}

// A modifier of shell behavior
type Modifier interface {
	Modify(shell Shell) error
}

type modifierFactory func(interface{}) (Modifier, error)

var modifierFactories map[string]modifierFactory = make(map[string]modifierFactory)

func newModifier(raw interface{}) (Modifier, error) {
	modType, args, err := util.SingleKeyMap(raw)
	if err != nil {
		return nil, fmt.Errorf("interpreting %q: %s", raw, err)
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
