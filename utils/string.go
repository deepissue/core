/*
 * Copyright 2022 The Go Authors<36625090@qq.com>. All rights reserved.
 * Use of this source code is governed by a MIT-style
 * license that can be found in the LICENSE file.
 */

package utils

import (
	"math/rand"
	"time"
)

const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
const numeric = "0123456789"

func RandomString(length int) string {
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	buf := make([]byte, length)
	for i := range length {
		buf[i] = chars[rnd.Intn(len(chars))]
	}
	return string(buf)
}

func RandomNumeric(length int) string {
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	buf := make([]byte, length)
	for i := range length {
		buf[i] = numeric[rnd.Intn(len(numeric))]
	}
	return string(buf)
}
