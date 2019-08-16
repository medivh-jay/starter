package log

import (
	"io"
	"log"
	"os"
)

func Start() {
	log.SetOutput(io.MultiWriter(os.Stdout))
	log.SetFlags(log.LstdFlags | log.Llongfile)
	log.SetPrefix("[STARTER] ")
}
