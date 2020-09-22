package auth

import (
	"crypto/sha256"
	"fmt"

	"github.com/MichaelMure/git-bug-migration/migration12/before/entity"
)

const (
	configKeyLoginPasswordLogin    = "login"
	configKeyLoginPasswordPassword = "password"
)

var _ Credential = &LoginPassword{}

type LoginPassword struct {
	*CredentialBase
	Login    string
	Password string
}

func NewLoginPassword(target, login, password string) *LoginPassword {
	return &LoginPassword{
		CredentialBase: newCredentialBase(target),
		Login:          login,
		Password:       password,
	}
}

func NewLoginPasswordFromConfig(conf map[string]string) (*LoginPassword, error) {
	base, err := newCredentialBaseFromConfig(conf)
	if err != nil {
		return nil, err
	}

	return &LoginPassword{
		CredentialBase: base,
		Login:          conf[configKeyLoginPasswordLogin],
		Password:       conf[configKeyLoginPasswordPassword],
	}, nil
}

func (lp *LoginPassword) ID() entity.Id {
	h := sha256.New()
	_, _ = h.Write(lp.SaltT)
	_, _ = h.Write([]byte(lp.TargetT))
	_, _ = h.Write([]byte(lp.Login))
	_, _ = h.Write([]byte(lp.Password))
	return entity.Id(fmt.Sprintf("%x", h.Sum(nil)))
}

func (lp *LoginPassword) Kind() CredentialKind {
	return KindLoginPassword
}

func (lp *LoginPassword) Validate() error {
	err := lp.CredentialBase.validate()
	if err != nil {
		return err
	}
	if lp.Login == "" {
		return fmt.Errorf("missing login")
	}
	if lp.Password == "" {
		return fmt.Errorf("missing password")
	}
	return nil
}

func (lp *LoginPassword) toConfig() map[string]string {
	return map[string]string{
		configKeyLoginPasswordLogin:    lp.Login,
		configKeyLoginPasswordPassword: lp.Password,
	}
}
