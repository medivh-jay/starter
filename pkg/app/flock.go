package app

import (
	"errors"
	"os"
	"syscall"
)

var (
	canNotOpenFile = errors.New("flock: open file error")
	lockFileError  = errors.New("flock: lock file error")
	writeFileError = errors.New("flock: write file error")
)

// 文件锁
type Flock struct {
	f    *os.File
	file string
}

// 获得文件锁
func FLock(file string) (*Flock, error) {
	f, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		Logger().WithField("log_type", "pkg.app.flock").Error("lock: open file error ", err)
		return nil, canNotOpenFile
	}

	lock := &Flock{
		f:    f,
		file: file,
	}
	err = lock.Lock()
	return lock, err
}

// 加锁
func (lock *Flock) Lock() error {
	err := syscall.Flock(int(lock.f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
	if err != nil {
		Logger().WithField("log_type", "pkg.app.flock").Error("lock: lock error: ", err)
		return lockFileError
	}
	return nil
}

// 向被锁文件中写入数据
func (lock *Flock) WriteTo(body string) error {
	_ = lock.f.Truncate(0)
	if _, err := lock.f.WriteString(body); err != nil {
		Logger().WithField("log_type", "pkg.app.flock").Error("write error: ", err)
		return writeFileError
	}
	return nil
}

// 解锁, 将同时删除锁文件
func (lock *Flock) UnLock() {
	_ = lock.f.Close()
	_ = os.Remove(lock.file)
}
