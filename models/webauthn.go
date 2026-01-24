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

// Value implements driver.Valuer.
//
// It stores only the raw protocol.AuthenticatorFlags value (Protocol Value),
// which is stable against future changes in how individual flags are interpreted.
func (f CredentialFlags) Value() (driver.Value, error) {
	return int64(f.toLib().ProtocolValue()), nil
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

	var v int64
	switch x := value.(type) {
	case int64:
		v = x
	case int32:
		v = int64(x)
	case int:
		v = int64(x)
	default:
		return fmt.Errorf("unsupported authenticator flags type %T", value)
	}

	// protocol.AuthenticatorFlags is a byte, so clamp to that range.
	if v < 0 || v > 0xFF {
		return fmt.Errorf("invalid authenticator flags value %d", v)
	}

	lib := webauthn.NewCredentialFlags(protocol.AuthenticatorFlags(v))
	*f = CredentialFlags(lib)
	return nil
}

type WebAuthnCredential struct {
	UserID          string                            `gorm:"index"`
	ID              []byte                            `gorm:"primaryKey;not null"`
	CreatedAt       CustomTime                        // filled by gorm, see https://gorm.io/docs/conventions.html#CreatedAt
	LastUsedAt      CustomTime                        ` gorm:"default:null"` // NOT filled by gorm
	Name            string                            `gorm:"not null"`
	PublicKey       []byte                            `gorm:"not null"`
	AttestationType string                            `gorm:"not null"`
	Transport       []protocol.AuthenticatorTransport `gorm:"serializer:json"`
	Flags           CredentialFlags
	Authenticator   webauthn.Authenticator         `gorm:"serializer:json"`
	Attestation     webauthn.CredentialAttestation `gorm:"serializer:json"`
}

func (c WebAuthnCredential) toLib() *webauthn.Credential {
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

func CredentialFromLib(lib *webauthn.Credential, userID string, name string) *WebAuthnCredential {
	return &WebAuthnCredential{
		Name:            name,
		UserID:          userID,
		ID:              lib.ID,
		PublicKey:       lib.PublicKey,
		AttestationType: lib.AttestationType,
		Transport:       lib.Transport,
		Flags:           CredentialFlags(lib.Flags),
		Authenticator:   lib.Authenticator,
		Attestation:     lib.Attestation,
	}
}

func CredentialFromLibWithNoUserData(lib *webauthn.Credential) *WebAuthnCredential {
	return &WebAuthnCredential{
		ID:              lib.ID,
		PublicKey:       lib.PublicKey,
		AttestationType: lib.AttestationType,
		Transport:       lib.Transport,
		Flags:           CredentialFlags(lib.Flags),
		Authenticator:   lib.Authenticator,
		Attestation:     lib.Attestation,
	}
}
