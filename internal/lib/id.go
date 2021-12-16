//
// mimixbox/internal/lib/id.go
//
// Copyright 2021 Naohiro CHIKAMATSU
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
package mb

import (
	"os/user"
	"strconv"
)

func LookupGid(groupId string) (int, error) {
	group, err := user.LookupGroupId(groupId)
	if err != nil {
		group, err = user.LookupGroup(groupId)
		if err != nil {
			return 0, err
		}
	}

	gid, err := strconv.Atoi(group.Gid)
	if err != nil {
		return 0, err
	}
	return gid, nil
}

func LookupUid(userId string) (int, error) {
	u, err := user.LookupId(userId)
	if err != nil {
		u, err = user.Lookup(userId)
		if err != nil {
			return 0, err
		}
	}
	uid, err := strconv.Atoi(u.Uid)
	if err != nil {
		return 0, err
	}
	return uid, nil
}
