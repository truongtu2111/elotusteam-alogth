package domain

import (
	"context"
	"github.com/google/uuid"
)

// UserRepository defines the interface for user data operations
type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByUsername(ctx context.Context, username string) (*User, error)
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*User, error)
	Search(ctx context.Context, query string, limit, offset int) ([]*User, error)
	GetByRole(ctx context.Context, role UserRole, limit, offset int) ([]*User, error)
	GetByStatus(ctx context.Context, status UserStatus, limit, offset int) ([]*User, error)
	Count(ctx context.Context) (int64, error)
	CountByStatus(ctx context.Context, status UserStatus) (int64, error)
	ExistsByEmail(ctx context.Context, email string) (bool, error)
	ExistsByUsername(ctx context.Context, username string) (bool, error)
	UpdateLastLogin(ctx context.Context, userID uuid.UUID) error
	UpdateEmailVerification(ctx context.Context, userID uuid.UUID, verified bool) error
}

// UserProfileRepository defines the interface for user profile operations
type UserProfileRepository interface {
	Create(ctx context.Context, profile *UserProfile) error
	GetByUserID(ctx context.Context, userID uuid.UUID) (*UserProfile, error)
	Update(ctx context.Context, profile *UserProfile) error
	Delete(ctx context.Context, userID uuid.UUID) error
	UpdatePreferences(ctx context.Context, userID uuid.UUID, preferences UserPreferences) error
}

// UserGroupRepository defines the interface for user group operations
type UserGroupRepository interface {
	Create(ctx context.Context, group *UserGroup) error
	GetByID(ctx context.Context, id uuid.UUID) (*UserGroup, error)
	GetByName(ctx context.Context, name string) (*UserGroup, error)
	Update(ctx context.Context, group *UserGroup) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*UserGroup, error)
	GetByOwner(ctx context.Context, ownerID uuid.UUID) ([]*UserGroup, error)
	GetPublicGroups(ctx context.Context, limit, offset int) ([]*UserGroup, error)
	Search(ctx context.Context, query string, limit, offset int) ([]*UserGroup, error)
	UpdateMemberCount(ctx context.Context, groupID uuid.UUID, count int) error
}

// UserGroupMembershipRepository defines the interface for group membership operations
type UserGroupMembershipRepository interface {
	Create(ctx context.Context, membership *UserGroupMembership) error
	GetByID(ctx context.Context, id uuid.UUID) (*UserGroupMembership, error)
	GetByUserAndGroup(ctx context.Context, userID, groupID uuid.UUID) (*UserGroupMembership, error)
	Update(ctx context.Context, membership *UserGroupMembership) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]*UserGroupMembership, error)
	GetByGroupID(ctx context.Context, groupID uuid.UUID, limit, offset int) ([]*UserGroupMembership, error)
	GetByGroupIDAndRole(ctx context.Context, groupID uuid.UUID, role GroupMemberRole) ([]*UserGroupMembership, error)
	GetByGroupIDAndStatus(ctx context.Context, groupID uuid.UUID, status MembershipStatus) ([]*UserGroupMembership, error)
	CountByGroupID(ctx context.Context, groupID uuid.UUID) (int64, error)
	CountByUserID(ctx context.Context, userID uuid.UUID) (int64, error)
	ExistsByUserAndGroup(ctx context.Context, userID, groupID uuid.UUID) (bool, error)
}

// UserSessionRepository defines the interface for user session operations
type UserSessionRepository interface {
	Create(ctx context.Context, session *UserSession) error
	GetByID(ctx context.Context, id uuid.UUID) (*UserSession, error)
	GetByToken(ctx context.Context, token string) (*UserSession, error)
	Update(ctx context.Context, session *UserSession) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]*UserSession, error)
	GetActiveSessions(ctx context.Context, userID uuid.UUID) ([]*UserSession, error)
	DeleteByUserID(ctx context.Context, userID uuid.UUID) error
	DeleteExpiredSessions(ctx context.Context) error
	UpdateLastUsed(ctx context.Context, sessionID uuid.UUID) error
	DeactivateSession(ctx context.Context, sessionID uuid.UUID) error
}

// UserConnectionRepository defines the interface for user connection operations
type UserConnectionRepository interface {
	Create(ctx context.Context, connection *UserConnection) error
	GetByID(ctx context.Context, id uuid.UUID) (*UserConnection, error)
	GetByUsers(ctx context.Context, requesterID, addresseeID uuid.UUID) (*UserConnection, error)
	Update(ctx context.Context, connection *UserConnection) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByRequesterID(ctx context.Context, requesterID uuid.UUID, connectionType ConnectionType) ([]*UserConnection, error)
	GetByAddresseeID(ctx context.Context, addresseeID uuid.UUID, connectionType ConnectionType) ([]*UserConnection, error)
	GetFriends(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*UserConnection, error)
	GetFollowers(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*UserConnection, error)
	GetFollowing(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*UserConnection, error)
	GetPendingRequests(ctx context.Context, userID uuid.UUID, connectionType ConnectionType) ([]*UserConnection, error)
	CountFriends(ctx context.Context, userID uuid.UUID) (int64, error)
	CountFollowers(ctx context.Context, userID uuid.UUID) (int64, error)
	CountFollowing(ctx context.Context, userID uuid.UUID) (int64, error)
	ExistsByUsers(ctx context.Context, requesterID, addresseeID uuid.UUID, connectionType ConnectionType) (bool, error)
}

// RepositoryManager aggregates all user-related repositories
type RepositoryManager interface {
	User() UserRepository
	UserProfile() UserProfileRepository
	UserGroup() UserGroupRepository
	UserGroupMembership() UserGroupMembershipRepository
	UserSession() UserSessionRepository
	UserConnection() UserConnectionRepository
	BeginTx(ctx context.Context) (RepositoryManager, error)
	Commit() error
	Rollback() error
}