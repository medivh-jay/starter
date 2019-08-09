package log

import (
	"errors"
	"starter/pkg/elastic"
)

type EsLog struct {
	Index string
}

func (esLog *EsLog) Write(p []byte) (n int, err error) {
	rs := elastic.InsertString(esLog.Index, string(p))
	if rs.Status == 200 {
		return len(p), nil
	}
	return len(p), errors.New("elasticsearch log post error")
}
