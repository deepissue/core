// Copyright 2022 The Go Authors<36625090@qq.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package utils

import (
	"crypto/md5"
	"fmt"
)

func MD5(b []byte) string {
	m := md5.New()
	m.Write(b)
	return fmt.Sprintf("%02x", m.Sum(nil))
}
