package proj

import (
	"encoding/json"
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
	Apply(shell Shell)
}

type modifierFactory func(interface{}) Modifier

var modifierFactories map[string]modifierFactory = make(map[string]modifierFactory)

type modifier struct {
	raw interface{}
}

type envModifier struct {
	variables map[string]string
	modifier
}

func newEnvModifier(args interface{}) Modifier {
	var_map, ok := args.(map[string]interface{})
	if !ok {
		log.Fatal("Invalid env shell modifier %j", args)
	}
	variables := make(map[string]string)
	for n, v := range var_map {
		variables[n], ok = v.(string)
		if !ok {
			log.Fatal("Invalid env shell modifier %j", args)
		}
	}

	return &envModifier{
		//raw:       raw, XXX??
		variables: variables,
	}
}

func (mod *envModifier) Apply(shell Shell) {
	for n, v := range mod.variables {
		log.Printf("applying %s=%s", n, v)
		// TODO: shell quoting
		shell.SetVariable(n, v)
	}
}

func init() {
	modifierFactories["env"] = newEnvModifier
}

func new_modifier(raw interface{}) Modifier {
	mod_type, args, err := singleKeyMap(raw)
	if err != nil {
		log.Fatal(err)
	}
	factory, ok := modifierFactories[mod_type]
	if !ok {
		log.Fatal("unknown modifier type %s", mod_type)
	}
	return factory(args)
}

// update a context based on a configuration; this amounts to appending the
// config's context modifiers to the context's modifiers
func (ctx *Context) Update(config Config) {
	for _, elt := range config.Modifiers {
		ctx.Modifiers = append(ctx.Modifiers, new_modifier(elt))
	}
}

/* transmitting contexts over file descriptors */

func load_context(cfd int) Context {
	if cfd == 0 {
		return Context{
			Shell:     "bash", // TODO from supported_shells
			Path:      []string{},
			Modifiers: make([]Modifier, 0),
		}
	}

	// read from the given file descriptor and close it
	ctxfile := os.NewFile(uintptr(cfd), "ctxfile")
	decoder := json.NewDecoder(ctxfile)
	context := Context{}
	err := decoder.Decode(&context)
	if err != nil {
		log.Panic(err)
	}
	ctxfile.Close()

	log.Printf("got context %#v\n", context)

	return context
}

func write_context(context Context, w *os.File) {
	encoder := json.NewEncoder(w)
	err := encoder.Encode(context)
	if err != nil {
		log.Panic(err)
	}
	w.Close()
}
