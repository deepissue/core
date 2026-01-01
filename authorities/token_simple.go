package authorities

import "errors"

type simpleTokenHandler struct {
	token string
}

func (s *simpleTokenHandler) GenerateToken(auth *Authorized) (string, error) {
	return s.token, nil
}

func (s *simpleTokenHandler) ParseToken(token string) (*Authorized, error) {
	if token != s.token {
		return nil, errors.New("invalid token")
	}
	return &Authorized{}, nil
}

func NewSimpleTokenHandler(token string) (TokenHandler, error) {
	if token == "" {
		return nil, errors.New("token cannot be empty")
	}
	return &simpleTokenHandler{token: token}, nil
}
