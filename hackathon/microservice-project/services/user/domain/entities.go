package domain

import (
	"time"
	"github.com/google/uuid"
)

// User represents a user entity in the system
type User struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Username    string    `json:"username" db:"username"`
	Email       string    `json:"email" db:"email"`
	PasswordHash string   `json:"-" db:"password_hash"`
	FirstName   string    `json:"first_name" db:"first_name"`
	LastName    string    `json:"last_name" db:"last_name"`
	DisplayName string    `json:"display_name" db:"display_name"`
	Avatar      string    `json:"avatar" db:"avatar"`
	Bio         string    `json:"bio" db:"bio"`
	Timezone    string    `json:"timezone" db:"timezone"`
	Language    string    `json:"language" db:"language"`
	Status      UserStatus `json:"status" db:"status"`
	Role        UserRole  `json:"role" db:"role"`
	EmailVerified bool    `json:"email_verified" db:"email_verified"`
	EmailVerifiedAt *time.Time `json:"email_verified_at,omitempty" db:"email_verified_at"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty" db:"last_login_at"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// UserStatus represents the status of a user
type UserStatus string

const (
	UserStatusActive    UserStatus = "active"
	UserStatusInactive  UserStatus = "inactive"
	UserStatusSuspended UserStatus = "suspended"
	UserStatusBanned    UserStatus = "banned"
	UserStatusPending   UserStatus = "pending"
)

// UserRole represents the role of a user
type UserRole string

const (
	UserRoleUser      UserRole = "user"
	UserRoleModerator UserRole = "moderator"
	UserRoleAdmin     UserRole = "admin"
	UserRoleSuperAdmin UserRole = "super_admin"
)

// UserProfile represents extended user profile information
type UserProfile struct {
	UserID      uuid.UUID `json:"user_id" db:"user_id"`
	PhoneNumber string    `json:"phone_number" db:"phone_number"`
	Address     string    `json:"address" db:"address"`
	City        string    `json:"city" db:"city"`
	Country     string    `json:"country" db:"country"`
	PostalCode  string    `json:"postal_code" db:"postal_code"`
	DateOfBirth *time.Time `json:"date_of_birth,omitempty" db:"date_of_birth"`
	Gender      string    `json:"gender" db:"gender"`
	Website     string    `json:"website" db:"website"`
	SocialLinks map[string]string `json:"social_links" db:"social_links"`
	Preferences UserPreferences   `json:"preferences" db:"preferences"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// UserPreferences represents user preferences
type UserPreferences struct {
	Notifications NotificationPreferences `json:"notifications"`
	Privacy       PrivacyPreferences      `json:"privacy"`
	Theme         string                  `json:"theme"`
	Language      string                  `json:"language"`
	Timezone      string                  `json:"timezone"`
}

// NotificationPreferences represents notification preferences
type NotificationPreferences struct {
	Email    bool `json:"email"`
	SMS      bool `json:"sms"`
	Push     bool `json:"push"`
	InApp    bool `json:"in_app"`
	Marketing bool `json:"marketing"`
}

// PrivacyPreferences represents privacy preferences
type PrivacyPreferences struct {
	ProfileVisibility string `json:"profile_visibility"` // public, friends, private
	ShowEmail         bool   `json:"show_email"`
	ShowPhone         bool   `json:"show_phone"`
	AllowMessages     bool   `json:"allow_messages"`
	AllowFriendRequests bool `json:"allow_friend_requests"`
}

// UserGroup represents a user group
type UserGroup struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	Type        GroupType `json:"type" db:"type"`
	OwnerID     uuid.UUID `json:"owner_id" db:"owner_id"`
	IsPublic    bool      `json:"is_public" db:"is_public"`
	MemberCount int       `json:"member_count" db:"member_count"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// GroupType represents the type of a group
type GroupType string

const (
	GroupTypePublic  GroupType = "public"
	GroupTypePrivate GroupType = "private"
	GroupTypeSecret  GroupType = "secret"
)

// UserGroupMembership represents a user's membership in a group
type UserGroupMembership struct {
	ID       uuid.UUID        `json:"id" db:"id"`
	UserID   uuid.UUID        `json:"user_id" db:"user_id"`
	GroupID  uuid.UUID        `json:"group_id" db:"group_id"`
	Role     GroupMemberRole  `json:"role" db:"role"`
	Status   MembershipStatus `json:"status" db:"status"`
	JoinedAt time.Time        `json:"joined_at" db:"joined_at"`
	UpdatedAt time.Time       `json:"updated_at" db:"updated_at"`
}

// GroupMemberRole represents a member's role in a group
type GroupMemberRole string

const (
	GroupMemberRoleMember     GroupMemberRole = "member"
	GroupMemberRoleModerator  GroupMemberRole = "moderator"
	GroupMemberRoleAdmin      GroupMemberRole = "admin"
	GroupMemberRoleOwner      GroupMemberRole = "owner"
)

// MembershipStatus represents the status of a group membership
type MembershipStatus string

const (
	MembershipStatusPending  MembershipStatus = "pending"
	MembershipStatusActive   MembershipStatus = "active"
	MembershipStatusInactive MembershipStatus = "inactive"
	MembershipStatusBanned   MembershipStatus = "banned"
)

// UserSession represents a user session
type UserSession struct {
	ID        uuid.UUID `json:"id" db:"id"`
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	Token     string    `json:"token" db:"token"`
	IPAddress string    `json:"ip_address" db:"ip_address"`
	UserAgent string    `json:"user_agent" db:"user_agent"`
	Location  string    `json:"location" db:"location"`
	Device    string    `json:"device" db:"device"`
	IsActive  bool      `json:"is_active" db:"is_active"`
	LastUsed  time.Time `json:"last_used" db:"last_used"`
	ExpiresAt time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// UserConnection represents a connection between users (friends, followers, etc.)
type UserConnection struct {
	ID           uuid.UUID        `json:"id" db:"id"`
	RequesterID  uuid.UUID        `json:"requester_id" db:"requester_id"`
	AddresseeID  uuid.UUID        `json:"addressee_id" db:"addressee_id"`
	Type         ConnectionType   `json:"type" db:"type"`
	Status       ConnectionStatus `json:"status" db:"status"`
	CreatedAt    time.Time        `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time        `json:"updated_at" db:"updated_at"`
}

// ConnectionType represents the type of connection between users
type ConnectionType string

const (
	ConnectionTypeFriend   ConnectionType = "friend"
	ConnectionTypeFollower ConnectionType = "follower"
	ConnectionTypeBlocked  ConnectionType = "blocked"
)

// ConnectionStatus represents the status of a connection
type ConnectionStatus string

const (
	ConnectionStatusPending  ConnectionStatus = "pending"
	ConnectionStatusAccepted ConnectionStatus = "accepted"
	ConnectionStatusRejected ConnectionStatus = "rejected"
	ConnectionStatusBlocked  ConnectionStatus = "blocked"
)