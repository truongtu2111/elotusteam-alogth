package usecases

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	userDomain "github.com/elotusteam/microservice-project/services/user/domain"
	"github.com/elotusteam/microservice-project/shared/config"
)

type userService struct {
	repos            userDomain.RepositoryManager
	passwordService  PasswordService
	notificationSvc  NotificationService
	activitySvc      ActivityService
	config          *config.Config
}

// NewUserService creates a new user service instance
func NewUserService(
	repos userDomain.RepositoryManager,
	passwordService PasswordService,
	notificationSvc NotificationService,
	activitySvc ActivityService,
	config *config.Config,
) UserService {
	return &userService{
		repos:           repos,
		passwordService: passwordService,
		notificationSvc: notificationSvc,
		activitySvc:     activitySvc,
		config:         config,
	}
}

func (s *userService) CreateUser(ctx context.Context, req *CreateUserRequest) (*CreateUserResponse, error) {
	// Validate request
	if err := s.validateCreateUserRequest(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Check if user already exists
	existingUser, _ := s.repos.User().GetByEmail(ctx, req.Email)
	if existingUser != nil {
		return nil, fmt.Errorf("user with email %s already exists", req.Email)
	}

	existingUser, _ = s.repos.User().GetByUsername(ctx, req.Username)
	if existingUser != nil {
		return nil, fmt.Errorf("user with username %s already exists", req.Username)
	}

	// Hash password
	hashedPassword, err := s.passwordService.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user entity
	user := &userDomain.User{
		ID:           uuid.New(),
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: hashedPassword,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		DisplayName:  fmt.Sprintf("%s %s", req.FirstName, req.LastName),
		Role:         req.Role,
		Status:       userDomain.UserStatusPending,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Set default role if not specified
	if user.Role == "" {
		user.Role = userDomain.UserRoleUser
	}

	// Create user in repository
	err = s.repos.User().Create(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Log activity
	if s.activitySvc != nil {
		_ = s.activitySvc.LogActivity(ctx, user.ID, "user_created", "user", &user.ID, map[string]interface{}{
			"username": user.Username,
			"email":    user.Email,
			"role":     user.Role,
		}, "", "")
	}

	// Send welcome email
	if s.notificationSvc != nil {
		go func() {
			_ = s.notificationSvc.SendWelcomeEmail(context.Background(), user.ID, user.Email)
		}()
	}

	return &CreateUserResponse{
		User: user,
	}, nil
}

func (s *userService) GetUser(ctx context.Context, userID uuid.UUID) (*userDomain.User, error) {
	user, err := s.repos.User().GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return user, nil
}

func (s *userService) GetUserByEmail(ctx context.Context, email string) (*userDomain.User, error) {
	user, err := s.repos.User().GetByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}
	return user, nil
}

func (s *userService) GetUserByUsername(ctx context.Context, username string) (*userDomain.User, error) {
	user, err := s.repos.User().GetByUsername(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}
	return user, nil
}

func (s *userService) UpdateUser(ctx context.Context, userID uuid.UUID, req *UpdateUserRequest) (*userDomain.User, error) {
	// Get existing user
	user, err := s.repos.User().GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Update fields if provided
	if req.Username != nil {
		// Check if username is already taken
		existingUser, _ := s.repos.User().GetByUsername(ctx, *req.Username)
		if existingUser != nil && existingUser.ID != userID {
			return nil, fmt.Errorf("username %s is already taken", *req.Username)
		}
		user.Username = *req.Username
	}

	if req.Email != nil {
		// Check if email is already taken
		existingUser, _ := s.repos.User().GetByEmail(ctx, *req.Email)
		if existingUser != nil && existingUser.ID != userID {
			return nil, fmt.Errorf("email %s is already taken", *req.Email)
		}
		user.Email = *req.Email
		user.EmailVerified = false // Reset email verification
	}

	if req.FirstName != nil {
		user.FirstName = *req.FirstName
	}

	if req.LastName != nil {
		user.LastName = *req.LastName
	}

	if req.DisplayName != nil {
		user.DisplayName = *req.DisplayName
	} else if req.FirstName != nil || req.LastName != nil {
		// Update display name if first or last name changed
		user.DisplayName = fmt.Sprintf("%s %s", user.FirstName, user.LastName)
	}

	if req.Bio != nil {
		user.Bio = *req.Bio
	}

	if req.Timezone != nil {
		user.Timezone = *req.Timezone
	}

	if req.Language != nil {
		user.Language = *req.Language
	}

	user.UpdatedAt = time.Now()

	// Update user in repository
	err = s.repos.User().Update(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	// Log activity
	if s.activitySvc != nil {
		_ = s.activitySvc.LogActivity(ctx, userID, "user_updated", "user", &userID, map[string]interface{}{
			"updated_fields": s.getUpdatedFields(req),
		}, "", "")
	}

	return user, nil
}

func (s *userService) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	// Get user first to ensure it exists
	user, err := s.repos.User().GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	// Soft delete by updating status and setting deleted_at
	user.Status = userDomain.UserStatusInactive
	now := time.Now()
	user.DeletedAt = &now
	user.UpdatedAt = now

	err = s.repos.User().Update(ctx, user)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	// Log activity
	if s.activitySvc != nil {
		_ = s.activitySvc.LogActivity(ctx, userID, "user_deleted", "user", &userID, map[string]interface{}{
			"username": user.Username,
			"email":    user.Email,
		}, "", "")
	}

	return nil
}

func (s *userService) ListUsers(ctx context.Context, req *ListUsersRequest) (*ListUsersResponse, error) {
	// Set default values
	if req.Limit <= 0 {
		req.Limit = 20
	}
	if req.Limit > 100 {
		req.Limit = 100
	}
	if req.SortBy == "" {
		req.SortBy = "created_at"
	}

	// Get users from repository
	users, err := s.repos.User().List(ctx, req.Limit, req.Offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	// Get total count
	total, err := s.repos.User().Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count users: %w", err)
	}

	return &ListUsersResponse{
		Users:   users,
		Total:   total,
		Limit:   req.Limit,
		Offset:  req.Offset,
		HasMore: int64(req.Offset+req.Limit) < total,
	}, nil
}

func (s *userService) SearchUsers(ctx context.Context, req *SearchUsersRequest) (*SearchUsersResponse, error) {
	// Set default values
	if req.Limit <= 0 {
		req.Limit = 20
	}
	if req.Limit > 100 {
		req.Limit = 100
	}
	if req.SortBy == "" {
		req.SortBy = "created_at"
	}

	// Search users in repository
	users, err := s.repos.User().Search(ctx, req.Query, req.Limit, req.Offset)
	if err != nil {
		return nil, fmt.Errorf("failed to search users: %w", err)
	}

	// For search, we'll use the length of results as total for simplicity
	total := int64(len(users))

	return &SearchUsersResponse{
		Users:   users,
		Total:   total,
		Limit:   req.Limit,
		Offset:  req.Offset,
		HasMore: int64(req.Offset+req.Limit) < total,
	}, nil
}

func (s *userService) ChangePassword(ctx context.Context, userID uuid.UUID, req *ChangePasswordRequest) error {
	// Get user
	user, err := s.repos.User().GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	// Verify current password
	if err := s.passwordService.VerifyPassword(user.PasswordHash, req.CurrentPassword); err != nil {
		return fmt.Errorf("current password is incorrect")
	}

	// Validate new password strength
	if violations := s.passwordService.ValidatePasswordStrength(req.NewPassword); len(violations) > 0 {
		return fmt.Errorf("password validation failed: %s", strings.Join(violations, ", "))
	}

	// Hash new password
	newPasswordHash, err := s.passwordService.HashPassword(req.NewPassword)
	if err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}

	// Update password
	user.PasswordHash = newPasswordHash
	user.UpdatedAt = time.Now()

	err = s.repos.User().Update(ctx, user)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	// Log activity
	if s.activitySvc != nil {
		_ = s.activitySvc.LogActivity(ctx, userID, "password_changed", "user", &userID, map[string]interface{}{}, "", "")
	}

	return nil
}

func (s *userService) UpdateUserStatus(ctx context.Context, userID uuid.UUID, status userDomain.UserStatus) error {
	// Get user
	user, err := s.repos.User().GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	// Update status
	oldStatus := user.Status
	user.Status = status
	user.UpdatedAt = time.Now()

	err = s.repos.User().Update(ctx, user)
	if err != nil {
		return fmt.Errorf("failed to update user status: %w", err)
	}

	// Log activity
	if s.activitySvc != nil {
		_ = s.activitySvc.LogActivity(ctx, userID, "status_updated", "user", &userID, map[string]interface{}{
			"old_status": oldStatus,
			"new_status": status,
		}, "", "")
	}

	return nil
}

func (s *userService) VerifyEmail(ctx context.Context, userID uuid.UUID) error {
	// Get user
	user, err := s.repos.User().GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	// Update email verification status
	user.EmailVerified = true
	user.EmailVerifiedAt = &time.Time{}
	*user.EmailVerifiedAt = time.Now()
	user.UpdatedAt = time.Now()

	// Update status to active if pending
	if user.Status == userDomain.UserStatusPending {
		user.Status = userDomain.UserStatusActive
	}

	err = s.repos.User().Update(ctx, user)
	if err != nil {
		return fmt.Errorf("failed to verify email: %w", err)
	}

	// Log activity
	if s.activitySvc != nil {
		_ = s.activitySvc.LogActivity(ctx, userID, "email_verified", "user", &userID, map[string]interface{}{}, "", "")
	}

	return nil
}

func (s *userService) GetUserStats(ctx context.Context, userID uuid.UUID) (*UserStats, error) {
	// Get user to ensure it exists
	_, err := s.repos.User().GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Get stats from various repositories
	stats := &UserStats{}

	// Get connection stats
	connectionRepo := s.repos.UserConnection()
	if connectionRepo != nil {
		friends, _ := connectionRepo.CountFriends(ctx, userID)
		followers, _ := connectionRepo.CountFollowers(ctx, userID)
		following, _ := connectionRepo.CountFollowing(ctx, userID)
		stats.FriendsCount = friends
		stats.FollowersCount = followers
		stats.FollowingCount = following
	}

	// Get group count
	membershipRepo := s.repos.UserGroupMembership()
	if membershipRepo != nil {
		groupCount, _ := membershipRepo.CountByUserID(ctx, userID)
		stats.GroupsCount = groupCount
	}

	// Get active sessions count
	sessionRepo := s.repos.UserSession()
	if sessionRepo != nil {
		activeSessions, _ := sessionRepo.GetActiveSessions(ctx, userID)
		stats.ActiveSessions = int64(len(activeSessions))
	}

	return stats, nil
}

// Helper functions

func (s *userService) validateCreateUserRequest(req *CreateUserRequest) error {
	if req.Username == "" {
		return fmt.Errorf("username is required")
	}
	if len(req.Username) < 3 || len(req.Username) > 50 {
		return fmt.Errorf("username must be between 3 and 50 characters")
	}
	if req.Email == "" {
		return fmt.Errorf("email is required")
	}
	if req.Password == "" {
		return fmt.Errorf("password is required")
	}
	if len(req.Password) < 8 {
		return fmt.Errorf("password must be at least 8 characters")
	}
	if req.FirstName == "" {
		return fmt.Errorf("first name is required")
	}
	if req.LastName == "" {
		return fmt.Errorf("last name is required")
	}

	// Validate password strength
	if violations := s.passwordService.ValidatePasswordStrength(req.Password); len(violations) > 0 {
		return fmt.Errorf("password validation failed: %s", strings.Join(violations, ", "))
	}

	return nil
}

func (s *userService) getUpdatedFields(req *UpdateUserRequest) []string {
	var fields []string
	if req.Username != nil {
		fields = append(fields, "username")
	}
	if req.Email != nil {
		fields = append(fields, "email")
	}
	if req.FirstName != nil {
		fields = append(fields, "first_name")
	}
	if req.LastName != nil {
		fields = append(fields, "last_name")
	}
	if req.DisplayName != nil {
		fields = append(fields, "display_name")
	}
	if req.Bio != nil {
		fields = append(fields, "bio")
	}
	if req.Timezone != nil {
		fields = append(fields, "timezone")
	}
	if req.Language != nil {
		fields = append(fields, "language")
	}
	return fields
}