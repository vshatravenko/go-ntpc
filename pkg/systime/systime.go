package systime

import (
	"syscall"
)

const (
	nanoPerSec = 1000000000
)

func UpdateSysTime(delta int64, apply bool) error {
	tv := &syscall.Timeval{}
	err := syscall.Gettimeofday(tv)
	if err != nil {
		return err
	}

	sec := delta / nanoPerSec
	usec := int32((delta - sec*nanoPerSec) / 1000)

	tv.Sec += sec
	tv.Usec += usec

	if apply {
		return syscall.Settimeofday(tv)
	}

	return nil
}
