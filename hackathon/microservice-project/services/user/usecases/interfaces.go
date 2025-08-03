package usecases

import (
	"context"
	"time"

	"github.com/google/uuid"
	userDomain "github.com/elotusteam/microservice-project/services/user/domain"
)

// UserService defines the interface for user management operations
type UserService interface {
	CreateUser(ctx context.Context, req *CreateUserRequest) (*CreateUserResponse, error)
	GetUser(ctx context.Context, userID uuid.UUID) (*userDomain.User, error)
	GetUserByEmail(ctx context.Context, email string) (*userDomain.User, error)
	GetUserByUsername(ctx context.Context, username string) (*userDomain.User, error)
	UpdateUser(ctx context.Context, userID uuid.UUID, req *UpdateUserRequest) (*userDomain.User, error)
	DeleteUser(ctx context.Context, userID uuid.UUID) error
	ListUsers(ctx context.Context, req *ListUsersRequest) (*ListUsersResponse, error)
	SearchUsers(ctx context.Context, req *SearchUsersRequest) (*SearchUsersResponse, error)
	ChangePassword(ctx context.Context, userID uuid.UUID, req *ChangePasswordRequest) error
	UpdateUserStatus(ctx context.Context, userID uuid.UUID, status userDomain.UserStatus) error
	VerifyEmail(ctx context.Context, userID uuid.UUID) error
	GetUserStats(ctx context.Context, userID uuid.UUID) (*UserStats, error)
}

// UserProfileService defines the interface for user profile operations
type UserProfileService interface {
	CreateProfile(ctx context.Context, userID uuid.UUID, req *CreateProfileRequest) (*userDomain.UserProfile, error)
	GetProfile(ctx context.Context, userID uuid.UUID) (*userDomain.UserProfile, error)
	UpdateProfile(ctx context.Context, userID uuid.UUID, req *UpdateProfileRequest) (*userDomain.UserProfile, error)
	DeleteProfile(ctx context.Context, userID uuid.UUID) error
	UpdatePreferences(ctx context.Context, userID uuid.UUID, preferences userDomain.UserPreferences) error
	GetPreferences(ctx context.Context, userID uuid.UUID) (*userDomain.UserPreferences, error)
	UploadAvatar(ctx context.Context, userID uuid.UUID, avatarData []byte, contentType string) (string, error)
}

// UserGroupService defines the interface for user group operations
type UserGroupService interface {
	CreateGroup(ctx context.Context, req *CreateGroupRequest) (*CreateGroupResponse, error)
	GetGroup(ctx context.Context, groupID uuid.UUID) (*userDomain.UserGroup, error)
	UpdateGroup(ctx context.Context, groupID uuid.UUID, req *UpdateGroupRequest) (*userDomain.UserGroup, error)
	DeleteGroup(ctx context.Context, groupID uuid.UUID, userID uuid.UUID) error
	ListGroups(ctx context.Context, req *ListGroupsRequest) (*ListGroupsResponse, error)
	SearchGroups(ctx context.Context, req *SearchGroupsRequest) (*SearchGroupsResponse, error)
	JoinGroup(ctx context.Context, req *JoinGroupRequest) (*JoinGroupResponse, error)
	LeaveGroup(ctx context.Context, groupID uuid.UUID, userID uuid.UUID) error
	GetGroupMembers(ctx context.Context, groupID uuid.UUID, req *GetGroupMembersRequest) (*GetGroupMembersResponse, error)
	UpdateMemberRole(ctx context.Context, groupID uuid.UUID, userID uuid.UUID, role userDomain.GroupMemberRole, requesterID uuid.UUID) error
	RemoveMember(ctx context.Context, groupID uuid.UUID, userID uuid.UUID, requesterID uuid.UUID) error
	GetUserGroups(ctx context.Context, userID uuid.UUID) ([]*userDomain.UserGroup, error)
}

