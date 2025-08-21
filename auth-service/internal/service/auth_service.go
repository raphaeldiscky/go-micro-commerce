// Package service provides business logic for the auth service.
package service

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-micro-template/pkg/logger"
	"github.com/raphaeldiscky/go-micro-template/pkg/mq"
	"golang.org/x/crypto/bcrypt"

	"github.com/raphaeldiscky/go-micro-template/auth-service/internal/config"
	"github.com/raphaeldiscky/go-micro-template/auth-service/internal/dto"
	"github.com/raphaeldiscky/go-micro-template/auth-service/internal/entity"
	"github.com/raphaeldiscky/go-micro-template/auth-service/internal/event"
	"github.com/raphaeldiscky/go-micro-template/auth-service/internal/httperror"
	"github.com/raphaeldiscky/go-micro-template/auth-service/internal/repository"
)

// AuthServiceInterface defines the methods for the auth service.
type AuthServiceInterface interface {
	Register(
		ctx context.Context,
		req *dto.RegisterRequest,
		clientIP, userAgent string,
	) (*dto.AuthResponse, error)
	Login(
		ctx context.Context,
		req *dto.LoginRequest,
		clientIP, userAgent string,
	) (*dto.AuthResponse, error)
	RefreshToken(ctx context.Context, req *dto.RefreshTokenRequest) (*dto.AuthResponse, error)
	Logout(ctx context.Context, req *dto.LogoutRequest) error
	LogoutAllSessions(ctx context.Context, userID uuid.UUID) error
	GetUser(ctx context.Context, userID uuid.UUID) (*dto.UserResponse, error)
	UpdateUser(
		ctx context.Context,
		userID uuid.UUID,
		req *dto.UpdateUserRequest,
	) (*dto.UserResponse, error)
	DeleteUser(ctx context.Context, userID uuid.UUID) error
	ChangePassword(ctx context.Context, userID uuid.UUID, req *dto.ChangePasswordRequest) error
	VerifyEmail(ctx context.Context, req *dto.VerifyEmailRequest) error
	ResendVerification(ctx context.Context, req *dto.ResendVerificationRequest) error
	GetActiveSessions(ctx context.Context, userID uuid.UUID) ([]*dto.SessionResponse, error)
}

// AuthService implements AuthServiceInterface.
type AuthService struct {
	dataStore                          repository.DataStore
	jwtConfig                          *config.JWTConfig
	logger                             logger.Logger
	emailVerificationRequestedProducer mq.KafkaProducerInterface
	userVerifiedProducer               mq.KafkaProducerInterface
}

// NewAuthService creates a new AuthService.
func NewAuthService(
	dataStore repository.DataStore,
	jwtConfig *config.JWTConfig,
	appLogger logger.Logger,
	emailVerificationRequestedProducer mq.KafkaProducerInterface,
	userVerifiedProducer mq.KafkaProducerInterface,
) AuthServiceInterface {
	return &AuthService{
		dataStore:                          dataStore,
		jwtConfig:                          jwtConfig,
		logger:                             appLogger,
		emailVerificationRequestedProducer: emailVerificationRequestedProducer,
		userVerifiedProducer:               userVerifiedProducer,
	}
}

