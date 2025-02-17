package services

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/muety/wakapi/helpers"
	"github.com/muety/wakapi/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// OTPService handles OTP-related operations

type OTPService struct {
	db *gorm.DB
}

func NewOTPService(db *gorm.DB) *OTPService {
	return &OTPService{db: db}
}

func GenerateOTPHash() (string, string, error) {
	buffer := make([]byte, 3)
	if _, err := rand.Read(buffer); err != nil {
		return "", "", err
	}

	// Convert to 6 digits
	number := int(buffer[0])<<16 | int(buffer[1])<<8 | int(buffer[2])
	pin := fmt.Sprintf("%06d", number%1000000)

	// Hash the PIN
	hashedPin, err := bcrypt.GenerateFromPassword([]byte(pin), bcrypt.DefaultCost)
	if err != nil {
		return "", "", err
	}

	OTPHash := string(hashedPin)
	return pin, OTPHash, nil
}

// getOrCreateUser retrieves existing user or creates new one
func (s *OTPService) getOrCreateUser(email string) (*models.User, error) {
	var user models.User

	err := s.db.Where("email = ?", email).First(&user).Error
	if err == nil {
		return &user, nil
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// Create new user with UUID
	user = models.User{
		ID:    uuid.New().String(), // Ensure you import "github.com/google/uuid"
		Email: email,
	}

	if err := s.db.Create(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

// CreateOTP creates a new OTP for a user
func (s *OTPService) CreateOTP(otpRequest models.InitiateOTPRequest) (*models.CreateOTPResponse, error) {
	// Get or create user
	_, err := s.getOrCreateUser(otpRequest.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to process user: %w", err)
	}

	// Check for existing non-expired OTP
	var existingOTP models.OTP
	now := time.Now().Unix()

	err = s.db.Where("email = ? AND expires_in > ? AND used = ?",
		otpRequest.Email, now, false).First(&existingOTP).Error

	if err == nil {
		return nil, errors.New("an active OTP already exists")
	}

	otpText, otpHash, err := GenerateOTPHash()
	if err != nil {
		return nil, err
	}

	// Delete OTPs that are either expired or belong to this email
	if err := s.db.Where(
		"email = ? OR expires_in < ?",
		otpRequest.Email,
		time.Now().Unix(),
	).Delete(&models.OTP{}).Error; err != nil {
		fmt.Println(fmt.Errorf("failed to delete OTPs: %w", err)) // this shouldn't fail the request
	}

	// Create new OTP instance
	otp := models.OTP{
		Email:           otpRequest.Email,
		CodeChallenge:   otpRequest.CodeChallenge,
		ChallengeMethod: otpRequest.ChallengeMethod,
		ExpiresIn:       time.Now().Add(3 * time.Minute).Unix(),
		Used:            false,
		OTPHash:         otpHash,
	}

	if err := s.db.Create(&otp).Error; err != nil {
		return nil, err
	}

	// invalidate existing OTPs for the same user

	// send OTP via email
	fmt.Println("Sending OTP:", otpText)

	return &models.CreateOTPResponse{
		Message: "OTP created successfully",
	}, nil
}

func verifyPKCEChallenge(codeVerifier, codeChallenge, method string) bool {
	method = strings.ToLower(method) // Normalize method case

	switch method {
	case "plain":
		return subtle.ConstantTimeCompare([]byte(codeVerifier), []byte(codeChallenge)) == 1
	case "s256":
		h := sha256.New()
		h.Write([]byte(codeVerifier))
		challenge := base64.RawURLEncoding.EncodeToString(h.Sum(nil)) // Correct encoding
		return subtle.ConstantTimeCompare([]byte(challenge), []byte(codeChallenge)) == 1
	default:
		return false
	}
}

// VerifyOTP verifies the provided OTP
func (s *OTPService) VerifyOTP(validateOtpRequest models.ValidateOTPRequest) (*models.VerifyOTPResponse, error) {
	var otp models.OTP
	now := time.Now().Unix()

	// Find any pending OTP that matches and hasn't expired
	err := s.db.Where("expires_in > ? AND used = ? AND email = ?",
		now, false, validateOtpRequest.Email).Order("created_at DESC").First(&otp).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &models.VerifyOTPResponse{
				Message: "Invalid or expired OTP",
				Valid:   false,
			}, nil
		}
		return nil, err
	}

	// Verify PIN
	err = bcrypt.CompareHashAndPassword([]byte(otp.OTPHash), []byte(validateOtpRequest.OTP))
	if err != nil {
		return &models.VerifyOTPResponse{
			Message: "Invalid OTP",
			Valid:   false,
		}, nil
	}

	// Verify PKCE challenge
	if !verifyPKCEChallenge(validateOtpRequest.CodeVerifier, otp.CodeChallenge, otp.ChallengeMethod) {
		return &models.VerifyOTPResponse{
			Message: "Invalid PKCE challenge",
			Valid:   false,
		}, nil
	}

	// Mark OTP as used
	otp.Used = true
	if err := s.db.Save(&otp).Error; err != nil {
		return nil, err
	}

	return &models.VerifyOTPResponse{
		Message: "OTP verified successfully",
		Valid:   true,
	}, nil
}

// HTTP Handlers
func CreateOTPHandler(service *OTPService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var otpRequest models.InitiateOTPRequest
		if err := json.NewDecoder(r.Body).Decode(&otpRequest); err != nil {
			helpers.RespondJSON(w, r, http.StatusBadRequest, map[string]interface{}{
				"message": "Invalid request",
				"status":  http.StatusBadRequest,
			})
			return
		}

		if otpRequest.Email == "" {
			helpers.RespondJSON(w, r, http.StatusBadRequest, map[string]interface{}{
				"message": "Email is Required",
				"status":  http.StatusBadRequest,
			})
			return
		}

		if otpRequest.CodeChallenge == "" {
			helpers.RespondJSON(w, r, http.StatusBadRequest, map[string]interface{}{
				"message": "code_challenge is Required",
				"status":  http.StatusBadRequest,
			})
			return
		}

		if otpRequest.ChallengeMethod == "" {
			helpers.RespondJSON(w, r, http.StatusBadRequest, map[string]interface{}{
				"message": "challenge_method is Required",
				"status":  http.StatusBadRequest,
			})
			return
		}

		if otpRequest.ChallengeMethod != "S256" && otpRequest.ChallengeMethod != "plain" {
			helpers.RespondJSON(w, r, http.StatusBadRequest, map[string]interface{}{
				"message": "challenge_method must be either S256 or plain",
				"status":  http.StatusBadRequest,
			})
			return
		}

		resp, err := service.CreateOTP(otpRequest)
		if err != nil {
			fmt.Println("Error creating OTP", err)
			helpers.RespondJSON(w, r, http.StatusBadRequest, map[string]interface{}{
				"message": err.Error(),
				"status":  http.StatusBadRequest,
			})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}

func VerifyOTPHandler(service *OTPService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var validateOtpRequest models.ValidateOTPRequest
		if err := json.NewDecoder(r.Body).Decode(&validateOtpRequest); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Validate request
		if validateOtpRequest.Email == "" || validateOtpRequest.OTP == "" || validateOtpRequest.CodeVerifier == "" {
			helpers.RespondJSON(w, r, http.StatusBadRequest, map[string]interface{}{
				"message": "Email, OTP, and code_verifier are required",
				"status":  http.StatusBadRequest,
			})
			return
		}

		resp, err := service.VerifyOTP(validateOtpRequest)
		if err != nil {
			fmt.Println("Error creating OTP", err)
			helpers.RespondJSON(w, r, http.StatusBadRequest, map[string]interface{}{
				"message": err.Error(),
				"status":  http.StatusBadRequest,
			})
			return
		}

		// generate login token

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}