// UserConnectionService defines the interface for user connection operations
type UserConnectionService interface {
	SendFriendRequest(ctx context.Context, req *SendFriendRequestRequest) (*SendFriendRequestResponse, error)
	AcceptFriendRequest(ctx context.Context, connectionID uuid.UUID, userID uuid.UUID) error
	RejectFriendRequest(ctx context.Context, connectionID uuid.UUID, userID uuid.UUID) error
	RemoveFriend(ctx context.Context, friendID uuid.UUID, userID uuid.UUID) error
	FollowUser(ctx context.Context, req *FollowUserRequest) (*FollowUserResponse, error)
	UnfollowUser(ctx context.Context, targetUserID uuid.UUID, userID uuid.UUID) error
	BlockUser(ctx context.Context, req *BlockUserRequest) (*BlockUserResponse, error)
	UnblockUser(ctx context.Context, targetUserID uuid.UUID, userID uuid.UUID) error
	GetFriends(ctx context.Context, userID uuid.UUID, req *GetConnectionsRequest) (*GetConnectionsResponse, error)
	GetFollowers(ctx context.Context, userID uuid.UUID, req *GetConnectionsRequest) (*GetConnectionsResponse, error)
	GetFollowing(ctx context.Context, userID uuid.UUID, req *GetConnectionsRequest) (*GetConnectionsResponse, error)
	GetPendingRequests(ctx context.Context, userID uuid.UUID, connectionType userDomain.ConnectionType) ([]*userDomain.UserConnection, error)
	GetConnectionStats(ctx context.Context, userID uuid.UUID) (*ConnectionStats, error)
}

// UserSessionService defines the interface for user session operations
type UserSessionService interface {
	CreateSession(ctx context.Context, req *CreateSessionRequest) (*CreateSessionResponse, error)
	GetSession(ctx context.Context, sessionID uuid.UUID) (*userDomain.UserSession, error)
	GetSessionByToken(ctx context.Context, token string) (*userDomain.UserSession, error)
	UpdateSession(ctx context.Context, sessionID uuid.UUID) error
	DeleteSession(ctx context.Context, sessionID uuid.UUID) error
	GetUserSessions(ctx context.Context, userID uuid.UUID) ([]*userDomain.UserSession, error)
	DeleteAllUserSessions(ctx context.Context, userID uuid.UUID) error
	CleanupExpiredSessions(ctx context.Context) error
}

// PasswordService defines the interface for password operations
type PasswordService interface {
	HashPassword(password string) (string, error)
	VerifyPassword(hashedPassword, password string) error
	ValidatePasswordStrength(password string) []string
	GenerateResetToken(userID uuid.UUID) (string, error)
	ValidateResetToken(token string) (uuid.UUID, error)
	InvalidateResetToken(token string) error
}

// NotificationService defines the interface for sending notifications
type NotificationService interface {
	SendWelcomeEmail(ctx context.Context, userID uuid.UUID, email string) error
	SendEmailVerification(ctx context.Context, userID uuid.UUID, email string, token string) error
	SendPasswordResetEmail(ctx context.Context, userID uuid.UUID, email string, token string) error
	SendFriendRequestNotification(ctx context.Context, requesterID, addresseeID uuid.UUID) error
	SendGroupInvitationNotification(ctx context.Context, groupID, inviterID, inviteeID uuid.UUID) error
}

// ActivityService defines the interface for logging activities
type ActivityService interface {
	LogActivity(ctx context.Context, userID uuid.UUID, action, resourceType string, resourceID *uuid.UUID, details map[string]interface{}, ipAddress, userAgent string) error
}

// Request/Response DTOs
type CreateUserRequest struct {
	Username  string `json:"username" validate:"required,min=3,max=50"`
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,min=8"`
	FirstName string `json:"first_name" validate:"required,min=1,max=50"`
	LastName  string `json:"last_name" validate:"required,min=1,max=50"`
	Role      userDomain.UserRole `json:"role,omitempty"`
}

type CreateUserResponse struct {
	User         *userDomain.User `json:"user"`
	AccessToken  string           `json:"access_token,omitempty"`
	RefreshToken string           `json:"refresh_token,omitempty"`
}