// Register creates a new user account and returns authentication tokens.
func (s *AuthService) Register(
	ctx context.Context,
	req *dto.RegisterRequest,
	clientIP, userAgent string,
) (*dto.AuthResponse, error) {
	res := new(dto.AuthResponse)

	err := s.dataStore.Atomic(ctx, func(ds repository.DataStore) error {
		userRepo := ds.UserRepository()
		sessionRepo := ds.SessionRepository()
		// Check if email already exists
		emailExists, err := userRepo.EmailExists(ctx, req.Email)
		if err != nil {
			s.logger.Error("Failed to check email existence", "error", err)

			return err
		}

		if emailExists {
			return httperror.NewUserAlreadyExistError()
		}

		// Check if username already exists
		usernameExists, err := userRepo.UsernameExists(ctx, req.Username)
		if err != nil {
			s.logger.Error("Failed to check username existence", "error", err)

			return err
		}

		if usernameExists {
			return httperror.NewUserAlreadyExistError()
		}

		// Hash password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			s.logger.Error("Failed to hash password", "error", err)

			return err
		}

		// Generate email verification token
		verificationToken, err := s.generateVerificationToken()
		if err != nil {
			s.logger.Error("Failed to generate verification token", "error", err)

			return err
		}

		// Create user
		user := &entity.User{
			Email:                   req.Email,
			Username:                req.Username,
			PasswordHash:            string(hashedPassword),
			FirstName:               req.FirstName,
			LastName:                req.LastName,
			Roles:                   []string{"user"},
			IsActive:                true,
			IsEmailVerified:         false,
			EmailVerificationToken:  &verificationToken,
			EmailVerificationSentAt: &time.Time{},
		}
		*user.EmailVerificationSentAt = time.Now()

		if err := userRepo.Create(ctx, user); err != nil {
			s.logger.Error("Failed to create user", "error", err)

			return err
		}

		// Generate tokens
		accessToken, err := s.generateAccessToken(user)
		if err != nil {
			s.logger.Error("Failed to generate access token", "error", err)

			return err
		}

		refreshToken, err := s.generateRefreshToken(user)
		if err != nil {
			s.logger.Error("Failed to generate refresh token", "error", err)

			return err
		}

		// Create session
		session := &entity.Session{
			UserID:       user.ID,
			RefreshToken: refreshToken,
			IPAddress:    clientIP,
			UserAgent:    userAgent,
			IsActive:     true,
			ExpiresAt:    time.Now().Add(7 * 24 * time.Hour), // 7 days
		}

		if err := sessionRepo.Create(ctx, session); err != nil {
			s.logger.Error("Failed to create session", "error", err)

			return err
		}

		// Publish email verification event
		s.logger.Info("sending email verification event")

		evt := event.NewEmailVerificationRequestedEvent(
			user.ID,
			user.Email,
			verificationToken,
		)

		if err = s.emailVerificationRequestedProducer.Send(ctx, evt); err != nil {
			s.logger.Error("failed to publish email verification event", "error", err)
		}

		res = &dto.AuthResponse{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			TokenType:    "Bearer",
			ExpiresIn:    int64(s.jwtConfig.ExpirationTime.Seconds()),
			User:         dto.MapToUserResponse(user),
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return res, nil
}

// Login authenticates a user and returns tokens.
func (s *AuthService) Login(
	ctx context.Context,
	req *dto.LoginRequest,
	clientIP, userAgent string,
) (*dto.AuthResponse, error) {
	res := new(dto.AuthResponse)

	err := s.dataStore.Atomic(ctx, func(ds repository.DataStore) error {
		userRepo := ds.UserRepository()
		sessionRepo := ds.SessionRepository()
		// Get user by email
		user, err := userRepo.GetByEmail(ctx, req.Email)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return httperror.NewInvalidCredentialError()
			}

			s.logger.Error("Failed to get user by email", "error", err)

			return err
		}

		// Check if user is active
		if !user.IsActive {
			s.logger.Error("User account is inactive", "user_id", user.ID, "email", user.Email)

			return httperror.NewInvalidCredentialError()
		}

		// Verify password
		if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
			return httperror.NewInvalidCredentialError()
		}

		// Update last login
		if err := userRepo.UpdateLastLogin(ctx, user.ID); err != nil {
			s.logger.Error("Failed to update last login", "error", err)
			// Don't fail the login for this
		}

		// Generate tokens
		accessToken, err := s.generateAccessToken(user)
		if err != nil {
			s.logger.Error("Failed to generate access token", "error", err)

			return err
		}

		refreshToken, err := s.generateRefreshToken(user)
		if err != nil {
			s.logger.Error("Failed to generate refresh token", "error", err)

			return err
		}

		// Create session
		session := &entity.Session{
			UserID:       user.ID,
			RefreshToken: refreshToken,
			IPAddress:    clientIP,
			UserAgent:    userAgent,
			IsActive:     true,
			ExpiresAt:    time.Now().Add(7 * 24 * time.Hour), // 7 days
		}

		if err := sessionRepo.Create(ctx, session); err != nil {
			s.logger.Error("Failed to create session", "error", err)

			return err
		}

		s.logger.Info("User logged in", "userID", user.ID, "email", user.Email)

		res = &dto.AuthResponse{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			TokenType:    "Bearer",
			ExpiresIn:    int64(s.jwtConfig.ExpirationTime.Seconds()),
			User:         dto.MapToUserResponse(user),
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return res, nil
}

// RefreshToken refreshes an access token.
func (s *AuthService) RefreshToken(
	ctx context.Context,
	req *dto.RefreshTokenRequest,
) (*dto.AuthResponse, error) {
	userRepo := s.dataStore.UserRepository()
	// Parse and validate refresh token
	token, err := jwt.Parse(req.RefreshToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}

		return []byte(s.jwtConfig.Secret), nil
	})
	if err != nil {
		return nil, httperror.NewInvalidRefreshTokenError()
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, httperror.NewInvalidRefreshTokenError()
	}

	// Extract user ID from claims
	userIDStr, ok := claims["user_id"].(string)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, errors.New("invalid user ID in token")
	}

	// Get user
	user, err := userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("user not found")
		}

		s.logger.Error("Failed to get user by ID", "error", err)

		return nil, errors.New("failed to get user")
	}

	// Check if user is still active
	if !user.IsActive {
		return nil, errors.New("user account is inactive")
	}

	// Generate new tokens
	accessToken, err := s.generateAccessToken(user)
	if err != nil {
		s.logger.Error("Failed to generate access token", "error", err)

		return nil, errors.New("failed to generate access token")
	}

	newRefreshToken, err := s.generateRefreshToken(user)
	if err != nil {
		s.logger.Error("Failed to generate refresh token", "error", err)

		return nil, errors.New("failed to generate refresh token")
	}

	return &dto.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(s.jwtConfig.ExpirationTime.Seconds()),
		User:         dto.MapToUserResponse(user),
	}, nil
}

