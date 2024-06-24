//go:build windows
// +build windows

package conpty

import (
	"fmt"
	"io"
	"os"
	"sync"
	"syscall"
	"unsafe"

	"github.com/charmbracelet/x/errors"
	"golang.org/x/sys/windows"
)

// Default size.
const (
	DefaultWidth  = 80
	DefaultHeight = 25
)

// ConPty represents a Windows Console Pseudo-terminal.
// https://learn.microsoft.com/en-us/windows/console/creating-a-pseudoconsole-session#preparing-the-communication-channels
type ConPty struct {
	hpc                 *windows.Handle
	inPipeFd, outPipeFd windows.Handle
	inPipe, outPipe     *os.File
	attrList            *windows.ProcThreadAttributeListContainer
	size                windows.Coord
	closeOnce           sync.Once
}

var (
	_ io.Writer = &ConPty{}
	_ io.Reader = &ConPty{}
)

// New creates a new ConPty device.
// Accepts a custom width, height, and flags that will get passed to
// windows.CreatePseudoConsole.
func New(w int, h int, flags int) (c *ConPty, err error) {
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
	}

	var ptyIn, ptyOut windows.Handle
	if err := windows.CreatePipe(&ptyIn, &c.inPipeFd, nil, 0); err != nil {
		return nil, fmt.Errorf("failed to create pipes for pseudo console: %w", err)
	}

	if err := windows.CreatePipe(&c.outPipeFd, &ptyOut, nil, 0); err != nil {
		return nil, fmt.Errorf("failed to create pipes for pseudo console: %w", err)
	}

	if err := windows.CreatePseudoConsole(c.size, ptyIn, ptyOut, uint32(flags), c.hpc); err != nil {
		return nil, fmt.Errorf("failed to create pseudo console: %w", err)
	}

	// We don't need the pty pipes anymore, these will get dup'd when the
	// new process starts.
	if err := windows.CloseHandle(ptyOut); err != nil {
		return nil, fmt.Errorf("failed to close pseudo console handle: %w", err)
	}
	if err := windows.CloseHandle(ptyIn); err != nil {
		return nil, fmt.Errorf("failed to close pseudo console handle: %w", err)
	}

	c.inPipe = os.NewFile(uintptr(c.inPipeFd), "|0")
	c.outPipe = os.NewFile(uintptr(c.outPipeFd), "|1")

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

	return
}

// Fd returns the ConPty handle.
func (p *ConPty) Fd() uintptr {
	return uintptr(*p.hpc)
}

// Close closes the ConPty device.
func (p *ConPty) Close() error {
	var err error
	p.closeOnce.Do(func() {
		if p.attrList != nil {
			p.attrList.Delete()
		}
		windows.ClosePseudoConsole(*p.hpc)
		err = errors.Join(p.inPipe.Close(), p.outPipe.Close())
	})
	return err
}

// InPipe returns the ConPty input pipe.
func (p *ConPty) InPipe() *os.File {
	return p.inPipe
}

// InPipeFd returns the ConPty input pipe file descriptor handle.
func (p *ConPty) InPipeFd() uintptr {
	return uintptr(p.inPipeFd)
}

// OutPipe returns the ConPty output pipe.
func (p *ConPty) OutPipe() *os.File {
	return p.outPipe
}

// OutPipeFd returns the ConPty output pipe file descriptor handle.
func (p *ConPty) OutPipeFd() uintptr {
	return uintptr(p.outPipeFd)
}

// Write safely writes bytes to the ConPty.
func (c *ConPty) Write(p []byte) (n int, err error) {
	var l uint32
	err = windows.WriteFile(c.inPipeFd, p, &l, nil)
	return int(l), err
}

// Read safely reads bytes from the ConPty.
func (c *ConPty) Read(p []byte) (n int, err error) {
	var l uint32
	err = windows.ReadFile(c.outPipeFd, p, &l, nil)
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
	return
}

var zeroAttr syscall.ProcAttr

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

	var zeroSec windows.SecurityAttributes
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
