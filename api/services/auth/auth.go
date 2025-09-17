package auth

import (
	"errors"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

var ErrUnauthenticated = errors.New("unauthenticated")
var ErrServer = errors.New("server error")

type Authenticator interface {
	Authenticate(w http.ResponseWriter, r *http.Request) (string, error)
}

type ServiceAccountAuthenticator struct {
	accounts map[string][]byte
}

type ServiceAccount struct {
	ID         string
	SecretHash []byte
}

func (a *ServiceAccountAuthenticator) Authenticate(w http.ResponseWriter, r *http.Request) (string, error) {
	user, pass, ok := r.BasicAuth()
	if !ok {
		w.Header().Add("WWW-Authenticate", "Basic realm=api")
		return "", ErrUnauthenticated
	}
	hash, ok := a.accounts[user]
	if !ok {
		w.Header().Add("WWW-Authenticate", "Basic realm=api")
		return "", ErrUnauthenticated
	}
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(pass)); err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			w.Header().Add("WWW-Authenticate", "Basic realm=api")
			return "", ErrUnauthenticated
		}
		return "", ErrServer
	}
	return user, nil
}

type ServiceAccountAuthenticatorOptions struct {
	Accounts []ServiceAccount
}

func NewServiceAccountAuthenticator(o ServiceAccountAuthenticatorOptions) Authenticator {
	a := ServiceAccountAuthenticator{
		accounts: map[string][]byte{},
	}
	for _, account := range o.Accounts {
		a.accounts[account.ID] = account.SecretHash
	}
	return &a
}
