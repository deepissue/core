// Copyright 2022 The Go Authors<36625090@qq.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// EncodeJSON Encodes/Marshals the given object into JSON
func EncodeJSON(in any) ([]byte, error) {
	if in == nil {
		return nil, fmt.Errorf("input for encoding is nil")
	}
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	if err := enc.Encode(in); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// JSONEncode Encodes/Marshals the given object into JSON
func JSONEncode(in any) []byte {
	if in == nil {
		return nil
	}
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	if err := enc.Encode(in); err != nil {
		return nil
	}
	return buf.Bytes()
}

// DecodeJSON tries to decompress the given data. The call to decompress, fails
// if the content was not compressed in the first place, which is identified by
// a canary byte before the compressed data. If the data is not compressed, it
// is JSON decoded directly. Otherwise, the decompressed data will be JSON
// decoded.
func DecodeJSON(data []byte, out any) error {
	if len(data) == 0 {
		return fmt.Errorf("'data' being decoded is nil")
	}
	if out == nil {
		return fmt.Errorf("output parameter 'out' is nil")
	}

	return DecodeJSONFromReader(bytes.NewReader(data), out)
}

// DecodeJSONFromReader Decodes/Unmarshals the given io.Reader pointing to a JSON, into a desired object
func DecodeJSONFromReader(r io.Reader, out any) error {
	if r == nil {
		return fmt.Errorf("'io.Reader' being decoded is nil")
	}
	if out == nil {
		return fmt.Errorf("output parameter 'out' is nil")
	}

	dec := json.NewDecoder(r)

	// While decoding JSON values, interpret the integer values as `json.Number`s instead of `float64`.
	dec.UseNumber()

	// Since 'out' is an interface representing a pointer, pass it to the decoder without an '&'
	return dec.Decode(out)
}

func EncodeToString(in any) string {
	out, err := EncodeJSON(in)
	if nil != err {
		return err.Error()
	}
	return strings.ReplaceAll(string(out), "\n", "")
}

func Swap(in any, out any) error {
	bs, err := EncodeJSON(in)
	if nil != err {
		return err
	}
	return DecodeJSON(bs, out)
}

func PrettyJSON(in any) string {
	data, _ := json.MarshalIndent(in, "", " ")
	return string(data)
}
