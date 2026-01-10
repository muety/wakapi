package services

import (
	"github.com/go-webauthn/webauthn/webauthn"

	"github.com/muety/wakapi/models"
	"github.com/muety/wakapi/repositories"
)

type WebAuthnService struct {
	repository repositories.IWebAuthnRepository
}

func NewWebAuthnService(webAuthnRepository repositories.IWebAuthnRepository) *WebAuthnService {
	srv := &WebAuthnService{
		repository: webAuthnRepository,
	}

	return srv
}

func (srv *WebAuthnService) CreateCredential(credential *webauthn.Credential, user *models.User, name string) (*models.Credential, error) {
	createdCredential, err := srv.repository.Insert(models.CredentialFromLib(credential, user.ID, name))
	return createdCredential, err
}
func (srv *WebAuthnService) GetCredentialsByUser(user *models.User) ([]*models.Credential, error) {
	credentials, err := srv.repository.GetByUser(user.ID)
	return credentials, err
}

func (srv *WebAuthnService) GetCredentialByUserAndName(user *models.User, name string) (*models.Credential, error) {
	credentials, err := srv.repository.GetByUserAndName(user.ID, name)
	return credentials, err
}

func (srv *WebAuthnService) DeleteCredential(credential *models.Credential) error {
	return srv.repository.Delete(credential)
}

func (srv *WebAuthnService) UpdateCredential(credential *webauthn.Credential) error {
	return srv.repository.Update(models.CredentialFromLibWithNoUserData(credential))
}

func (srv *WebAuthnService) LoadCredentialIntoUser(user *models.User) error {
	credentials, err := srv.GetCredentialsByUser(user)
	if err != nil {
		return err
	}
	user.Credentials = credentials
	return nil
}