type UpdateUserRequest struct {
	Username    *string `json:"username,omitempty"`
	Email       *string `json:"email,omitempty"`
	FirstName   *string `json:"first_name,omitempty"`
	LastName    *string `json:"last_name,omitempty"`
	DisplayName *string `json:"display_name,omitempty"`
	Bio         *string `json:"bio,omitempty"`
	Timezone    *string `json:"timezone,omitempty"`
	Language    *string `json:"language,omitempty"`
}

type ListUsersRequest struct {
	Limit    int                    `json:"limit"`
	Offset   int                    `json:"offset"`
	SortBy   string                 `json:"sort_by"`
	SortDesc bool                   `json:"sort_desc"`
	Status   *userDomain.UserStatus `json:"status,omitempty"`
	Role     *userDomain.UserRole   `json:"role,omitempty"`
}

type ListUsersResponse struct {
	Users   []*userDomain.User `json:"users"`
	Total   int64              `json:"total"`
	Limit   int                `json:"limit"`
	Offset  int                `json:"offset"`
	HasMore bool               `json:"has_more"`
}

type SearchUsersRequest struct {
	Query    string `json:"query"`
	Limit    int    `json:"limit"`
	Offset   int    `json:"offset"`
	SortBy   string `json:"sort_by"`
	SortDesc bool   `json:"sort_desc"`
}

type SearchUsersResponse struct {
	Users   []*userDomain.User `json:"users"`
	Total   int64              `json:"total"`
	Limit   int                `json:"limit"`
	Offset  int                `json:"offset"`
	HasMore bool               `json:"has_more"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,min=8"`
}

type UserStats struct {
	TotalFiles      int64 `json:"total_files"`
	TotalStorage    int64 `json:"total_storage"`
	FriendsCount    int64 `json:"friends_count"`
	FollowersCount  int64 `json:"followers_count"`
	FollowingCount  int64 `json:"following_count"`
	GroupsCount     int64 `json:"groups_count"`
	ActiveSessions  int64 `json:"active_sessions"`
}

type CreateProfileRequest struct {
	PhoneNumber string                           `json:"phone_number,omitempty"`
	Address     string                           `json:"address,omitempty"`
	City        string                           `json:"city,omitempty"`
	Country     string                           `json:"country,omitempty"`
	PostalCode  string                           `json:"postal_code,omitempty"`
	DateOfBirth *time.Time                       `json:"date_of_birth,omitempty"`
	Gender      string                           `json:"gender,omitempty"`
	Website     string                           `json:"website,omitempty"`
	SocialLinks map[string]string                `json:"social_links,omitempty"`
	Preferences userDomain.UserPreferences       `json:"preferences,omitempty"`
}

type UpdateProfileRequest struct {
	PhoneNumber *string                          `json:"phone_number,omitempty"`
	Address     *string                          `json:"address,omitempty"`
	City        *string                          `json:"city,omitempty"`
	Country     *string                          `json:"country,omitempty"`
	PostalCode  *string                          `json:"postal_code,omitempty"`
	DateOfBirth *time.Time                       `json:"date_of_birth,omitempty"`
	Gender      *string                          `json:"gender,omitempty"`
	Website     *string                          `json:"website,omitempty"`
	SocialLinks map[string]string                `json:"social_links,omitempty"`
}

type CreateGroupRequest struct {
	Name        string                  `json:"name" validate:"required,min=3,max=100"`
	Description string                  `json:"description,omitempty"`
	Type        userDomain.GroupType    `json:"type" validate:"required"`
	OwnerID     uuid.UUID               `json:"owner_id" validate:"required"`
	IsPublic    bool                    `json:"is_public"`
}

type CreateGroupResponse struct {
	Group      *userDomain.UserGroup           `json:"group"`
	Membership *userDomain.UserGroupMembership `json:"membership"`
}

type UpdateGroupRequest struct {
	Name        *string               `json:"name,omitempty"`
	Description *string               `json:"description,omitempty"`
	Type        *userDomain.GroupType `json:"type,omitempty"`
	IsPublic    *bool                 `json:"is_public,omitempty"`
}

