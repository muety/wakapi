// ai generated below
package mocks

import (
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/muety/wakapi/models"
	"github.com/stretchr/testify/mock"
)

type WebAuthnServiceMock struct {
	mock.Mock
}

func (m *WebAuthnServiceMock) CreateCredential(c *webauthn.Credential, u *models.User, name string) (*models.WebAuthnCredential, error) {
	args := m.Called(c, u, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.WebAuthnCredential), args.Error(1)
}

func (m *WebAuthnServiceMock) GetCredentialsByUser(u *models.User) ([]*models.WebAuthnCredential, error) {
	args := m.Called(u)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.WebAuthnCredential), args.Error(1)
}

func (m *WebAuthnServiceMock) GetCredentialByUserAndName(u *models.User, name string) (*models.WebAuthnCredential, error) {
	args := m.Called(u, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.WebAuthnCredential), args.Error(1)
}

func (m *WebAuthnServiceMock) LoadCredentialIntoUser(u *models.User) error {
	args := m.Called(u)
	return args.Error(0)
}

func (m *WebAuthnServiceMock) DeleteCredential(c *models.WebAuthnCredential) error {
	args := m.Called(c)
	return args.Error(0)
}

func (m *WebAuthnServiceMock) UpdateCredential(c *webauthn.Credential) error {
	args := m.Called(c)
	return args.Error(0)
}
