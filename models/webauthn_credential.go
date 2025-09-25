package models

import (
	"encoding/json"
	"time"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
)

// WebAuthnCredential represents a single WebAuthn credential for a user
type WebAuthnCredential struct {
	ID              string                            `json:"id" gorm:"type:varchar(255)"`
	Name            string                            `json:"name"`
	CredentialID    []byte                            `json:"credential_id" gorm:"type:blob"`
	PublicKey       []byte                            `json:"public_key" gorm:"type:blob"`
	AttestationType string                            `json:"attestation_type"`
	Transport       []protocol.AuthenticatorTransport `json:"transport" gorm:"-"`
	TransportJSON   string                            `json:"-" gorm:"column:transport"`
	Flags           struct {
		UserPresent    bool `json:"user_present"`
		UserVerified   bool `json:"user_verified"`
		BackupEligible bool `json:"backup_eligible"`
		BackupState    bool `json:"backup_state"`
	} `json:"flags" gorm:"embedded;embeddedPrefix:flag_"`
	Authenticator struct {
		AAGUID       []byte `json:"aaguid" gorm:"type:blob"`
		SignCount    uint32 `json:"sign_count"`
		CloneWarning bool   `json:"clone_warning"`
	} `json:"authenticator" gorm:"embedded;embeddedPrefix:auth_"`
	CreatedAt CustomTime `json:"created_at" swaggertype:"string" format:"date" example:"2006-01-02 15:04:05.000"`
	UpdatedAt CustomTime `json:"updated_at" swaggertype:"string" format:"date" example:"2006-01-02 15:04:05.000"`
}

// User WebAuthn credentials - embedded in user model
type UserWebAuthn struct {
	CredentialsJSON string `json:"-" gorm:"column:webauthn_credentials;type:text"`
	credentials     []WebAuthnCredential
	SessionData     string    `json:"-" gorm:"column:webauthn_session;type:text"`
	SessionExpiry   time.Time `json:"-" gorm:"column:webauthn_session_expiry"`
}

// WebAuthnCredentials returns the user's WebAuthn credentials
func (u *User) WebAuthnCredentials() []webauthn.Credential {
	// Load credentials from JSON if not already loaded
	if len(u.webAuthnCredentials) == 0 && u.WebAuthn.CredentialsJSON != "" {
		json.Unmarshal([]byte(u.WebAuthn.CredentialsJSON), &u.webAuthnCredentials)
	}

	credentials := make([]webauthn.Credential, len(u.webAuthnCredentials))
	for i, cred := range u.webAuthnCredentials {
		// Unmarshal transport if needed
		if len(cred.Transport) == 0 && cred.TransportJSON != "" {
			var transportStrings []string
			if err := json.Unmarshal([]byte(cred.TransportJSON), &transportStrings); err == nil {
				cred.Transport = make([]protocol.AuthenticatorTransport, len(transportStrings))
				for i, t := range transportStrings {
					cred.Transport[i] = protocol.AuthenticatorTransport(t)
				}
			}
		}

		credentials[i] = webauthn.Credential{
			ID:              cred.CredentialID,
			PublicKey:       cred.PublicKey,
			AttestationType: cred.AttestationType,
			Transport:       cred.Transport,
			Flags: webauthn.CredentialFlags{
				UserPresent:    cred.Flags.UserPresent,
				UserVerified:   cred.Flags.UserVerified,
				BackupEligible: cred.Flags.BackupEligible,
				BackupState:    cred.Flags.BackupState,
			},
			Authenticator: webauthn.Authenticator{
				AAGUID:       cred.Authenticator.AAGUID,
				SignCount:    cred.Authenticator.SignCount,
				CloneWarning: cred.Authenticator.CloneWarning,
			},
		}
	}
	return credentials
}

// WebAuthnID returns the user's ID as required by webauthn.User interface
func (u *User) WebAuthnID() []byte {
	return []byte(u.ID)
}

// WebAuthnName returns the user's username as required by webauthn.User interface
func (u *User) WebAuthnName() string {
	return u.ID
}

// WebAuthnDisplayName returns the user's display name as required by webauthn.User interface
func (u *User) WebAuthnDisplayName() string {
	return u.ID
}

// WebAuthnIcon returns the user's icon URL as required by webauthn.User interface
func (u *User) WebAuthnIcon() string {
	return ""
}

