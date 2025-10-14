package signer

import (
	"errors"
)

var errInvalidSign = errors.New("invalid signature")

type Signer interface {
	Sign(keyId string, resp any) (string, error)
	Verify(keyId, sign string, req any) error
}

var _md5Signer = NewMD5Signer()

func MD5KvSign(key string, data any) (string, error) {
	return _md5Signer.Sign(key, data)
}

func MD5KvVerify(key, sign string, data any) error {
	return _md5Signer.Verify(key, sign, data)
}
