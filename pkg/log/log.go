package log

import (
	"io"
	"log"
	"os"
)

func init() {
	log.SetOutput(io.MultiWriter(os.Stdout))
	log.SetFlags(log.LstdFlags | log.Llongfile)
	log.SetPrefix("[SNYU] ")
}
