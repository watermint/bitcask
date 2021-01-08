// +build windows

package flock

import (
	"golang.org/x/sys/windows"
	"math"
	"os"
)

func lock_sys(path string, nonBlocking bool) (fh *os.File, err error) {
	fh, err = os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			_ = fh.Close()
		}
	}()

	var flag uint32
	flag = windows.LOCKFILE_EXCLUSIVE_LOCK
	if nonBlocking {
		flag |= windows.LOCKFILE_FAIL_IMMEDIATELY
	}

	err = windows.LockFileEx(windows.Handle(fh.Fd()), flag, 0, math.MaxUint32, math.MaxUint32, new(windows.Overlapped))
	if err != nil {
		return nil, err
	}

	return fh, nil
}

func unlock_sys(fh *os.File, path string) (err error) {
	if !sameFileId(fh, path) {
		return ErrInodeChangedAtPath
	}

	err = windows.UnlockFileEx(windows.Handle(fh.Fd()), 0, math.MaxUint32, math.MaxUint32, new(windows.Overlapped))
	_ = fh.Close()
	if err != nil {
		return err
	}

	//err = os.Remove(path)
	//if err != nil {
	//	return err
	//}
	return nil
}

func sameFileId(fh *os.File, path string) bool {
	var data windows.ByHandleFileInformation
	if err := windows.GetFileInformationByHandle(windows.Handle(fh.Fd()), &data); err != nil {
		return false
	}

	fh2, err := os.OpenFile(path, os.O_RDWR, 0666)
	if err != nil {
		return false
	}
	defer func() {
		_ = fh2.Close()
	}()
	var data2 windows.ByHandleFileInformation
	if err := windows.GetFileInformationByHandle(windows.Handle(fh2.Fd()), &data2); err != nil {
		return false
	}

	return data.FileIndexLow == data2.FileIndexLow &&
		data.FileIndexHigh == data2.FileIndexHigh
}
