package authorities

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"testing"

	"github.com/deepissue/core/utils"
	"github.com/hashicorp/hcl"
)

var block = `
  auth_type         = "redis"
  pkcs8_private_key = "MIICeAIBADANBgkqhkiG9w0BAQEFAASCAmIwggJeAgEAAoGBAOVxpmJr4ELzX67oQl8YCrHPk61sRwESc8kAFDm9PwrY/Wd/PqBVsCQUFYBmo5dSukdJ/ZkeyXqA9pArnlqn/G42EVUjPPNURiex4W6LbSHXr/96Wt/0Ov7d+8ETkmLUZ+QsdB+9S6CrkG9pfhdUKLBoJ/YPujOhDBQvWNQSnXzXAgMBAAECgYAeTQ8LKnH4hYmaYMP7KQKojuBS49zQsG4oGmGRaoO73AJDO9O6evaDHT/lsChkoKFHLudV5HH5QrTNP2VvVYYJjAcslxVchQssuagplZtbjuixNPfv2ey9qPXafHMbdPZy97uZTZkaxQ0aMNpFOGKk/m5KOXTt8lhsZBKmpb9IqQJBAO72peFpUdCWW0Fvy4Xw9VSZq09EHHForxu6YHRu4sdAXoasLf8vmoIfHBsD87Tat01K6pxw1YaBhDry9Zkr4LMCQQD1zUKMoa9YVYDA3ty8R9DAmkYoguhAV3Sm2cf1jIF/p5kazja+L6c2BGk5sxM/AG/rLMS04vw4lPO8s2boPv1NAkEAj+Q3eKc5m7eeFaYi0HGK2Ll7vUxPMD8QCktNH29R4RcylDeDrwDUMfxXqTDVBBcbf1BYO4F6IfdFT1XTa7tPHwJBAImvDkYEE1ohmttueqqkd5RLVl0+5qWT123Ws6EhsTA2SxauyA9EVh913RNK8c7qicZr70t7kdiH5veeblhNYEkCQQDrSM+LzGB2CipariZdInt/Jkp5YVlPy6Xf8D6DUxmuSgYJSbuWrtP8dAeQuZ48gEuZZbsjjNw/ngfaXxnPHt/4"
  pkcs1_public_key  = "MIGJAoGBAOVxpmJr4ELzX67oQl8YCrHPk61sRwESc8kAFDm9PwrY/Wd/PqBVsCQUFYBmo5dSukdJ/ZkeyXqA9pArnlqn/G42EVUjPPNURiex4W6LbSHXr/96Wt/0Ov7d+8ETkmLUZ+QsdB+9S6CrkG9pfhdUKLBoJ/YPujOhDBQvWNQSnXzXAgMBAAE="
  anon_methods      = ["admin.account.login", "user.sms.verify_code", "user.account.wechat_mini",
                      "user.hotel.list","user.hotel.detail","user.address.list", "user.hotel.index"]
  default_policy    = "deny"
`

func TestNewAuthorized(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
		return
	}
	t.Log(key.Size())

	pri, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		t.Fatal(err)
		return
	}
	pub := x509.MarshalPKCS1PublicKey(&key.PublicKey)
	settings := Settings{
		PKCS8PrivateKey: base64.StdEncoding.EncodeToString(pri),
		PKCS1PublicKey:  base64.StdEncoding.EncodeToString(pub),
		//Timeout:         300000000000,
		//AuthType:        AuthTypeRedis,
	}
	t.Log(settings.Timeout)
	if err := hcl.Decode(&settings, block); err != nil {
		t.Fatal(err)
		return
	}

	handler, err := NewJwtTokenHandler("app", &settings)
	if err != nil {
		t.Fatal(err)
		return
	}
	t.Log(settings.Timeout)
	auth, err := NewAuthorization(&settings, handler)
	if err != nil {
		t.Fatal(err)
		return
	}

	authed := NewAuthorized("12321", "liping", map[string]any{}, []string{})
	authed.Principal["location"] = "shanghai"

	jwt, err := handler.GenerateToken(authed)
	if err != nil {
		t.Fatal(err)
		return
	}
	t.Log(jwt)
	authorized, err := auth.Authentication(context.Background(), "Bearer "+jwt)
	if err != nil {
		t.Fatal(err)
		return
	}
	t.Log(utils.EncodeToString(authorized))
}
