// Package service provides business logic for the auth service.
package service

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/kafka"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/utils/encryptutils"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/utils/jwtutils"

	"github.com/raphaeldiscky/go-micro-commerce/auth-service/internal/dto"
	"github.com/raphaeldiscky/go-micro-commerce/auth-service/internal/entity"
	"github.com/raphaeldiscky/go-micro-commerce/auth-service/internal/httperror"
	"github.com/raphaeldiscky/go-micro-commerce/auth-service/internal/mapper"
	"github.com/raphaeldiscky/go-micro-commerce/auth-service/internal/mq"
	"github.com/raphaeldiscky/go-micro-commerce/auth-service/internal/repository"
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
	jwtUtils                           jwtutils.JWTInterface
	hasher                             encryptutils.HasherInterface
	logger                             logger.Logger
	emailVerificationRequestedProducer kafka.ProducerInterface
	userVerifiedProducer               kafka.ProducerInterface
}

// NewAuthService creates a new AuthService.
func NewAuthService(
	dataStore repository.DataStore,
	jwtUtils jwtutils.JWTInterface,
	hasher encryptutils.HasherInterface,
	appLogger logger.Logger,
	emailVerificationRequestedProducer kafka.ProducerInterface,
	userVerifiedProducer kafka.ProducerInterface,
) AuthServiceInterface {
	return &AuthService{
		dataStore:                          dataStore,
		jwtUtils:                           jwtUtils,
		hasher:                             hasher,
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
			return httperror.NewInternalServerError("failed to check email")
		}

		if emailExists {
			return httperror.NewUserAlreadyExistError()
		}

		// Check if username already exists
		usernameExists, err := userRepo.UsernameExists(ctx, req.Username)
		if err != nil {
			return httperror.NewInternalServerError("failed to check username")
		}

		if usernameExists {
			return httperror.NewUserAlreadyExistError()
		}

		// Hash password
		hashedPassword, err := s.hasher.Hash(req.Password)
		if err != nil {
			return httperror.NewInvalidCredentialError()
		}

		// Generate email verification token
		verificationToken, err := s.generateVerificationToken()
		if err != nil {
			return httperror.NewInvalidCredentialError()
		}

		// Create user
		user := &entity.User{
			Email:                   req.Email,
			Username:                req.Username,
			PasswordHash:            hashedPassword,
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
			return httperror.NewInternalServerError("failed to create user")
		}

		// Generate tokens using JWT utils
		accessToken, err := s.jwtUtils.GenerateAccessToken(
			user.ID.String(),
			user.Email,
			user.Roles,
			user.IsActive,
		)
		if err != nil {
			return httperror.NewInvalidCredentialError()
		}

		refreshToken, err := s.jwtUtils.GenerateRefreshToken(user.ID.String())
		if err != nil {
			return httperror.NewInvalidRefreshTokenError()
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
			return httperror.NewInternalServerError("failed to create session")
		}

		// Publish email verification event
		s.logger.Info("sending email verification event")

		evt := mq.NewEmailVerificationRequestedEvent(
			user.ID,
			user.Email,
			verificationToken,
		)

		if err = s.emailVerificationRequestedProducer.Send(ctx, evt); err != nil {
			s.logger.Error("failed to publish email verification event", "error", err)
		}

		expTime, err := s.jwtUtils.GetExpirationTime(accessToken)
		if err != nil {
			s.logger.Error("Failed to get access token expiration time", "error", err)

			return err
		}

		res = &dto.AuthResponse{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			TokenType:    "Bearer",
			ExpiresIn:    expTime,
			User:         mapper.MapToUserResponse(user),
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
		if !s.hasher.Check(req.Password, user.PasswordHash) {
			return httperror.NewInvalidCredentialError()
		}

		// Update last login
		if err := userRepo.UpdateLastLogin(ctx, user.ID); err != nil {
			s.logger.Error("Failed to update last login", "error", err)
			// Don't fail the login for this
		}

		// Generate tokens using JWT utils
		accessToken, err := s.jwtUtils.GenerateAccessToken(
			user.ID.String(),
			user.Email,
			user.Roles,
			user.IsActive,
		)
		if err != nil {
			s.logger.Error("Failed to generate access token", "error", err)

			return err
		}

		refreshToken, err := s.jwtUtils.GenerateRefreshToken(user.ID.String())
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

		expTime, err := s.jwtUtils.GetExpirationTime(accessToken)
		if err != nil {
			s.logger.Error("Failed to get access token expiration time", "error", err)

			return err
		}

		res = &dto.AuthResponse{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			TokenType:    "Bearer",
			ExpiresIn:    expTime,
			User:         mapper.MapToUserResponse(user),
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

	// Validate refresh token using JWT utils
	claims, err := s.jwtUtils.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		return nil, httperror.NewInvalidRefreshTokenError()
	}

	// Parse user ID from claims
	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return nil, errors.New("invalid user ID in token")
	}

	// Get user
	user, err := userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, httperror.NewNotFoundError("user not found")
		}

		s.logger.Error("Failed to get user by ID", "error", err)

		return nil, httperror.NewInternalServerError("failed to get user")
	}

	// Check if user is still active
	if !user.IsActive {
		return nil, httperror.NewForbiddenError("user account is inactive")
	}

	// Generate new tokens using JWT utils
	accessToken, err := s.jwtUtils.GenerateAccessToken(
		user.ID.String(),
		user.Email,
		user.Roles,
		user.IsActive,
	)
	if err != nil {
		s.logger.Error("Failed to generate access token", "error", err)

		return nil, httperror.NewInternalServerError("failed to generate access token")
	}

	newRefreshToken, err := s.jwtUtils.GenerateRefreshToken(user.ID.String())
	if err != nil {
		s.logger.Error("Failed to generate refresh token", "error", err)

		return nil, httperror.NewInternalServerError("failed to generate refresh token")
	}

	expTime, err := s.jwtUtils.GetExpirationTime(accessToken)
	if err != nil {
		s.logger.Error("Failed to get access token expiration time", "error", err)

		return nil, httperror.NewInternalServerError("failed to get access token expiration time")
	}

	return &dto.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    expTime,
		User:         mapper.MapToUserResponse(user),
	}, nil
}

