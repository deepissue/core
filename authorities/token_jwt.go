package authorities

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	"strings"
	"time"
)

type Claims struct {
	jwt.RegisteredClaims
	Principal []byte
}

type jwtTokenHandler struct {
	pri      *rsa.PrivateKey
	pub      *rsa.PublicKey
	settings *Settings
	app      string
}

func NewJwtTokenHandler(app string, settings *Settings) (TokenHandler, error) {
	h := &jwtTokenHandler{
		app:      app,
		pri:      nil,
		pub:      nil,
		settings: settings,
	}

	block, err := base64.StdEncoding.DecodeString(settings.PKCS1PublicKey)
	if err != nil {
		return nil, fmt.Errorf("decoding public key: %v", err)
	}
	h.pub, err = x509.ParsePKCS1PublicKey(block)
	if err != nil {
		return nil, fmt.Errorf("parser public key: %v", err)
	}

	block, err = base64.StdEncoding.DecodeString(settings.PKCS8PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("decoding private key: %v", err)
	}
	pri, err := x509.ParsePKCS8PrivateKey(block)
	if err != nil {
		return nil, fmt.Errorf("parser PKCS8 private key: %v", err)
	}
	var ok bool
	h.pri, ok = pri.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("invalid PKCS8 private key")
	}

	return h, nil
}

// GenerateToken 产生token的函数
// 返回 Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9................................
func (m *jwtTokenHandler) GenerateToken(auth *Authorized) (string, error) {
	claims := Claims{}
	claims.ID = auth.ID.String()
	claims.Audience = jwt.ClaimStrings{auth.Account}
	claims.Issuer = m.app
	claims.Principal, _ = json.Marshal(auth)

	if m.settings.Timeout > 0 {
		claims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(m.settings.Timeout * time.Nanosecond))
	}

	tokenClaims := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err := tokenClaims.SignedString([]byte(m.settings.SimpleKey))
	if err != nil {
		return "", fmt.Errorf("jwt signing failed: %v", err)
	}

	return token, nil
}

// ParseToken
// 验证token的函数 Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9................................
func (m *jwtTokenHandler) parseToken(token string) (*Claims, error) {

	fields := strings.Split(token, " ")
	if len(fields) != 2 || fields[0] != "Bearer" {
		return nil, errors.New("invalid token, value must be: 'Bearer ......'")
	}

	tokenClaims, err := jwt.ParseWithClaims(fields[1], &Claims{}, func(token *jwt.Token) (interface{}, error) {
		token.Method.Alg()
		return []byte(m.settings.SimpleKey), nil
	})

	if err != nil {
		return nil, fmt.Errorf("token key with claims: %v", err)
	}

	if tokenClaims != nil {
		if claims, ok := tokenClaims.Claims.(*Claims); ok && tokenClaims.Valid {
			return claims, nil
		} else {
			return nil, tokenClaims.Claims.Valid()
		}
	}

	return nil, err
}

// ParseToken 解密验证信息
func (m *jwtTokenHandler) ParseToken(token string) (*Authorized, error) {
	claims, err := m.parseToken(token)
	if err != nil {
		return nil, err
	}

	var principal Principal
	if err := json.Unmarshal(claims.Principal, &principal); err != nil {
		return nil, err
	}

	authorized := &Authorized{
		ID:        ID(claims.ID),
		Principal: principal,
	}
	if len(claims.Audience) > 0 {
		authorized.Account = claims.Audience[0]
	}

	return authorized, nil
}