// GetUser gets user profile.
func (s *AuthService) GetUser(ctx context.Context, userID uuid.UUID) (*dto.UserResponse, error) {
	userRepo := s.dataStore.UserRepository()

	user, err := userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("user not found")
		}

		s.logger.Error("Failed to get user by ID", "error", err)

		return nil, errors.New("failed to get user")
	}

	userResponse := dto.MapToUserResponse(user)

	return userResponse, nil
}

// UpdateUser updates user profile.
func (s *AuthService) UpdateUser(
	ctx context.Context,
	userID uuid.UUID,
	req *dto.UpdateUserRequest,
) (*dto.UserResponse, error) {
	res := new(dto.UserResponse)

	err := s.dataStore.Atomic(ctx, func(ds repository.DataStore) error {
		userRepo := ds.UserRepository()
		// Get existing user
		user, err := userRepo.GetByID(ctx, userID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return errors.New("user not found")
			}

			s.logger.Error("Failed to get user by ID", "error", err)

			return errors.New("failed to get user")
		}

		// Update fields if provided
		if req.FirstName != "" {
			user.FirstName = req.FirstName
		}

		if req.LastName != "" {
			user.LastName = req.LastName
		}

		if req.Username != "" && req.Username != user.Username {
			// Check if new username is available
			usernameExists, err := userRepo.UsernameExists(ctx, req.Username)
			if err != nil {
				s.logger.Error("Failed to check username existence", "error", err)

				return errors.New("failed to check username existence")
			}

			if usernameExists {
				return errors.New("username already taken")
			}

			user.Username = req.Username
		}

		// Update user
		updatedUser, err := userRepo.Update(ctx, user)
		if err != nil {
			s.logger.Error("Failed to update user", "error", err)

			return errors.New("failed to update user")
		}

		s.logger.Info("User profile updated", "user_id", userID)

		res = dto.MapToUserResponse(updatedUser)

		return nil
	})
	if err != nil {
		return nil, err
	}

	return res, nil
}

