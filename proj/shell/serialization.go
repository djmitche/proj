package shell

import (
	"encoding/json"
	"log"
	"os"
)

/* transmitting contexts over file descriptors */

func LoadContext(cfd int) (*Context, error) {
	if cfd == 0 {
		return &Context{
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
		return nil, err
	}
	ctxfile.Close()

	log.Printf("got context %#v\n", context)

	return &context, nil
}

func WriteContext(context *Context, w *os.File) error {
	encoder := json.NewEncoder(w)
	defer w.Close()
	return encoder.Encode(context)
}
