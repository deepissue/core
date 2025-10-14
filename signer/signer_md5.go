package signer

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/deepissue/core/utils"
)

type signer struct {
}

func (m *signer) build(req any) (*bytes.Buffer, error) {
	params := make(map[string]any)
	data, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(data, &params); err != nil {
		return nil, err
	}

	var keys []string
	for k, _ := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var buf bytes.Buffer
	for _, k := range keys {
		if strings.ToLower(k) == "sign" {
			continue
		}
		if utils.IsEmpty(params[k]) {
			continue
		}
		buf.WriteString(fmt.Sprintf("%v", params[k]))
	}

	return &buf, nil
}

type md5Signer struct {
	signer
}

func NewMD5Signer() *md5Signer {
	return &md5Signer{}
}

// Sign 签名方法
// keyId keys id
func (m *md5Signer) Sign(key string, resp any) (string, error) {
	buf, err := m.build(resp)
	if err != nil {
		return "", err
	}

	buf.WriteString(key)
	fmt.Println(buf.String())
	sign := m.md5(buf.Bytes())
	return sign, nil
}

// Verify 签名校验方法
func (m *md5Signer) Verify(key, sign string, req any) error {

	buf, err := m.build(req)
	if err != nil {
		return err
	}

	buf.WriteString(key)
	newSign := m.md5(buf.Bytes())
	if newSign != sign {
		return errInvalidSign
	}
	return nil
}

func (m *md5Signer) md5(data []byte) string {
	md := md5.New()
	md.Write(data)
	return hex.EncodeToString(md.Sum(nil))
}