// DeleteUser deletes a user.
func (s *AuthService) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	userRepo := s.dataStore.UserRepository()
	if err := userRepo.Delete(ctx, userID); err != nil {
		s.logger.Error("Failed to delete user", "error", err)

		return errors.New("failed to delete user")
	}

	sessionRepo := s.dataStore.SessionRepository()
	if err := sessionRepo.DeactivateAllUserSessions(ctx, userID); err != nil {
		s.logger.Error("Failed to delete user sessions", "error", err)
	}

	s.logger.Info("User deleted", "user_id", userID)

	return nil
}

// ChangePassword changes user password.
func (s *AuthService) ChangePassword(
	ctx context.Context,
	userID uuid.UUID,
	req *dto.ChangePasswordRequest,
) error {
	userRepo := s.dataStore.UserRepository()
	// Get user
	user, err := userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("user not found")
		}

		s.logger.Error("Failed to get user by ID", "error", err)

		return errors.New("failed to get user")
	}

	// Verify current password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.CurrentPassword)); err != nil {
		return errors.New("current password is incorrect")
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error("Failed to hash password", "error", err)

		return errors.New("failed to hash password")
	}

	// Update password
	user.PasswordHash = string(hashedPassword)

	_, err = userRepo.Update(ctx, user)
	if err != nil {
		s.logger.Error("Failed to update user password", "error", err)

		return errors.New("failed to update password")
	}

	s.logger.Info("User password changed", "user_id", userID)

	return nil
}

// VerifyEmail verifies user email.
func (s *AuthService) VerifyEmail(ctx context.Context, req *dto.VerifyEmailRequest) error {
	userRepo := s.dataStore.UserRepository()
	// Get user by verification token
	user, err := userRepo.GetByEmailVerificationToken(ctx, req.Token)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("invalid verification token: " + req.Token)
		}

		s.logger.Error("Failed to get user by verification token", "error", err)

		return errors.New("verification failed")
	}

	// Check if email is already verified
	if user.IsEmailVerified {
		return errors.New("email already verified")
	}

	// Verify email
	if err := userRepo.VerifyEmail(ctx, user.ID); err != nil {
		s.logger.Error("Failed to verify email", "error", err)

		return errors.New("failed to verify email")
	}

	// Publish email verification requested event
	evt := event.NewUserVerifiedEvent(
		user.ID,
		user.Email,
	)

	s.logger.Info(
		"sending user verified event",
		"user_id",
		user.ID,
		"email",
		user.Email,
		"event",
		evt,
	)

	if err = s.userVerifiedProducer.Send(ctx, evt); err != nil {
		s.logger.Error("failed to publish user verified event", "error", err)
	}

	return nil
}

// ResendVerification resends email verification.
func (s *AuthService) ResendVerification(
	ctx context.Context,
	req *dto.ResendVerificationRequest,
) error {
	userRepo := s.dataStore.UserRepository()
	// Get user by email
	user, err := userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("user not found")
		}

		s.logger.Error("Failed to get user by email", "error", err)

		return errors.New("failed to get user")
	}

	// Check if email is already verified
	if user.IsEmailVerified {
		return errors.New("email already verified")
	}

	// Generate new verification token
	verificationToken, err := s.generateVerificationToken()
	if err != nil {
		s.logger.Error("Failed to generate verification token", "error", err)

		return errors.New("failed to generate verification token")
	}

	// Set new verification token
	if err := userRepo.SetEmailVerificationToken(ctx, user.ID, verificationToken); err != nil {
		s.logger.Error("Failed to set verification token", "error", err)

		return errors.New("failed to set verification token")
	}

	// Publish email verification event
	s.logger.Info("resending email verification event")

	evt := event.NewEmailVerificationRequestedEvent(
		user.ID,
		user.Email,
		verificationToken,
	)

	if err = s.emailVerificationRequestedProducer.Send(ctx, evt); err != nil {
		s.logger.Error("failed to publish email verification event", "error", err)
	}

	return nil
}

