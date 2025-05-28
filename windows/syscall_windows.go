package windows

import "golang.org/x/sys/windows"

// NewLazySystemDLL is a helper function to create a LazyDLL for the system
// DLLs.
var NewLazySystemDLL = windows.NewLazySystemDLL

// Handle is a type alias for windows.Handle, which represents a Windows
// handle.
type Handle = windows.Handle

//sys	ReadConsoleInput(console Handle, buf *InputRecord, toread uint32, read *uint32) (err error) = kernel32.ReadConsoleInputW
//sys	PeekConsoleInput(console Handle, buf *InputRecord, toread uint32, read *uint32) (err error) = kernel32.PeekConsoleInputW
//sys	GetNumberOfConsoleInputEvents(console Handle, numevents *uint32) (err error) = kernel32.GetNumberOfConsoleInputEvents
//sys	FlushConsoleInputBuffer(console Handle) (err error) = kernel32.FlushConsoleInputBuffer
