package auth

import "net/http"

type MockAuthenticator struct{}

const testSubjectHeader = "X-Test-Sub"
const testErrorHeader = "X-Test-Error"

func (a *MockAuthenticator) Authenticate(w http.ResponseWriter, r *http.Request) (string, error) {
	msg := r.Header.Get(testErrorHeader)
	if msg != "" {
		return "", ErrServer
	}
	sub := r.Header.Get(testSubjectHeader)
	if sub == "" {
		return "", ErrUnauthenticated
	}
	return sub, nil
}

func NewMockAuthenticator() Authenticator {
	return &MockAuthenticator{}
}

func MockLogin(r *http.Request, subject string) {
	r.Header.Set(testSubjectHeader, subject)
}
