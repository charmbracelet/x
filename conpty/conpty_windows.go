//go:build windows
// +build windows

package conpty

import (
	"errors"
	"fmt"
	"io"
	"sync"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

// ConPty represents a Windows Console Pseudo-terminal.
// https://learn.microsoft.com/en-us/windows/console/creating-a-pseudoconsole-session#preparing-the-communication-channels
type ConPty struct {
	hpc                       *windows.Handle
	inPipeWrite, inPipeRead   windows.Handle
	outPipeWrite, outPipeRead windows.Handle
	attrList                  *windows.ProcThreadAttributeListContainer
	size                      windows.Coord
	closeOnce                 sync.Once
}

var (
	_ io.Writer = &ConPty{}
	_ io.Reader = &ConPty{}
)

// CreatePipes is a helper function to create connected input and output pipes.
func CreatePipes() (inPipeRead, inPipeWrite, outPipeRead, outPipeWrite uintptr, err error) {
	var inPipeReadHandle, inPipeWriteHandle windows.Handle
	var outPipeReadHandle, outPipeWriteHandle windows.Handle
	pSec := &windows.SecurityAttributes{Length: uint32(unsafe.Sizeof(zeroSec)), InheritHandle: 1}

	if err := windows.CreatePipe(&inPipeReadHandle, &inPipeWriteHandle, pSec, 0); err != nil {
		return 0, 0, 0, 0, fmt.Errorf("failed to create input pipes for pseudo console: %w", err)
	}

	if err := windows.CreatePipe(&outPipeReadHandle, &outPipeWriteHandle, pSec, 0); err != nil {
		return 0, 0, 0, 0, fmt.Errorf("failed to create output pipes for pseudo console: %w", err)
	}

	return uintptr(inPipeReadHandle), uintptr(inPipeWriteHandle),
		uintptr(outPipeReadHandle), uintptr(outPipeWriteHandle),
		nil
}

// New creates a new ConPty device.
// Accepts a custom width, height, and flags that will get passed to
// windows.CreatePseudoConsole.
func New(w int, h int, flags int) (*ConPty, error) {
	inPipeRead, inPipeWrite, outPipeRead, outPipeWrite, err := CreatePipes()
	if err != nil {
		return nil, fmt.Errorf("failed to create pipes for pseudo console: %w", err)
	}

	c, err := NewWithPipes(inPipeRead, inPipeWrite, outPipeRead, outPipeWrite, w, h, flags)
	if err != nil {
		return nil, err
	}

	return c, nil
}

// NewWithPipes creates a new ConPty device with the provided pipe handles.
// This is useful for when you want to use existing pipes, such as when
// using a ConPty with a process that has already been created, or when
// you want to use a ConPty with a specific set of pipes for input and output.
//
// The PTY-slave end (input read and output write) of the pipes can be closed
// after the ConPty is created, as the ConPty will take ownership of the handles
// and dup them for the new process that will be spawned. The PTY-master end of
// the pipes will be used to communicate with the pseudo console.
func NewWithPipes(inPipeRead, inPipeWrite, outPipeRead, outPipeWrite uintptr, w int, h int, flags int) (c *ConPty, err error) {
	if w <= 0 {
		w = DefaultWidth
	}
	if h <= 0 {
		h = DefaultHeight
	}

	c = &ConPty{
		hpc: new(windows.Handle),
		size: windows.Coord{
			X: int16(w), Y: int16(h),
		},
		inPipeWrite:  windows.Handle(inPipeWrite),
		inPipeRead:   windows.Handle(inPipeRead),
		outPipeWrite: windows.Handle(outPipeWrite),
		outPipeRead:  windows.Handle(outPipeRead),
	}

	if err := windows.CreatePseudoConsole(c.size, windows.Handle(inPipeRead), windows.Handle(outPipeWrite), uint32(flags), c.hpc); err != nil {
		return nil, fmt.Errorf("failed to create pseudo console: %w", err)
	}

	// Allocate an attribute list that's large enough to do the operations we care about
	// 1. Pseudo console setup
	c.attrList, err = windows.NewProcThreadAttributeList(1)
	if err != nil {
		return nil, err
	}

	if err := c.attrList.Update(
		windows.PROC_THREAD_ATTRIBUTE_PSEUDOCONSOLE,
		unsafe.Pointer(*c.hpc),
		unsafe.Sizeof(*c.hpc),
	); err != nil {
		return nil, fmt.Errorf("failed to update proc thread attributes for pseudo console: %w", err)
	}

	return c, err
}

// Fd returns the ConPty handle.
func (p *ConPty) Fd() uintptr {
	return uintptr(*p.hpc)
}

// Close closes the ConPty device.
func (p *ConPty) Close() error {
	var err error
	p.closeOnce.Do(func() {
		// Ensure that we have the PTY-end of the pipes closed.
		_ = windows.CloseHandle(p.inPipeRead)
		_ = windows.CloseHandle(p.outPipeWrite)
		if p.attrList != nil {
			p.attrList.Delete()
		}
		windows.ClosePseudoConsole(*p.hpc)
		err = errors.Join(
			windows.CloseHandle(p.inPipeWrite),
			windows.CloseHandle(p.outPipeRead),
		)
	})
	return err
}

// InPipeReadFd returns the ConPty input pipe read file descriptor handle.
func (p *ConPty) InPipeReadFd() uintptr {
	return uintptr(p.inPipeRead)
}

// InPipeWriteFd returns the ConPty input pipe write file descriptor handle.
func (p *ConPty) InPipeWriteFd() uintptr {
	return uintptr(p.inPipeWrite)
}

// OutPipeReadFd returns the ConPty output pipe read file descriptor handle.
func (p *ConPty) OutPipeReadFd() uintptr {
	return uintptr(p.outPipeRead)
}

// OutPipeWriteFd returns the ConPty output pipe write file descriptor handle.
func (p *ConPty) OutPipeWriteFd() uintptr {
	return uintptr(p.outPipeWrite)
}

// Write safely writes bytes to master end of the ConPty.
func (c *ConPty) Write(p []byte) (n int, err error) {
	var l uint32
	err = windows.WriteFile(c.inPipeWrite, p, &l, nil)
	return int(l), err
}

// Read safely reads bytes from master end of the ConPty.
func (c *ConPty) Read(p []byte) (n int, err error) {
	var l uint32
	err = windows.ReadFile(c.outPipeRead, p, &l, nil)
	return int(l), err
}

// Resize resizes the pseudo-console.
func (c *ConPty) Resize(w int, h int) error {
	size := windows.Coord{X: int16(w), Y: int16(h)}
	if err := windows.ResizePseudoConsole(*c.hpc, size); err != nil {
		return fmt.Errorf("failed to resize pseudo console: %w", err)
	}
	c.size = size
	return nil
}

// Size returns the current pseudo-console size.
func (c *ConPty) Size() (w int, h int, err error) {
	w = int(c.size.X)
	h = int(c.size.Y)
	return w, h, err
}

var (
	zeroAttr syscall.ProcAttr
	zeroSec  windows.SecurityAttributes
)

// Spawn spawns a new process attached to the pseudo-console.
func (c *ConPty) Spawn(name string, args []string, attr *syscall.ProcAttr) (pid int, handle uintptr, err error) {
	if attr == nil {
		attr = &zeroAttr
	}

	argv0, err := lookExtensions(name, attr.Dir)
	if err != nil {
		return 0, 0, err
	}
	if len(attr.Dir) != 0 {
		// Windows CreateProcess looks for argv0 relative to the current
		// directory, and, only once the new process is started, it does
		// Chdir(attr.Dir). We are adjusting for that difference here by
		// making argv0 absolute.
		var err error
		argv0, err = joinExeDirAndFName(attr.Dir, argv0)
		if err != nil {
			return 0, 0, err
		}
	}

	argv0p, err := windows.UTF16PtrFromString(argv0)
	if err != nil {
		return 0, 0, err
	}

	var cmdline string
	if attr.Sys != nil && attr.Sys.CmdLine != "" {
		cmdline = attr.Sys.CmdLine
	} else {
		cmdline = windows.ComposeCommandLine(args)
	}
	argvp, err := windows.UTF16PtrFromString(cmdline)
	if err != nil {
		return 0, 0, err
	}

	var dirp *uint16
	if len(attr.Dir) != 0 {
		dirp, err = windows.UTF16PtrFromString(attr.Dir)
		if err != nil {
			return 0, 0, err
		}
	}

	if attr.Env == nil {
		attr.Env, err = execEnvDefault(attr.Sys)
		if err != nil {
			return 0, 0, err
		}
	}

	siEx := new(windows.StartupInfoEx)
	siEx.Flags = windows.STARTF_USESTDHANDLES

	pi := new(windows.ProcessInformation)

	// Need EXTENDED_STARTUPINFO_PRESENT as we're making use of the attribute list field.
	flags := uint32(windows.CREATE_UNICODE_ENVIRONMENT) | windows.EXTENDED_STARTUPINFO_PRESENT
	if attr.Sys != nil && attr.Sys.CreationFlags != 0 {
		flags |= attr.Sys.CreationFlags
	}

	pSec := &windows.SecurityAttributes{Length: uint32(unsafe.Sizeof(zeroSec)), InheritHandle: 1}
	if attr.Sys != nil && attr.Sys.ProcessAttributes != nil {
		pSec = &windows.SecurityAttributes{
			Length:        attr.Sys.ProcessAttributes.Length,
			InheritHandle: attr.Sys.ProcessAttributes.InheritHandle,
		}
	}
	tSec := &windows.SecurityAttributes{Length: uint32(unsafe.Sizeof(zeroSec)), InheritHandle: 1}
	if attr.Sys != nil && attr.Sys.ThreadAttributes != nil {
		tSec = &windows.SecurityAttributes{
			Length:        attr.Sys.ThreadAttributes.Length,
			InheritHandle: attr.Sys.ThreadAttributes.InheritHandle,
		}
	}

	siEx.ProcThreadAttributeList = c.attrList.List() //nolint:govet // unusedwrite: ProcThreadAttributeList will be read in syscall
	siEx.Cb = uint32(unsafe.Sizeof(*siEx))
	if attr.Sys != nil && attr.Sys.Token != 0 {
		err = windows.CreateProcessAsUser(
			windows.Token(attr.Sys.Token),
			argv0p,
			argvp,
			pSec,
			tSec,
			false,
			flags,
			createEnvBlock(addCriticalEnv(dedupEnvCase(true, attr.Env))),
			dirp,
			&siEx.StartupInfo,
			pi,
		)
	} else {
		err = windows.CreateProcess(
			argv0p,
			argvp,
			pSec,
			tSec,
			false,
			flags,
			createEnvBlock(addCriticalEnv(dedupEnvCase(true, attr.Env))),
			dirp,
			&siEx.StartupInfo,
			pi,
		)
	}
	if err != nil {
		return 0, 0, fmt.Errorf("failed to create process: %w", err)
	}

	defer windows.CloseHandle(pi.Thread)

	return int(pi.ProcessId), uintptr(pi.Process), nil
}
