package auth

import "net/http"

type MockAuthenticator struct{}

const testSubjectHeader = "X-Test-Sub"
const testErrorHeader = "X-Test-Error"

func (a *MockAuthenticator) Authenticate(w http.ResponseWriter, r *http.Request) (string, error) {
	msg := r.Header.Get(testErrorHeader)
	if msg != "" {
		return "", ServerError
	}
	sub := r.Header.Get(testSubjectHeader)
	if sub == "" {
		return "", UnauthenticatedError
	}
	return sub, nil
}

func NewMockAuthenticator() Authenticator {
	return &MockAuthenticator{}
}

func MockLogin(r *http.Request, subject string) {
	r.Header.Set(testSubjectHeader, subject)
}