// CredentialExcludeList returns a list of existing credential IDs to exclude during registration
func (u *User) CredentialExcludeList() []protocol.CredentialDescriptor {
	// Load credentials from JSON if not already loaded
	if len(u.webAuthnCredentials) == 0 && u.WebAuthn.CredentialsJSON != "" {
		json.Unmarshal([]byte(u.WebAuthn.CredentialsJSON), &u.webAuthnCredentials)
	}

	credentialExcludeList := []protocol.CredentialDescriptor{}
	for _, cred := range u.webAuthnCredentials {
		// Unmarshal transport if needed
		if len(cred.Transport) == 0 && cred.TransportJSON != "" {
			var transportStrings []string
			if err := json.Unmarshal([]byte(cred.TransportJSON), &transportStrings); err == nil {
				cred.Transport = make([]protocol.AuthenticatorTransport, len(transportStrings))
				for i, t := range transportStrings {
					cred.Transport[i] = protocol.AuthenticatorTransport(t)
				}
			}
		}

		descriptor := protocol.CredentialDescriptor{
			Type:         protocol.PublicKeyCredentialType,
			CredentialID: cred.CredentialID,
			Transport:    cred.Transport,
		}
		credentialExcludeList = append(credentialExcludeList, descriptor)
	}
	return credentialExcludeList
}

// AddWebAuthnCredential adds a new WebAuthn credential to the user
func (u *User) AddWebAuthnCredential(credential *WebAuthnCredential) {	
	// Load existing credentials
	if len(u.webAuthnCredentials) == 0 && u.WebAuthn.CredentialsJSON != "" {
		json.Unmarshal([]byte(u.WebAuthn.CredentialsJSON), &u.webAuthnCredentials)
	}

	// Marshal transport
	if len(credential.Transport) > 0 {
		transportStrings := make([]string, len(credential.Transport))
		for i, t := range credential.Transport {
			transportStrings[i] = string(t)
		}
		transportBytes, _ := json.Marshal(transportStrings)
		credential.TransportJSON = string(transportBytes)
	}

	// Add to slice
	u.webAuthnCredentials = append(u.webAuthnCredentials, *credential)

	// Marshal back to JSON
	credentialsBytes, _ := json.Marshal(u.webAuthnCredentials)
	u.WebAuthn.CredentialsJSON = string(credentialsBytes)
}

// RemoveWebAuthnCredential removes a WebAuthn credential by credential ID
func (u *User) RemoveWebAuthnCredential(credentialID []byte) bool {
	// Load existing credentials
	if len(u.webAuthnCredentials) == 0 && u.WebAuthn.CredentialsJSON != "" {
		json.Unmarshal([]byte(u.WebAuthn.CredentialsJSON), &u.webAuthnCredentials)
	}

	// Find and remove credential
	for i, cred := range u.webAuthnCredentials {
		if string(cred.CredentialID) == string(credentialID) {
			u.webAuthnCredentials = append(u.webAuthnCredentials[:i], u.webAuthnCredentials[i+1:]...)

			// Marshal back to JSON
			credentialsBytes, _ := json.Marshal(u.webAuthnCredentials)
			u.WebAuthn.CredentialsJSON = string(credentialsBytes)
			return true
		}
	}
	return false
}

// UpdateWebAuthnCredential updates a WebAuthn credential
func (u *User) UpdateWebAuthnCredential(credentialID []byte, updatedCred *WebAuthnCredential) bool {
	// Load existing credentials
	if len(u.webAuthnCredentials) == 0 && u.WebAuthn.CredentialsJSON != "" {
		json.Unmarshal([]byte(u.WebAuthn.CredentialsJSON), &u.webAuthnCredentials)
	}

	// Find and update credential
	for i, cred := range u.webAuthnCredentials {
		if string(cred.CredentialID) == string(credentialID) {
			// Marshal transport
			if len(updatedCred.Transport) > 0 {
				transportBytes, _ := json.Marshal(updatedCred.Transport)
				updatedCred.TransportJSON = string(transportBytes)
			}

			u.webAuthnCredentials[i] = *updatedCred

			// Marshal back to JSON
			credentialsBytes, _ := json.Marshal(u.webAuthnCredentials)
			u.WebAuthn.CredentialsJSON = string(credentialsBytes)
			return true
		}
	}
	return false
}

// GetWebAuthnCredentials returns the raw credential slice for management
func (u *User) GetWebAuthnCredentials() []WebAuthnCredential {
	// Load credentials from JSON if not already loaded
	if len(u.webAuthnCredentials) == 0 && u.WebAuthn.CredentialsJSON != "" {
		json.Unmarshal([]byte(u.WebAuthn.CredentialsJSON), &u.webAuthnCredentials)
	}
	return u.webAuthnCredentials
}

// HasWebAuthnCredentials returns true if the user has any WebAuthn credentials
func (u *User) HasWebAuthnCredentials() bool {
	return u.WebAuthn.CredentialsJSON != "" && u.WebAuthn.CredentialsJSON != "[]"
}
