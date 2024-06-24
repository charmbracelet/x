//go:build windows
// +build windows

package conpty

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"unicode/utf16"
	"unsafe"

	"golang.org/x/sys/windows"
)

// Below are a bunch of helpers for working with Windows' CreateProcess family
// of functions. These are mostly exact copies of the same utilities found in
// the go stdlib.

func lookExtensions(path, dir string) (string, error) {
	if filepath.Base(path) == path {
		path = filepath.Join(".", path)
	}

	if dir == "" {
		return exec.LookPath(path)
	}

	if filepath.VolumeName(path) != "" {
		return exec.LookPath(path)
	}

	if len(path) > 1 && os.IsPathSeparator(path[0]) {
		return exec.LookPath(path)
	}

	dirandpath := filepath.Join(dir, path)

	// We assume that LookPath will only add file extension.
	lp, err := exec.LookPath(dirandpath)
	if err != nil {
		return "", err
	}

	ext := strings.TrimPrefix(lp, dirandpath)

	return path + ext, nil
}

func execEnvDefault(sys *syscall.SysProcAttr) (env []string, err error) {
	if sys == nil || sys.Token == 0 {
		return syscall.Environ(), nil
	}

	var block *uint16
	err = windows.CreateEnvironmentBlock(&block, windows.Token(sys.Token), false)
	if err != nil {
		return nil, err
	}

	defer windows.DestroyEnvironmentBlock(block)
	blockp := uintptr(unsafe.Pointer(block))

	for {
		// find NUL terminator
		end := unsafe.Pointer(blockp)
		for *(*uint16)(end) != 0 {
			end = unsafe.Pointer(uintptr(end) + 2)
		}

		n := (uintptr(end) - uintptr(unsafe.Pointer(blockp))) / 2
		if n == 0 {
			// environment block ends with empty string
			break
		}

		entry := (*[(1 << 30) - 1]uint16)(unsafe.Pointer(blockp))[:n:n]
		env = append(env, string(utf16.Decode(entry)))
		blockp += 2 * (uintptr(len(entry)) + 1)
	}
	return
}

func isSlash(c uint8) bool {
	return c == '\\' || c == '/'
}

func normalizeDir(dir string) (name string, err error) {
	ndir, err := syscall.FullPath(dir)
	if err != nil {
		return "", err
	}
	if len(ndir) > 2 && isSlash(ndir[0]) && isSlash(ndir[1]) {
		// dir cannot have \\server\share\path form
		return "", syscall.EINVAL
	}
	return ndir, nil
}

func volToUpper(ch int) int {
	if 'a' <= ch && ch <= 'z' {
		ch += 'A' - 'a'
	}
	return ch
}

func joinExeDirAndFName(dir, p string) (name string, err error) {
	if len(p) == 0 {
		return "", syscall.EINVAL
	}
	if len(p) > 2 && isSlash(p[0]) && isSlash(p[1]) {
		// \\server\share\path form
		return p, nil
	}
	if len(p) > 1 && p[1] == ':' {
		// has drive letter
		if len(p) == 2 {
			return "", syscall.EINVAL
		}
		if isSlash(p[2]) {
			return p, nil
		} else {
			d, err := normalizeDir(dir)
			if err != nil {
				return "", err
			}
			if volToUpper(int(p[0])) == volToUpper(int(d[0])) {
				return syscall.FullPath(d + "\\" + p[2:])
			} else {
				return syscall.FullPath(p)
			}
		}
	} else {
		// no drive letter
		d, err := normalizeDir(dir)
		if err != nil {
			return "", err
		}
		if isSlash(p[0]) {
			return windows.FullPath(d[:2] + p)
		} else {
			return windows.FullPath(d + "\\" + p)
		}
	}
}

// createEnvBlock converts an array of environment strings into
// the representation required by CreateProcess: a sequence of NUL
// terminated strings followed by a nil.
// Last bytes are two UCS-2 NULs, or four NUL bytes.
func createEnvBlock(envv []string) *uint16 {
	if len(envv) == 0 {
		return &utf16.Encode([]rune("\x00\x00"))[0]
	}
	length := 0
	for _, s := range envv {
		length += len(s) + 1
	}
	length++

	b := make([]byte, length)
	i := 0
	for _, s := range envv {
		l := len(s)
		copy(b[i:i+l], []byte(s))
		copy(b[i+l:i+l+1], []byte{0})
		i = i + l + 1
	}
	copy(b[i:i+1], []byte{0})

	return &utf16.Encode([]rune(string(b)))[0]
}

// dedupEnvCase is dedupEnv with a case option for testing.
// If caseInsensitive is true, the case of keys is ignored.
func dedupEnvCase(caseInsensitive bool, env []string) []string {
	out := make([]string, 0, len(env))
	saw := make(map[string]int, len(env)) // key => index into out
	for _, kv := range env {
		eq := strings.Index(kv, "=")
		if eq < 0 {
			out = append(out, kv)
			continue
		}
		k := kv[:eq]
		if caseInsensitive {
			k = strings.ToLower(k)
		}
		if dupIdx, isDup := saw[k]; isDup {
			out[dupIdx] = kv
			continue
		}
		saw[k] = len(out)
		out = append(out, kv)
	}
	return out
}

// addCriticalEnv adds any critical environment variables that are required
// (or at least almost always required) on the operating system.
// Currently this is only used for Windows.
func addCriticalEnv(env []string) []string {
	for _, kv := range env {
		eq := strings.Index(kv, "=")
		if eq < 0 {
			continue
		}
		k := kv[:eq]
		if strings.EqualFold(k, "SYSTEMROOT") {
			// We already have it.
			return env
		}
	}
	return append(env, "SYSTEMROOT="+os.Getenv("SYSTEMROOT"))
}