// generateAccessToken generates a JWT access token.
func (s *AuthService) generateAccessToken(user *entity.User) (string, error) {
	claims := jwt.MapClaims{
		"user_id":   user.ID.String(),
		"email":     user.Email,
		"roles":     user.Roles,
		"is_active": user.IsActive,
		"exp":       time.Now().Add(s.jwtConfig.ExpirationTime).Unix(),
		"iat":       time.Now().Unix(),
		"iss":       s.jwtConfig.Issuer,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(s.jwtConfig.Secret))
}

// generateRefreshToken generates a JWT refresh token.
func (s *AuthService) generateRefreshToken(user *entity.User) (string, error) {
	claims := jwt.MapClaims{
		"user_id": user.ID.String(),
		"exp":     time.Now().Add(s.jwtConfig.RefreshTime).Unix(),
		"iat":     time.Now().Unix(),
		"iss":     s.jwtConfig.Issuer,
		"type":    "refresh",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(s.jwtConfig.Secret))
}

// generateVerificationToken generates a random verification token.
func (s *AuthService) generateVerificationToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	return hex.EncodeToString(bytes), nil
}

// GetActiveSessions returns all active sessions for a user.
func (s *AuthService) GetActiveSessions(
	ctx context.Context,
	userID uuid.UUID,
) ([]*dto.SessionResponse, error) {
	sessionRepo := s.dataStore.SessionRepository()

	sessions, err := sessionRepo.GetActiveSessionsByUserID(ctx, userID)
	if err != nil {
		s.logger.Error("failed to get active sessions", "userID", userID, "error", err)

		return nil, err
	}

	sessionDTOs := make([]*dto.SessionResponse, len(sessions))
	for i, session := range sessions {
		sessionDTOs[i] = &dto.SessionResponse{
			ID:         session.ID,
			IPAddress:  session.IPAddress,
			UserAgent:  session.UserAgent,
			IsActive:   session.IsActive,
			CreatedAt:  session.CreatedAt,
			ExpiresAt:  session.ExpiresAt,
			LastUsedAt: session.LastUsedAt,
		}
	}

	return sessionDTOs, nil
}

// Logout deactivates a specific session.
func (s *AuthService) Logout(ctx context.Context, req *dto.LogoutRequest) error {
	sessionRepo := s.dataStore.SessionRepository()

	if req.RefreshToken == "" {
		return errors.New("refresh token is required")
	}

	session, err := sessionRepo.GetByRefreshToken(ctx, req.RefreshToken)
	if err != nil {
		s.logger.Error("failed to find session by refresh token", "error", err)

		return errors.New("invalid session")
	}

	if session == nil {
		return errors.New("session not found")
	}

	err = sessionRepo.DeactivateSession(ctx, session.ID)
	if err != nil {
		s.logger.Error("failed to deactivate session", "sessionID", session.ID, "error", err)

		return err
	}

	s.logger.Info("user logged out", "sessionID", session.ID, "userID", session.UserID)

	return nil
}

// LogoutAllSessions deactivates all sessions for a user.
func (s *AuthService) LogoutAllSessions(ctx context.Context, userID uuid.UUID) error {
	sessionRepo := s.dataStore.SessionRepository()

	err := sessionRepo.DeactivateAllUserSessions(ctx, userID)
	if err != nil {
		s.logger.Error("failed to logout all sessions", "userID", userID, "error", err)

		return err
	}

	s.logger.Info("logged out all sessions for user", "userID", userID)

	return nil
}
