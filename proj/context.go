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
	Modifiers []interface{}
}

func load_context(cfd int) Context {
	if cfd == 0 {
		return Context{
			Shell:     "bash", // TODO from supported_shells
			Path:      []string{},
			Modifiers: make([]interface{}, 0),
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
