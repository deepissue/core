package server

import (
	"errors"
	"slices"
	"strings"

	"github.com/deepissue/core/authorities"
)

const ClientIDKey = "X-Client-ID"
const AuthorizationKey = "Authorization"
const InternalSecretKey = "X-Internal-Secret"

func (m *HttpServer) Authorization(ctx *Context) error {
	if nil == m.authorization {
		m.logger.Warn("Validation interface was called, but the validator component is nil")
		return nil
	}
	endpoint := strings.TrimPrefix(ctx.Request.URL.Path, m.path)
	contained := slices.Contains(m.authorization.Settings().AnonEndpoints, endpoint)
	m.logger.Debug("auth", "path", endpoint, "contained", contained)
	if contained {
		return nil
	}
	if m.authorization.Settings().DefaultPolicy == authorities.AuthorizationPolicyAllow {
		return nil
	}
	token := ctx.GetHeader(AuthorizationKey)
	if "" == token {
		return errors.New("authorization token required")
	}

	authorized, err := m.authorization.Authentication(ctx, token)
	if nil != err {
		return errors.New("invalid token")
	}
	ctx.Authorized = authorized
	return nil
}

// Permission unused
func (m *HttpServer) Permission(endpoint string, authorized *authorities.Authorized) bool {
	if slices.Contains(m.authorization.Settings().AnonEndpoints, endpoint) {
		return true
	}
	return authorized.HasPermissions(endpoint)
}
