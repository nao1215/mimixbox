//
// mimixbox/internal/applets/pmutils/halt/wtmp.go
//
// Copyright 2021 Naohiro CHIKAMATSU, polynomialspace
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package halt

import (
	"encoding/binary"
	"os"
	"time"
)

// This file is the wtmp record encoding seam: it owns the Linux struct utmp
// layout and the append-to-wtmp side effect. The shutdown planner and reboot
// executor in planner.go drive it; behavior and the on-disk record are unchanged.

// utmpRecordSize is the size of a Linux struct utmp record.
const utmpRecordSize = 384

// runLevel is the utmp record type for a run-level (shutdown) entry.
const runLevel = 1

// encodeWtmpRecord serializes a "shutdown" login record, as sysvinit's halt
// writes, so utilities such as "who" and "last" can report the shutdown time.
// The record layout matches Linux's struct utmp.
func encodeWtmpRecord(now time.Time) []byte {
	rec := make([]byte, utmpRecordSize)
	binary.LittleEndian.PutUint16(rec[0:], runLevel)                        // ut_type
	copy(rec[8:40], "~~")                                                   // ut_line[32]
	copy(rec[40:44], "~~")                                                  // ut_id[4]
	copy(rec[44:76], "shutdown")                                            // ut_user[32]
	binary.LittleEndian.PutUint32(rec[340:], uint32(now.Unix()))            // ut_tv.tv_sec
	binary.LittleEndian.PutUint32(rec[344:], uint32(now.Nanosecond()/1000)) // ut_tv.tv_usec
	return rec
}

// writeWtmp appends a "shutdown" login record to path. The encoding is owned by
// encodeWtmpRecord; this function performs only the append side effect.
func writeWtmp(path string, now time.Time) error {
	rec := encodeWtmpRecord(now)

	f, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644) //nolint:gosec // wtmp is world-readable
	if err != nil {
		return err
	}
	defer func() {
		if cerr := f.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()
	_, err = f.Write(rec)
	return err
}
