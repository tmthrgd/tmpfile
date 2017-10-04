// Copyright 2017 Tom Thorogood. All rights reserved.
// Use of this source code is governed by a Modified
// BSD License that can be found in the LICENSE file.

// +build linux

package tmpfile

import (
	"io/ioutil"
	"os"
	"runtime"
	"strconv"
	"syscall"

	"github.com/tmthrgd/atomics"
	"golang.org/x/sys/unix"
)

var missingTMPFILE atomics.Bool

// TempFile creates a new temporary file in the directory dir, opens
// the file for reading and writing, and returns the resulting
// *os.File.
//
// If dir is the empty string, TempFile uses the default directory
// for temporary files (see os.TempDir).
//
// Multiple programs calling TempFile simultaneously will not choose
// the same file.
//
// If remove is true, it is the caller's responsibility to remove the
// file when no longer needed. In that case, the caller can use
// f.Name() to find the pathname of the file. This will be true, if
// the kernel or filesystem does not support O_TMPFILE. In this case,
// ioutil.TempFile is used as a fallback,
func TempFile(dir string) (f *os.File, remove bool, err error) {
	if dir == "" {
		dir = os.TempDir()
	}

	if missingTMPFILE.Load() {
		f, err := ioutil.TempFile(dir, "")
		return f, err == nil, err
	}

	fd, err := unix.Open(dir, unix.O_RDWR|unix.O_TMPFILE|unix.O_CLOEXEC, 0600)

	switch err {
	case nil:
	case syscall.EISDIR:
		missingTMPFILE.Set()
		fallthrough
	case syscall.EOPNOTSUPP:
		f, err := ioutil.TempFile(dir, "")
		return f, err == nil, err
	default:
		return nil, false, &os.PathError{
			Op:   "open",
			Path: dir,
			Err:  err,
		}
	}

	path := "/proc/self/fd/" + strconv.FormatUint(uint64(fd), 10)
	return os.NewFile(uintptr(fd), path), false, nil
}

// Link links the *os.File returned by TempFile into the filesystem
// at the given path, making it permanent.
//
// If TempFile was forced to fallback to ioutil.TempFile, this calls
// os.Rename with the file path.
//
// If f was not returned by TempFile, the behaviour of Link is
// undefined.
func Link(f *os.File, newpath string) error {
	if missingTMPFILE.Load() {
		return os.Rename(f.Name(), newpath)
	}

	r0, _, errno := unix.Syscall(unix.SYS_FCNTL, f.Fd(), unix.F_GETFL, 0)
	if errno != 0 {
		return &os.LinkError{
			Op:  "link",
			Old: f.Name(),
			New: newpath,
			Err: errno,
		}
	}

	if r0&unix.O_TMPFILE != unix.O_TMPFILE {
		return os.Rename(f.Name(), newpath)
	}

	err := unix.Linkat(unix.AT_FDCWD, f.Name(), unix.AT_FDCWD, newpath,
		unix.AT_SYMLINK_FOLLOW)
	if err != nil {
		return &os.LinkError{
			Op:  "link",
			Old: f.Name(),
			New: newpath,
			Err: err,
		}
	}

	runtime.KeepAlive(f)
	return nil
}
