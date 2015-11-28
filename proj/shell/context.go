package shell

import (
	"encoding/json"
	"fmt"
	"github.com/djmitche/proj/proj/config"
	"github.com/djmitche/proj/proj/util"
	"log"
	"os"
)

/* Context handling */

type Context struct {
	Shell     string
	Path      []string
	Modifiers []Modifier
}

type Modifier interface {
	//MarhshalJSON() ([]byte, error) -- TODO
	Apply(shell Shell) error
}

type modifierFactory func(interface{}) (Modifier, error)

var modifierFactories map[string]modifierFactory = make(map[string]modifierFactory)

type modifier struct {
	raw interface{}
}

type envModifier struct {
	variables map[string]string
	modifier
}

func newEnvModifier(args interface{}) (Modifier, error) {
	varMap, ok := args.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("Invalid env shell modifier %j", args)
	}
	variables := make(map[string]string)
	for n, v := range varMap {
		variables[n], ok = v.(string)
		if !ok {
			return nil, fmt.Errorf("Invalid env shell modifier %j", args)
		}
	}

	return &envModifier{
		//raw:       raw, XXX??
		variables: variables,
	}, nil
}

func (mod *envModifier) Apply(shell Shell) error {
	for n, v := range mod.variables {
		log.Printf("applying %s=%s", n, v)
		// TODO: shell quoting
		err := shell.SetVariable(n, v)
		if err != nil {
			return err
		}
	}
	return nil
}

func init() {
	modifierFactories["env"] = newEnvModifier
}

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

/* transmitting contexts over file descriptors */

func LoadContext(cfd int) (Context, error) {
	if cfd == 0 {
		return Context{
			Shell:     "bash", // TODO from supportedShells
			Path:      []string{},
			Modifiers: make([]Modifier, 0),
		}, nil
	}

	// read from the given file descriptor and close it
	ctxfile := os.NewFile(uintptr(cfd), "ctxfile")
	decoder := json.NewDecoder(ctxfile)
	context := Context{}
	err := decoder.Decode(&context)
	if err != nil {
		return Context{}, err
	}
	ctxfile.Close()

	log.Printf("got context %#v\n", context)

	return context, nil
}

func WriteContext(context Context, w *os.File) error {
	encoder := json.NewEncoder(w)
	defer w.Close()
	return encoder.Encode(context)
}
