package conpty

import (
	"os"
	"syscall"
)

type pty interface {
	Close() error
	Fd() uintptr
	InPipe() *os.File
	InPipeFd() uintptr
	OutPipe() *os.File
	OutPipeFd() uintptr
	Read(p []byte) (n int, err error)
	Resize(w int, h int) error
	Size() (w int, h int, err error)
	Spawn(name string, args []string, attr *syscall.ProcAttr) (pid int, handle uintptr, err error)
	Write(p []byte) (n int, err error)
}

var _ pty = &ConPty{}
