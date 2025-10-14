/*
 * Copyright 2022 The Go Authors<36625090@qq.com>. All rights reserved.
 * Use of this source code is governed by a MIT-style
 * license that can be found in the LICENSE file.
 */

package utils

import "testing"

func TestNewID(t *testing.T) {
	for i := 0; i < 10; i++ {
		t.Log(NewID(), len(NewID()))
	}
}
