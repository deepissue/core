// Copyright 2022 The Go Authors<36625090@qq.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package utils

import (
	"fmt"
	"math/rand"
	"net"
	"strings"
	"time"

	"github.com/google/uuid"
)

var hostId int64

func init() {
	addrs, _ := net.InterfaceAddrs()
	for _, addr := range addrs {
		a, _, err := net.ParseCIDR(addr.String())
		if err != nil {
			continue
		}
		if !a.IsLoopback() && a.IsPrivate() {
			hostId = InetAtoN(a.To4().String())
			break
		}
	}
}

// NewID 00xxxxxxx
func NewID() string {
	now := time.Now()
	rnd := rand.Int63n(now.UnixNano())
	number := fmt.Sprintf("%s%d%d", now.Format("20060102150405"), hostId, rnd)
	return number[:32]
}

func CleanedUUID() string {
	return strings.ReplaceAll(uuid.NewString(), "-", "")
}