type ListGroupsRequest struct {
	Limit    int                   `json:"limit"`
	Offset   int                   `json:"offset"`
	SortBy   string                `json:"sort_by"`
	SortDesc bool                  `json:"sort_desc"`
	Type     *userDomain.GroupType `json:"type,omitempty"`
	OwnerID  *uuid.UUID            `json:"owner_id,omitempty"`
	PublicOnly bool                `json:"public_only"`
}

type ListGroupsResponse struct {
	Groups  []*userDomain.UserGroup `json:"groups"`
	Total   int64                   `json:"total"`
	Limit   int                     `json:"limit"`
	Offset  int                     `json:"offset"`
	HasMore bool                    `json:"has_more"`
}

type SearchGroupsRequest struct {
	Query    string `json:"query"`
	Limit    int    `json:"limit"`
	Offset   int    `json:"offset"`
	SortBy   string `json:"sort_by"`
	SortDesc bool   `json:"sort_desc"`
}

type SearchGroupsResponse struct {
	Groups  []*userDomain.UserGroup `json:"groups"`
	Total   int64                   `json:"total"`
	Limit   int                     `json:"limit"`
	Offset  int                     `json:"offset"`
	HasMore bool                    `json:"has_more"`
}

type JoinGroupRequest struct {
	GroupID uuid.UUID `json:"group_id" validate:"required"`
	UserID  uuid.UUID `json:"user_id" validate:"required"`
	Message string    `json:"message,omitempty"`
}

type JoinGroupResponse struct {
	Membership *userDomain.UserGroupMembership `json:"membership"`
	RequiresApproval bool                      `json:"requires_approval"`
}

type GetGroupMembersRequest struct {
	Limit  int                              `json:"limit"`
	Offset int                              `json:"offset"`
	Role   *userDomain.GroupMemberRole      `json:"role,omitempty"`
	Status *userDomain.MembershipStatus     `json:"status,omitempty"`
}

type GetGroupMembersResponse struct {
	Members []*userDomain.UserGroupMembership `json:"members"`
	Total   int64                             `json:"total"`
	Limit   int                               `json:"limit"`
	Offset  int                               `json:"offset"`
	HasMore bool                              `json:"has_more"`
}

type SendFriendRequestRequest struct {
	RequesterID uuid.UUID `json:"requester_id" validate:"required"`
	AddresseeID uuid.UUID `json:"addressee_id" validate:"required"`
	Message     string    `json:"message,omitempty"`
}

type SendFriendRequestResponse struct {
	Connection *userDomain.UserConnection `json:"connection"`
}

type FollowUserRequest struct {
	FollowerID uuid.UUID `json:"follower_id" validate:"required"`
	FolloweeID uuid.UUID `json:"followee_id" validate:"required"`
}

type FollowUserResponse struct {
	Connection *userDomain.UserConnection `json:"connection"`
}

type BlockUserRequest struct {
	BlockerID uuid.UUID `json:"blocker_id" validate:"required"`
	BlockedID uuid.UUID `json:"blocked_id" validate:"required"`
	Reason    string    `json:"reason,omitempty"`
}

type BlockUserResponse struct {
	Connection *userDomain.UserConnection `json:"connection"`
}

type GetConnectionsRequest struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

type GetConnectionsResponse struct {
	Connections []*userDomain.UserConnection `json:"connections"`
	Total       int64                        `json:"total"`
	Limit       int                          `json:"limit"`
	Offset      int                          `json:"offset"`
	HasMore     bool                         `json:"has_more"`
}

type ConnectionStats struct {
	FriendsCount   int64 `json:"friends_count"`
	FollowersCount int64 `json:"followers_count"`
	FollowingCount int64 `json:"following_count"`
	PendingRequests int64 `json:"pending_requests"`
}

type CreateSessionRequest struct {
	UserID    uuid.UUID `json:"user_id" validate:"required"`
	Token     string    `json:"token" validate:"required"`
	IPAddress string    `json:"ip_address"`
	UserAgent string    `json:"user_agent"`
	Location  string    `json:"location,omitempty"`
	Device    string    `json:"device,omitempty"`
	ExpiresAt time.Time `json:"expires_at" validate:"required"`
}

type CreateSessionResponse struct {
	Session *userDomain.UserSession `json:"session"`
}