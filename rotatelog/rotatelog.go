package rotatelog

import (
	"os"
	"sync"
	"time"
)

type RotateWriter struct {
	lock     sync.Mutex
	filename string
	fp       *os.File
}

var rw RotateWriter

func Rotate() error {
	return rw.Rotate()
}
func Default(f string) *RotateWriter {
	rw.filename = f
	return &rw
}

func New(filename string) (*RotateWriter, error) {
	w := &RotateWriter{filename: filename}
	err := w.Rotate()
	return w, err
}

func (w *RotateWriter) Write(output []byte) (int, error) {
	w.lock.Lock()
	defer w.lock.Unlock()
	return w.fp.Write(output)
}

func (w *RotateWriter) Rotate() error {
	w.lock.Lock()
	defer w.lock.Unlock()

	var err error
	if w.fp != nil {
		err = w.fp.Close()
		w.fp = nil
		if err != nil {
			return err
		}
	}

	_, err = os.Stat(w.filename)
	if err == nil {
		err = os.Rename(w.filename, w.filename+"."+time.Now().Format(time.RFC3339))
		if err != nil {
			return err
		}
	}

	w.fp, err = os.Create(w.filename)
	return err
}
