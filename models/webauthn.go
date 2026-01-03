package models

import (
	"database/sql/driver"
	"fmt"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
)

// CredentialFlags is a database- and JSON-friendly wrapper around webauthn.CredentialFlags.
//
// For storage, we only persist the underlying protocol.AuthenticatorFlags "Protocol Value"
// as recommended by the go-webauthn documentation. The higher-level boolean flags should be
// reconstructed via webauthn.NewCredentialFlags when scanning from the database.
type CredentialFlags webauthn.CredentialFlags

// toLib converts the local wrapper type to the library type.
func (f CredentialFlags) toLib() webauthn.CredentialFlags {
	return webauthn.CredentialFlags(f)
}

// credentialFlagsFromLib converts the library type back to the local wrapper type.
func credentialFlagsFromLib(flags webauthn.CredentialFlags) CredentialFlags {
	return CredentialFlags(flags)
}

// Value implements driver.Valuer.
//
// It stores only the raw protocol.AuthenticatorFlags value (Protocol Value),
// which is stable against future changes in how individual flags are interpreted.
func (f CredentialFlags) Value() (driver.Value, error) {
	lib := f.toLib()
	raw := lib.ProtocolValue()
	return int64(raw), nil
}

// Scan implements sql.Scanner.
//
// It expects the stored value to be an integer representation of
// protocol.AuthenticatorFlags and reconstructs the higher-level
// webauthn.CredentialFlags via webauthn.NewCredentialFlags.
func (f *CredentialFlags) Scan(value interface{}) error {
	if value == nil {
		*f = CredentialFlags{}
		return nil
	}
	v := value.(int64)

	// protocol.AuthenticatorFlags is a byte, so clamp to that range.
	if v < 0 || v > 0xFF {
		return fmt.Errorf("invalid authenticator flags value %d", v)
	}

	lib := webauthn.NewCredentialFlags(protocol.AuthenticatorFlags(v))
	*f = credentialFlagsFromLib(lib)
	return nil
}

type Credential struct {
	UserID          string                            `json:"-" gorm:"index"`
	ID              []byte                            `json:"-" gorm:"primaryKey;not null"`
	CreatedAt       CustomTime                        `json:"created_at"`                       // filled by gorm, see https://gorm.io/docs/conventions.html#CreatedAt
	LastUsedAt      CustomTime                        `json:"last_used_at" gorm:"default:null"` // NOT filled by gorm
	Name            string                            `json:"-" gorm:"not null"`
	PublicKey       []byte                            `json:"-" gorm:"not null"`
	AttestationType string                            `json:"-" gorm:"not null"`
	Transport       []protocol.AuthenticatorTransport `json:"-" gorm:"serializer:json"`
	Flags           CredentialFlags                   `json:"-"`
	Authenticator   webauthn.Authenticator            `json:"-" gorm:"serializer:json"`
	Attestation     webauthn.CredentialAttestation    `json:"-" gorm:"serializer:json"`
}

func (c Credential) toLib() *webauthn.Credential {
	return &webauthn.Credential{
		ID:              c.ID,
		PublicKey:       c.PublicKey,
		AttestationType: c.AttestationType,
		Transport:       c.Transport,
		Flags:           c.Flags.toLib(),
		Authenticator:   c.Authenticator,
		Attestation:     c.Attestation,
	}
}

func CredentialFromLib(lib *webauthn.Credential, userID string, name string) *Credential {
	return &Credential{
		Name:            name,
		UserID:          userID,
		ID:              lib.ID,
		PublicKey:       lib.PublicKey,
		AttestationType: lib.AttestationType,
		Transport:       lib.Transport,
		Flags:           credentialFlagsFromLib(lib.Flags),
		Authenticator:   lib.Authenticator,
		Attestation:     lib.Attestation,
	}
}

func CredentialFromLibWithNoUserData(lib *webauthn.Credential) *Credential {
	return &Credential{
		ID:              lib.ID,
		PublicKey:       lib.PublicKey,
		AttestationType: lib.AttestationType,
		Transport:       lib.Transport,
		Flags:           credentialFlagsFromLib(lib.Flags),
		Authenticator:   lib.Authenticator,
		Attestation:     lib.Attestation,
	}
}
