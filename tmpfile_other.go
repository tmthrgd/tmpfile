// Copyright 2017 Tom Thorogood. All rights reserved.
// Use of this source code is governed by a Modified
// BSD License that can be found in the LICENSE file.

// +build !linux

package tmpfile

import (
	"io/ioutil"
	"os"
)

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
	f, err = ioutil.TempFile(dir, "")
	return f, err == nil, err
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
	return os.Rename(f.Name(), newpath)
}