// GetUser gets user profile.
func (s *AuthService) GetUser(ctx context.Context, userID uuid.UUID) (*dto.UserResponse, error) {
	userRepo := s.dataStore.UserRepository()

	user, err := userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, httperror.NewNotFoundError("user not found")
		}

		s.logger.Error("Failed to get user by ID", "error", err)

		return nil, httperror.NewInternalServerError("failed to get user")
	}

	userResponse := mapper.MapToUserResponse(user)

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
				return httperror.NewNotFoundError("user not found")
			}

			s.logger.Error("Failed to get user by ID", "error", err)

			return httperror.NewInternalServerError("failed to get user")
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

				return httperror.NewInternalServerError("failed to check username existence")
			}

			if usernameExists {
				return httperror.NewUserAlreadyExistError()
			}

			user.Username = req.Username
		}

		// Update user
		updatedUser, err := userRepo.Update(ctx, user)
		if err != nil {
			s.logger.Error("Failed to update user", "error", err)

			return httperror.NewInternalServerError("failed to update user")
		}

		s.logger.Info("User profile updated", "user_id", userID)

		res = mapper.MapToUserResponse(updatedUser)

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

		return httperror.NewInternalServerError("failed to delete user")
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
			return httperror.NewNotFoundError("user not found")
		}

		s.logger.Error("Failed to get user by ID", "error", err)

		return httperror.NewInternalServerError("failed to get user")
	}

	// Verify current password
	if !s.hasher.Check(req.CurrentPassword, user.PasswordHash) {
		return httperror.NewInvalidCredentialError()
	}

	// Hash new password
	hashedPassword, err := s.hasher.Hash(req.NewPassword)
	if err != nil {
		s.logger.Error("Failed to hash password", "error", err)

		return httperror.NewInternalServerError("failed to hash password")
	}

	// Update password
	user.PasswordHash = hashedPassword

	_, err = userRepo.Update(ctx, user)
	if err != nil {
		s.logger.Error("Failed to update user password", "error", err)

		return httperror.NewInternalServerError("failed to update password")
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
			return httperror.NewInvalidCredentialError()
		}

		s.logger.Error("Failed to get user by verification token", "error", err)

		return httperror.NewInternalServerError("verification failed")
	}

	// Check if email is already verified
	if user.IsEmailVerified {
		return httperror.NewBadRequestError("email already verified")
	}

	// Verify email
	if err := userRepo.VerifyEmail(ctx, user.ID); err != nil {
		s.logger.Error("Failed to verify email", "error", err)

		return httperror.NewInternalServerError("failed to verify email")
	}

	// Publish email verification requested event
	evt := mq.NewUserVerifiedEvent(user.ID, user.Email)

	s.logger.Info(
		"sending user verified event",
		"user_id", user.ID,
		"email", user.Email,
		"event", evt,
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
			return httperror.NewNotFoundError("user not found")
		}

		s.logger.Error("Failed to get user by email", "error", err)

		return httperror.NewInternalServerError("failed to get user")
	}

	// Check if email is already verified
	if user.IsEmailVerified {
		return httperror.NewBadRequestError("email already verified")
	}

	// Generate new verification token
	verificationToken, err := s.generateVerificationToken()
	if err != nil {
		s.logger.Error("Failed to generate verification token", "error", err)

		return httperror.NewInternalServerError("failed to generate verification token")
	}

	// Set new verification token
	if err := userRepo.SetEmailVerificationToken(ctx, user.ID, verificationToken); err != nil {
		s.logger.Error("Failed to set verification token", "error", err)

		return httperror.NewInternalServerError("failed to set verification token")
	}

	// Publish email verification event
	s.logger.Info("resending email verification event")

	evt := mq.NewEmailVerificationRequestedEvent(
		user.ID,
		user.Email,
		verificationToken,
	)

	if err = s.emailVerificationRequestedProducer.Send(ctx, evt); err != nil {
		s.logger.Error("failed to publish email verification event", "error", err)
	}

	return nil
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
		return httperror.NewInvalidCredentialError()
	}

	session, err := sessionRepo.GetByRefreshToken(ctx, req.RefreshToken)
	if err != nil {
		s.logger.Error("failed to find session by refresh token", "error", err)

		return httperror.NewInvalidCredentialError()
	}

	if session == nil {
		return httperror.NewInvalidCredentialError()
	}

	err = sessionRepo.DeactivateSession(ctx, session.ID)
	if err != nil {
		s.logger.Error("failed to deactivate session", "sessionID", session.ID, "error", err)

		return httperror.NewInternalServerError("failed to deactivate session")
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

		return httperror.NewInternalServerError("failed to logout all sessions")
	}

	s.logger.Info("logged out all sessions for user", "userID", userID)

	return nil
}
