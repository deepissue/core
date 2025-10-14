package authorities

type noopTokenHandler struct {
}

func (n noopTokenHandler) GenerateToken(auth *Authorized) (string, error) {
	return "", nil
}

func (n noopTokenHandler) ParseToken(token string) (*Authorized, error) {
	return &Authorized{}, nil
}

func NewNoopTokenHandler() (TokenHandler, error) {
	return &noopTokenHandler{}, nil
}
