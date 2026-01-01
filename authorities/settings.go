package authorities

import "time"

type AuthorizationPolicy string

const (
	AuthorizationPolicyDeny  AuthorizationPolicy = "deny"
	AuthorizationPolicyAllow AuthorizationPolicy = "allow"
)

type AuthType string

const (
	AuthTypeJwt   AuthType = "jwt"
	AuthTypeRedis AuthType = "redis"
)

// Settings for the application authorization
type Settings struct {
	AuthType AuthType `json:"auth_type" hcl:"auth_type" toml:"auth_type"`

	//PKCS8 ciphertext block
	PKCS8PrivateKey string `hcl:"pkcs8_private_key" json:"pkcs8_private_key" toml:"pkcs8_private_key"`
	//PKCS1 ciphertext block
	PKCS1PublicKey string `hcl:"pkcs1_public_key" json:"pkcs1_public_key" toml:"pkcs1_public_key"`

	//Timeout token timeout
	Timeout time.Duration `hcl:"timeout" json:"timeout" toml:"timeout" default:"24"`

	//AnonEndpoints anonymous endpoints
	AnonEndpoints []string `hcl:"anon_endpoints" json:"anon_endpoints" toml:"anon_endpoints"`

	//Disabled
	DefaultPolicy AuthorizationPolicy `hcl:"default_policy" json:"default_policy" toml:"default_policy" default:"deny"`

	InternalSecret string `hcl:"internal_secret" json:"internal_secret" toml:"internal_secret"`
}
