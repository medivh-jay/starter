package app

// 以下代码来自 gin, recovery.go
import (
	"bytes"
	"io/ioutil"
	"runtime"
)

var (
	dunno     = []byte("???")
	centerDot = []byte("·")
	dot       = []byte(".")
	slash     = []byte("/")
)

func function(pc uintptr) []byte {
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return dunno
	}
	name := []byte(fn.Name())
	if lastSlash := bytes.LastIndex(name, slash); lastSlash >= 0 {
		name = name[lastSlash+1:]
	}
	if period := bytes.Index(name, dot); period >= 0 {
		name = name[period+1:]
	}
	name = bytes.Replace(name, centerDot, dot, -1)
	return name
}

func source(lines [][]byte, n int) []byte {
	if n < 0 || n >= len(lines) {
		return []byte("???")
	}
	return bytes.TrimSpace(lines[n])
}

func Stack(skip int) []map[string]interface{} {
	var stack = make([]map[string]interface{}, 0, 0)

	var lines [][]byte
	var lastFile string
	for i := skip; ; i++ {
		stackInfo := make(map[string]interface{})
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		stackInfo["file"] = file
		stackInfo["line"] = line
		stackInfo["func"] = string(function(pc))

		if file != lastFile {
			data, err := ioutil.ReadFile(file)
			if err != nil {
				continue
			}
			lines = bytes.Split(data, []byte{'\n'})
			lastFile = file
		}

		stackInfo["source"] = string(source(lines, line))
		stack = append(stack, stackInfo)
	}
	return stack
}
