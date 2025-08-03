package infrastructure

import (
	"context"
	"fmt"
	"sync"
	"time"

	authDomain "github.com/elotusteam/microservice-project/services/auth/domain"
	sharedDomain "github.com/elotusteam/microservice-project/shared/domain"
	"github.com/elotusteam/microservice-project/shared/data"
)

// CacheItem represents a cached item with expiration
type CacheItem struct {
	Value     interface{}
	ExpiresAt time.Time
}

// MockRepositoryManager implements authDomain.RepositoryManager for testing/demo purposes
type MockRepositoryManager struct {
	users         map[string]*sharedDomain.User
	sessions      map[string]*sharedDomain.Session
	revokedTokens map[string]*sharedDomain.RevokedToken
	loginAttempts map[string][]*authDomain.LoginAttempt
	passwordResets map[string]*authDomain.PasswordResetToken
	activityLogs  map[string]*sharedDomain.ActivityLog
	cache         map[string]CacheItem
	mu            sync.RWMutex
}

// NewMockRepositoryManager creates a new mock repository manager
func NewMockRepositoryManager() authDomain.RepositoryManager {
	return &MockRepositoryManager{
		users:         make(map[string]*sharedDomain.User),
		sessions:      make(map[string]*sharedDomain.Session),
		revokedTokens: make(map[string]*sharedDomain.RevokedToken),
		loginAttempts: make(map[string][]*authDomain.LoginAttempt),
		passwordResets: make(map[string]*authDomain.PasswordResetToken),
		activityLogs:  make(map[string]*sharedDomain.ActivityLog),
		cache:         make(map[string]CacheItem),
	}
}

// GetUserRepository returns the user repository
func (m *MockRepositoryManager) GetUserRepository() authDomain.UserRepository {
	return &MockUserRepository{manager: m}
}

// GetSessionRepository returns the session repository
func (m *MockRepositoryManager) GetSessionRepository() authDomain.SessionRepository {
	return &MockSessionRepository{manager: m}
}

// GetRevokedTokenRepository returns the revoked token repository
func (m *MockRepositoryManager) GetRevokedTokenRepository() authDomain.RevokedTokenRepository {
	return &MockRevokedTokenRepository{manager: m}
}

// GetLoginAttemptRepository returns the login attempt repository
func (m *MockRepositoryManager) GetLoginAttemptRepository() authDomain.LoginAttemptRepository {
	return &MockLoginAttemptRepository{manager: m}
}

// GetPasswordResetTokenRepository returns the password reset repository
func (m *MockRepositoryManager) GetPasswordResetTokenRepository() authDomain.PasswordResetTokenRepository {
	return &MockPasswordResetRepository{manager: m}
}

// GetActivityLogRepository returns the activity log repository
func (m *MockRepositoryManager) GetActivityLogRepository() authDomain.ActivityLogRepository {
	return &MockActivityLogRepository{manager: m}
}

// GetCacheRepository returns the cache repository
func (m *MockRepositoryManager) GetCacheRepository() authDomain.AuthCacheRepository {
	return &MockCacheRepository{manager: m}
}

// BeginTransaction starts a new transaction
func (m *MockRepositoryManager) BeginTransaction(ctx context.Context) (data.Transaction, error) {
	return &MockTransaction{ctx: ctx}, nil // Mock implementation
}

// WithTransaction executes a function within a transaction
func (m *MockRepositoryManager) WithTransaction(ctx context.Context, fn func(tx data.Transaction) error) error {
	tx, err := m.BeginTransaction(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	
	if err := fn(tx); err != nil {
		return err
	}
	
	return tx.Commit()
}

// Close closes all repository connections
func (m *MockRepositoryManager) Close() error {
	return nil // Mock implementation
}

// Health checks the health of all repositories
func (m *MockRepositoryManager) Health(ctx context.Context) error {
	return nil // Mock implementation
}

// MockTransaction implements data.Transaction
type MockTransaction struct {
	ctx context.Context
}

func (t *MockTransaction) Commit() error {
	return nil
}

func (t *MockTransaction) Rollback() error {
	return nil
}

func (t *MockTransaction) Context() context.Context {
	if t.ctx != nil {
		return t.ctx
	}
	return context.Background()
}

// MockUserRepository implements domain.UserRepository
type MockUserRepository struct {
	manager *MockRepositoryManager
}

func (r *MockUserRepository) Create(ctx context.Context, user *sharedDomain.User) error {
	r.manager.mu.Lock()
	defer r.manager.mu.Unlock()
	r.manager.users[user.ID] = user
	return nil
}

func (r *MockUserRepository) GetByID(ctx context.Context, id string) (*sharedDomain.User, error) {
	r.manager.mu.RLock()
	defer r.manager.mu.RUnlock()
	user, exists := r.manager.users[id]
	if !exists {
		return nil, fmt.Errorf("user not found")
	}
	return user, nil
}

func (r *MockUserRepository) GetByEmail(ctx context.Context, email string) (*sharedDomain.User, error) {
	r.manager.mu.RLock()
	defer r.manager.mu.RUnlock()
	for _, user := range r.manager.users {
		if user.Email == email {
			return user, nil
		}
	}
	return nil, fmt.Errorf("user not found")
}

func (r *MockUserRepository) GetByUsername(ctx context.Context, username string) (*sharedDomain.User, error) {
	r.manager.mu.RLock()
	defer r.manager.mu.RUnlock()
	for _, user := range r.manager.users {
		if user.Username == username {
			return user, nil
		}
	}
	return nil, fmt.Errorf("user not found")
}

func (r *MockUserRepository) Update(ctx context.Context, user *sharedDomain.User) error {
	r.manager.mu.Lock()
	defer r.manager.mu.Unlock()
	r.manager.users[user.ID] = user
	return nil
}

func (r *MockUserRepository) Delete(ctx context.Context, id string) error {
	r.manager.mu.Lock()
	defer r.manager.mu.Unlock()
	delete(r.manager.users, id)
	return nil
}

func (r *MockUserRepository) List(ctx context.Context, pagination *data.Pagination) (*data.PaginatedResult, error) {
	r.manager.mu.RLock()
	defer r.manager.mu.RUnlock()
	
	users := make([]*sharedDomain.User, 0, len(r.manager.users))
	for _, user := range r.manager.users {
		users = append(users, user)
	}
	
	return &data.PaginatedResult{
		Data:  users,
		Total: int64(len(users)),
	}, nil
}

func (r *MockUserRepository) Count(ctx context.Context) (int64, error) {
	r.manager.mu.RLock()
	defer r.manager.mu.RUnlock()
	return int64(len(r.manager.users)), nil
}

func (r *MockUserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	r.manager.mu.RLock()
	defer r.manager.mu.RUnlock()
	for _, user := range r.manager.users {
		if user.Email == email {
			return true, nil
		}
	}
	return false, nil
}

func (r *MockUserRepository) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	r.manager.mu.RLock()
	defer r.manager.mu.RUnlock()
	for _, user := range r.manager.users {
		if user.Username == username {
			return true, nil
		}
	}
	return false, nil
}

func (r *MockUserRepository) Search(ctx context.Context, criteria map[string]interface{}, pagination *data.Pagination) (*data.PaginatedResult, error) {
	r.manager.mu.RLock()
	defer r.manager.mu.RUnlock()
	
	users := make([]*sharedDomain.User, 0, len(r.manager.users))
	for _, user := range r.manager.users {
		users = append(users, user)
	}
	
	return &data.PaginatedResult{
		Data:  users,
		Total: int64(len(users)),
	}, nil
}

func (r *MockUserRepository) GetUsersByRole(ctx context.Context, role sharedDomain.UserRole, pagination *data.Pagination) (*data.PaginatedResult, error) {
	r.manager.mu.RLock()
	defer r.manager.mu.RUnlock()
	var users []*sharedDomain.User
	for _, user := range r.manager.users {
		if user.Role == role {
			users = append(users, user)
		}
	}
	totalPages := (len(users) + int(pagination.PageSize) - 1) / int(pagination.PageSize)
	return &data.PaginatedResult{
		Data:       users,
		Total:      int64(len(users)),
		Page:       pagination.Page,
		PageSize:   pagination.PageSize,
		TotalPages: totalPages,
		HasNext:    pagination.Page < totalPages,
		HasPrev:    pagination.Page > 1,
	}, nil
}

func (r *MockUserRepository) GetUsersByStatus(ctx context.Context, status sharedDomain.UserStatus, pagination *data.Pagination) (*data.PaginatedResult, error) {
	r.manager.mu.RLock()
	defer r.manager.mu.RUnlock()
	var users []*sharedDomain.User
	for _, user := range r.manager.users {
		if user.Status == status {
			users = append(users, user)
		}
	}
	totalPages := (len(users) + int(pagination.PageSize) - 1) / int(pagination.PageSize)
	return &data.PaginatedResult{
		Data:       users,
		Total:      int64(len(users)),
		Page:       pagination.Page,
		PageSize:   pagination.PageSize,
		TotalPages: totalPages,
		HasNext:    pagination.Page < totalPages,
		HasPrev:    pagination.Page > 1,
	}, nil
}

func (r *MockUserRepository) UpdateLastLogin(ctx context.Context, id string, lastLogin time.Time) error {
	r.manager.mu.Lock()
	defer r.manager.mu.Unlock()
	if user, exists := r.manager.users[id]; exists {
		user.LastLoginAt = &lastLogin
		return nil
	}
	return fmt.Errorf("user not found")
}

func (r *MockUserRepository) UpdatePassword(ctx context.Context, userID string, hashedPassword string) error {
	r.manager.mu.Lock()
	defer r.manager.mu.Unlock()
	if user, exists := r.manager.users[userID]; exists {
		user.PasswordHash = hashedPassword
	}
	return nil
}

func (r *MockUserRepository) ActivateUser(ctx context.Context, userID string) error {
	r.manager.mu.Lock()
	defer r.manager.mu.Unlock()
	if user, exists := r.manager.users[userID]; exists {
		user.Status = sharedDomain.UserStatusActive
	}
	return nil
}

func (r *MockUserRepository) DeactivateUser(ctx context.Context, userID string) error {
	r.manager.mu.Lock()
	defer r.manager.mu.Unlock()
	if user, exists := r.manager.users[userID]; exists {
		user.Status = sharedDomain.UserStatusInactive
	}
	return nil
}

func (r *MockUserRepository) CountUsers(ctx context.Context) (int64, error) {
	r.manager.mu.RLock()
	defer r.manager.mu.RUnlock()
	return int64(len(r.manager.users)), nil
}



func (r *MockUserRepository) Health(ctx context.Context) error {
	return nil
}

func (r *MockUserRepository) Close() error {
	return nil
}

func (r *MockUserRepository) GetActiveUserCount(ctx context.Context) (int64, error) {
	r.manager.mu.RLock()
	defer r.manager.mu.RUnlock()
	count := int64(0)
	for _, user := range r.manager.users {
		if user.Status == sharedDomain.UserStatusActive {
			count++
		}
	}
	return count, nil
}



// MockSessionRepository implements domain.SessionRepository
type MockSessionRepository struct {
	manager *MockRepositoryManager
}

func (r *MockSessionRepository) Create(ctx context.Context, session *sharedDomain.Session) error {
	r.manager.mu.Lock()
	defer r.manager.mu.Unlock()
	r.manager.sessions[session.ID] = session
	return nil
}

func (r *MockSessionRepository) GetByID(ctx context.Context, id string) (*sharedDomain.Session, error) {
	r.manager.mu.RLock()
	defer r.manager.mu.RUnlock()
	session, exists := r.manager.sessions[id]
	if !exists {
		return nil, fmt.Errorf("session not found")
	}
	return session, nil
}

func (r *MockSessionRepository) GetByTokenID(ctx context.Context, tokenID string) (*sharedDomain.Session, error) {
	r.manager.mu.RLock()
	defer r.manager.mu.RUnlock()
	for _, session := range r.manager.sessions {
		if session.TokenID == tokenID {
			return session, nil
		}
	}
	return nil, fmt.Errorf("session not found")
}

func (r *MockSessionRepository) GetByUserID(ctx context.Context, userID string) ([]*sharedDomain.Session, error) {
	r.manager.mu.RLock()
	defer r.manager.mu.RUnlock()
	var sessions []*sharedDomain.Session
	for _, session := range r.manager.sessions {
		if session.UserID == userID {
			sessions = append(sessions, session)
		}
	}
	return sessions, nil
}

func (r *MockSessionRepository) Update(ctx context.Context, session *sharedDomain.Session) error {
	r.manager.mu.Lock()
	defer r.manager.mu.Unlock()
	r.manager.sessions[session.ID] = session
	return nil
}

func (r *MockSessionRepository) Delete(ctx context.Context, id string) error {
	r.manager.mu.Lock()
	defer r.manager.mu.Unlock()
	delete(r.manager.sessions, id)
	return nil
}

func (r *MockSessionRepository) DeleteByTokenID(ctx context.Context, tokenID string) error {
	r.manager.mu.Lock()
	defer r.manager.mu.Unlock()
	for id, session := range r.manager.sessions {
		if session.TokenID == tokenID {
			delete(r.manager.sessions, id)
			break
		}
	}
	return nil
}

func (r *MockSessionRepository) DeleteByUserID(ctx context.Context, userID string) error {
	r.manager.mu.Lock()
	defer r.manager.mu.Unlock()
	for id, session := range r.manager.sessions {
		if session.UserID == userID {
			delete(r.manager.sessions, id)
		}
	}
	return nil
}

func (r *MockSessionRepository) Health(ctx context.Context) error {
	return nil
}

func (r *MockSessionRepository) Close() error {
	return nil
}

func (r *MockSessionRepository) UpdateLastUsed(ctx context.Context, sessionID string, lastUsed time.Time) error {
	r.manager.mu.Lock()
	defer r.manager.mu.Unlock()
	for _, session := range r.manager.sessions {
		if session.ID == sessionID {
			session.LastUsedAt = lastUsed
			break
		}
	}
	return nil
}

func (r *MockSessionRepository) DeleteExpired(ctx context.Context) error {
	r.manager.mu.Lock()
	defer r.manager.mu.Unlock()
	// Mock implementation - remove expired sessions
	now := time.Now()
	for id, session := range r.manager.sessions {
		if session.ExpiresAt.Before(now) {
			delete(r.manager.sessions, id)
		}
	}
	return nil
}

func (r *MockSessionRepository) GetActiveSessionCount(ctx context.Context) (int64, error) {
	r.manager.mu.RLock()
	defer r.manager.mu.RUnlock()
	now := time.Now()
	count := int64(0)
	for _, session := range r.manager.sessions {
		if session.ExpiresAt.After(now) {
			count++
		}
	}
	return count, nil
}

func (r *MockSessionRepository) GetActiveSessionsForUser(ctx context.Context, userID string) (int, error) {
	r.manager.mu.RLock()
	defer r.manager.mu.RUnlock()
	now := time.Now()
	count := 0
	for _, session := range r.manager.sessions {
		if session.UserID == userID && session.ExpiresAt.After(now) {
			count++
		}
	}
	return count, nil
}

func (r *MockSessionRepository) GetActiveSessions(ctx context.Context, userID string) ([]*sharedDomain.Session, error) {
	r.manager.mu.RLock()
	defer r.manager.mu.RUnlock()
	var sessions []*sharedDomain.Session
	for _, session := range r.manager.sessions {
		if session.UserID == userID && session.Status == sharedDomain.SessionStatusActive && time.Now().Before(session.ExpiresAt) {
			sessions = append(sessions, session)
		}
	}
	return sessions, nil
}

func (r *MockSessionRepository) RevokeSession(ctx context.Context, sessionID string) error {
	r.manager.mu.Lock()
	defer r.manager.mu.Unlock()
	for _, session := range r.manager.sessions {
		if session.ID == sessionID {
			session.Status = sharedDomain.SessionStatusRevoked
			break
		}
	}
	return nil
}

func (r *MockSessionRepository) RevokeAllUserSessions(ctx context.Context, userID string) error {
	r.manager.mu.Lock()
	defer r.manager.mu.Unlock()
	for _, session := range r.manager.sessions {
		if session.UserID == userID {
			session.Status = sharedDomain.SessionStatusRevoked
		}
	}
	return nil
}

// MockRevokedTokenRepository implements domain.RevokedTokenRepository
type MockRevokedTokenRepository struct {
	manager *MockRepositoryManager
}

func (r *MockRevokedTokenRepository) Create(ctx context.Context, token *sharedDomain.RevokedToken) error {
	r.manager.mu.Lock()
	defer r.manager.mu.Unlock()
	r.manager.revokedTokens[token.TokenID] = token
	return nil
}

func (r *MockRevokedTokenRepository) GetByTokenID(ctx context.Context, tokenID string) (*sharedDomain.RevokedToken, error) {
	r.manager.mu.RLock()
	defer r.manager.mu.RUnlock()
	token, exists := r.manager.revokedTokens[tokenID]
	if !exists {
		return nil, fmt.Errorf("revoked token not found")
	}
	return token, nil
}

func (r *MockRevokedTokenRepository) GetByUserID(ctx context.Context, userID string) ([]*sharedDomain.RevokedToken, error) {
	r.manager.mu.RLock()
	defer r.manager.mu.RUnlock()
	var tokens []*sharedDomain.RevokedToken
	for _, token := range r.manager.revokedTokens {
		if token.UserID == userID {
			tokens = append(tokens, token)
		}
	}
	return tokens, nil
}

func (r *MockRevokedTokenRepository) GetRevokedTokenCount(ctx context.Context) (int64, error) {
	r.manager.mu.RLock()
	defer r.manager.mu.RUnlock()
	return int64(len(r.manager.revokedTokens)), nil
}

func (r *MockRevokedTokenRepository) RevokeToken(ctx context.Context, tokenID, userID, reason string, expiresAt time.Time) error {
	r.manager.mu.Lock()
	defer r.manager.mu.Unlock()
	token := &sharedDomain.RevokedToken{
		TokenID:   tokenID,
		UserID:    userID,
		Reason:    reason,
		RevokedAt: time.Now(),
		ExpiresAt: expiresAt,
	}
	r.manager.revokedTokens[tokenID] = token
	return nil
}

func (r *MockRevokedTokenRepository) IsRevoked(ctx context.Context, tokenID string) (bool, error) {
	r.manager.mu.RLock()
	defer r.manager.mu.RUnlock()
	token, exists := r.manager.revokedTokens[tokenID]
	if !exists {
		return false, nil
	}
	// Check if token is expired
	if !token.ExpiresAt.IsZero() && time.Now().After(token.ExpiresAt) {
		return false, nil
	}
	return true, nil
}

func (r *MockRevokedTokenRepository) RevokeAllUserTokens(ctx context.Context, userID, reason string) error {
	r.manager.mu.Lock()
	defer r.manager.mu.Unlock()
	// In a real implementation, this would revoke all tokens for a user
	// For mock, we'll just add a placeholder entry
	r.manager.revokedTokens["all_"+userID] = &sharedDomain.RevokedToken{
		TokenID:   "all_" + userID,
		UserID:    userID,
		Reason:    reason,
		RevokedAt: time.Now(),
	}
	return nil
}

func (r *MockRevokedTokenRepository) Health(ctx context.Context) error {
	return nil
}

func (r *MockRevokedTokenRepository) Close() error {
	return nil
}

func (r *MockRevokedTokenRepository) DeleteExpired(ctx context.Context) error {
	r.manager.mu.Lock()
	defer r.manager.mu.Unlock()
	now := time.Now()
	for tokenID, token := range r.manager.revokedTokens {
		if !token.ExpiresAt.IsZero() && token.ExpiresAt.Before(now) {
			delete(r.manager.revokedTokens, tokenID)
		}
	}
	return nil
}

func (r *MockRevokedTokenRepository) CleanupExpired(ctx context.Context) (int64, error) {
	r.manager.mu.Lock()
	defer r.manager.mu.Unlock()
	now := time.Now()
	count := int64(0)
	for tokenID, token := range r.manager.revokedTokens {
		if !token.ExpiresAt.IsZero() && token.ExpiresAt.Before(now) {
			delete(r.manager.revokedTokens, tokenID)
			count++
		}
	}
	return count, nil
}

// MockLoginAttemptRepository implements domain.LoginAttemptRepository
type MockLoginAttemptRepository struct {
	manager *MockRepositoryManager
}

func (r *MockLoginAttemptRepository) Create(ctx context.Context, attempt *authDomain.LoginAttempt) error {
	r.manager.mu.Lock()
	defer r.manager.mu.Unlock()
	r.manager.loginAttempts[attempt.Identifier] = append(r.manager.loginAttempts[attempt.Identifier], attempt)
	return nil
}

func (r *MockLoginAttemptRepository) CountFailedAttempts(ctx context.Context, identifier string, since time.Time) (int64, error) {
	r.manager.mu.RLock()
	defer r.manager.mu.RUnlock()
	var count int64
	for _, attempt := range r.manager.loginAttempts[identifier] {
		if !attempt.Success && attempt.Timestamp.After(since) {
			count++
		}
	}
	return count, nil
}

func (r *MockLoginAttemptRepository) CountFailedAttemptsByUserID(ctx context.Context, userID string, since time.Time) (int64, error) {
	r.manager.mu.RLock()
	defer r.manager.mu.RUnlock()
	var count int64
	for _, attempts := range r.manager.loginAttempts {
		for _, attempt := range attempts {
			if attempt.Identifier == userID && !attempt.Success && attempt.Timestamp.After(since) {
				count++
			}
		}
	}
	return count, nil
}

func (r *MockLoginAttemptRepository) Health(ctx context.Context) error {
	return nil
}

func (r *MockLoginAttemptRepository) Close() error {
	return nil
}

func (r *MockLoginAttemptRepository) CountAttemptsByIP(ctx context.Context, ipAddress string, since time.Time) (int64, error) {
	r.manager.mu.RLock()
	defer r.manager.mu.RUnlock()
	var count int64
	for _, attempts := range r.manager.loginAttempts {
		for _, attempt := range attempts {
			if attempt.IPAddress == ipAddress && attempt.Timestamp.After(since) {
				count++
			}
		}
	}
	return count, nil
}

func (r *MockLoginAttemptRepository) DeleteOldAttempts(ctx context.Context, before time.Time) error {
	r.manager.mu.Lock()
	defer r.manager.mu.Unlock()
	for identifier, attempts := range r.manager.loginAttempts {
		var filteredAttempts []*authDomain.LoginAttempt
		for _, attempt := range attempts {
			if !attempt.Timestamp.Before(before) {
				filteredAttempts = append(filteredAttempts, attempt)
			}
		}
		r.manager.loginAttempts[identifier] = filteredAttempts
	}
	return nil
}

func (r *MockLoginAttemptRepository) CleanupOldAttempts(ctx context.Context, maxAge time.Duration) (int64, error) {
	r.manager.mu.Lock()
	defer r.manager.mu.Unlock()
	cutoff := time.Now().Add(-maxAge)
	var deletedCount int64
	for identifier, attempts := range r.manager.loginAttempts {
		var filtered []*authDomain.LoginAttempt
		for _, attempt := range attempts {
			if attempt.Timestamp.After(cutoff) {
				filtered = append(filtered, attempt)
			} else {
				deletedCount++
			}
		}
		r.manager.loginAttempts[identifier] = filtered
	}
	return deletedCount, nil
}

// MockPasswordResetRepository implements domain.PasswordResetRepository
type MockPasswordResetRepository struct {
	manager *MockRepositoryManager
}

func (r *MockPasswordResetRepository) Create(ctx context.Context, token *authDomain.PasswordResetToken) error {
	r.manager.mu.Lock()
	defer r.manager.mu.Unlock()
	r.manager.passwordResets[token.Token] = token
	return nil
}

func (r *MockPasswordResetRepository) GetByToken(ctx context.Context, token string) (*authDomain.PasswordResetToken, error) {
	r.manager.mu.RLock()
	defer r.manager.mu.RUnlock()
	resetToken, exists := r.manager.passwordResets[token]
	if !exists {
		return nil, fmt.Errorf("token not found")
	}
	return resetToken, nil
}

func (r *MockPasswordResetRepository) MarkAsUsed(ctx context.Context, token string) error {
	r.manager.mu.Lock()
	defer r.manager.mu.Unlock()
	if resetToken, exists := r.manager.passwordResets[token]; exists {
		resetToken.Used = true
		return nil
	}
	return fmt.Errorf("token not found")
}

func (r *MockPasswordResetRepository) DeleteExpired(ctx context.Context) error {
	r.manager.mu.Lock()
	defer r.manager.mu.Unlock()
	now := time.Now()
	for token, resetToken := range r.manager.passwordResets {
		if resetToken.ExpiresAt.Before(now) {
			delete(r.manager.passwordResets, token)
		}
	}
	return nil
}

func (r *MockPasswordResetRepository) Health(ctx context.Context) error {
	return nil
}

func (r *MockPasswordResetRepository) Close() error {
	return nil
}

func (r *MockPasswordResetRepository) Delete(ctx context.Context, token string) error {
	r.manager.mu.Lock()
	defer r.manager.mu.Unlock()
	delete(r.manager.passwordResets, token)
	return nil
}

func (r *MockPasswordResetRepository) DeleteByUserID(ctx context.Context, userID string) error {
	r.manager.mu.Lock()
	defer r.manager.mu.Unlock()
	for token, resetToken := range r.manager.passwordResets {
		if resetToken.UserID == userID {
			delete(r.manager.passwordResets, token)
		}
	}
	return nil
}

func (r *MockPasswordResetRepository) CleanupExpired(ctx context.Context) (int64, error) {
	r.manager.mu.Lock()
	defer r.manager.mu.Unlock()
	now := time.Now()
	var deletedCount int64
	for token, resetToken := range r.manager.passwordResets {
		if resetToken.ExpiresAt.Before(now) {
			delete(r.manager.passwordResets, token)
			deletedCount++
		}
	}
	return deletedCount, nil
}

func (r *MockPasswordResetRepository) GetActiveTokensCount(ctx context.Context, userID string) (int64, error) {
	r.manager.mu.RLock()
	defer r.manager.mu.RUnlock()
	now := time.Now()
	var count int64
	for _, resetToken := range r.manager.passwordResets {
		if resetToken.UserID == userID && now.Before(resetToken.ExpiresAt) {
			count++
		}
	}
	return count, nil
}

func (r *MockPasswordResetRepository) GetByUserID(ctx context.Context, userID string) ([]*authDomain.PasswordResetToken, error) {
	r.manager.mu.RLock()
	defer r.manager.mu.RUnlock()
	var tokens []*authDomain.PasswordResetToken
	for _, resetToken := range r.manager.passwordResets {
		if resetToken.UserID == userID {
			tokens = append(tokens, resetToken)
		}
	}
	return tokens, nil
}

// MockActivityLogRepository implements domain.ActivityLogRepository
type MockActivityLogRepository struct {
	manager *MockRepositoryManager
}

func (r *MockActivityLogRepository) Create(ctx context.Context, log *sharedDomain.ActivityLog) error {
	r.manager.mu.Lock()
	defer r.manager.mu.Unlock()
	r.manager.activityLogs[log.ID] = log
	return nil
}

func (r *MockActivityLogRepository) GetByUserID(ctx context.Context, userID string, pagination *data.Pagination) (*data.PaginatedResult, error) {
	r.manager.mu.RLock()
	defer r.manager.mu.RUnlock()
	var filtered []*sharedDomain.ActivityLog
	for _, log := range r.manager.activityLogs {
		if log.UserID != nil && *log.UserID == userID {
			filtered = append(filtered, log)
		}
	}
	// Apply pagination
	start := (pagination.Page - 1) * pagination.PageSize
	end := start + pagination.PageSize
	if start > len(filtered) {
		start = len(filtered)
	}
	if end > len(filtered) {
		end = len(filtered)
	}
	paginatedData := filtered[start:end]
	totalPages := (len(filtered) + pagination.PageSize - 1) / pagination.PageSize
	return &data.PaginatedResult{
		Data:       paginatedData,
		Total:      int64(len(filtered)),
		Page:       pagination.Page,
		PageSize:   pagination.PageSize,
		TotalPages: totalPages,
		HasNext:    pagination.Page < totalPages,
		HasPrev:    pagination.Page > 1,
	}, nil
}

func (r *MockActivityLogRepository) Health(ctx context.Context) error {
	return nil
}

func (r *MockActivityLogRepository) Close() error {
	return nil
}

func (r *MockActivityLogRepository) CleanupOldLogs(ctx context.Context, olderThan time.Duration) (int64, error) {
	r.manager.mu.Lock()
	defer r.manager.mu.Unlock()
	// Mock implementation - remove old logs
	cutoff := time.Now().Add(-olderThan)
	deleted := int64(0)
	for id, log := range r.manager.activityLogs {
		if log.Timestamp.Before(cutoff) {
			delete(r.manager.activityLogs, id)
			deleted++
		}
	}
	return deleted, nil
}

func (r *MockActivityLogRepository) GetByAction(ctx context.Context, action string, pagination *data.Pagination) (*data.PaginatedResult, error) {
	r.manager.mu.RLock()
	defer r.manager.mu.RUnlock()
	var filtered []*sharedDomain.ActivityLog
	for _, log := range r.manager.activityLogs {
		if log.Action == action {
			filtered = append(filtered, log)
		}
	}
	// Apply pagination
	start := (pagination.Page - 1) * pagination.PageSize
	end := start + pagination.PageSize
	if start > len(filtered) {
		start = len(filtered)
	}
	if end > len(filtered) {
		end = len(filtered)
	}
	paginatedData := filtered[start:end]
	totalPages := (len(filtered) + pagination.PageSize - 1) / pagination.PageSize
	return &data.PaginatedResult{
		Data:       paginatedData,
		Total:      int64(len(filtered)),
		Page:       pagination.Page,
		PageSize:   pagination.PageSize,
		TotalPages: totalPages,
		HasNext:    pagination.Page < totalPages,
		HasPrev:    pagination.Page > 1,
	}, nil
}

func (r *MockActivityLogRepository) GetByResourceType(ctx context.Context, resourceType string, pagination *data.Pagination) (*data.PaginatedResult, error) {
	r.manager.mu.RLock()
	defer r.manager.mu.RUnlock()
	var filtered []*sharedDomain.ActivityLog
	for _, log := range r.manager.activityLogs {
		if log.ResourceType == resourceType {
			filtered = append(filtered, log)
		}
	}
	// Apply pagination
	start := (pagination.Page - 1) * pagination.PageSize
	end := start + pagination.PageSize
	if start > len(filtered) {
		start = len(filtered)
	}
	if end > len(filtered) {
		end = len(filtered)
	}
	paginatedData := filtered[start:end]
	totalPages := (len(filtered) + pagination.PageSize - 1) / pagination.PageSize
	return &data.PaginatedResult{
		Data:       paginatedData,
		Total:      int64(len(filtered)),
		Page:       pagination.Page,
		PageSize:   pagination.PageSize,
		TotalPages: totalPages,
		HasNext:    pagination.Page < totalPages,
		HasPrev:    pagination.Page > 1,
	}, nil
}

func (r *MockActivityLogRepository) GetByTimeRange(ctx context.Context, startTime, endTime time.Time, pagination *data.Pagination) (*data.PaginatedResult, error) {
	r.manager.mu.RLock()
	defer r.manager.mu.RUnlock()
	var filtered []*sharedDomain.ActivityLog
	for _, log := range r.manager.activityLogs {
		if log.Timestamp.After(startTime) && log.Timestamp.Before(endTime) {
			filtered = append(filtered, log)
		}
	}
	// Apply pagination
	start := (pagination.Page - 1) * pagination.PageSize
	end := start + pagination.PageSize
	if start > len(filtered) {
		start = len(filtered)
	}
	if end > len(filtered) {
		end = len(filtered)
	}
	paginatedData := filtered[start:end]
	totalPages := (len(filtered) + pagination.PageSize - 1) / pagination.PageSize
	return &data.PaginatedResult{
		Data:       paginatedData,
		Total:      int64(len(filtered)),
		Page:       pagination.Page,
		PageSize:   pagination.PageSize,
		TotalPages: totalPages,
		HasNext:    pagination.Page < totalPages,
		HasPrev:    pagination.Page > 1,
	}, nil
}

func (r *MockActivityLogRepository) GetSecurityEvents(ctx context.Context, since time.Time, pagination *data.Pagination) (*data.PaginatedResult, error) {
	r.manager.mu.RLock()
	defer r.manager.mu.RUnlock()
	var filtered []*sharedDomain.ActivityLog
	for _, log := range r.manager.activityLogs {
		// Filter for security-related events since the specified time
		if log.Timestamp.After(since) && (log.Action == "login" || log.Action == "logout" || log.Action == "password_change" || log.Action == "account_locked" || log.Action == "suspicious_activity") {
			filtered = append(filtered, log)
		}
	}
	// Apply pagination
	start := (pagination.Page - 1) * pagination.PageSize
	end := start + pagination.PageSize
	if start > len(filtered) {
		start = len(filtered)
	}
	if end > len(filtered) {
		end = len(filtered)
	}
	paginatedData := filtered[start:end]
	totalPages := (len(filtered) + pagination.PageSize - 1) / pagination.PageSize
	return &data.PaginatedResult{
		Data:       paginatedData,
		Total:      int64(len(filtered)),
		Page:       pagination.Page,
		PageSize:   pagination.PageSize,
		TotalPages: totalPages,
		HasNext:    pagination.Page < totalPages,
		HasPrev:    pagination.Page > 1,
	}, nil
}

func (r *MockActivityLogRepository) GetUserActivity(ctx context.Context, userID string, pagination *data.Pagination) (*data.PaginatedResult, error) {
	r.manager.mu.RLock()
	defer r.manager.mu.RUnlock()
	var filtered []*sharedDomain.ActivityLog
	for _, log := range r.manager.activityLogs {
		if log.UserID == userID {
			filtered = append(filtered, log)
		}
	}
	// Apply pagination
	start := (pagination.Page - 1) * pagination.PageSize
	end := start + pagination.PageSize
	if start > len(filtered) {
		start = len(filtered)
	}
	if end > len(filtered) {
		end = len(filtered)
	}
	paginatedData := filtered[start:end]
	totalPages := (len(filtered) + pagination.PageSize - 1) / pagination.PageSize
	return &data.PaginatedResult{
		Data:       paginatedData,
		Total:      int64(len(filtered)),
		Page:       pagination.Page,
		PageSize:   pagination.PageSize,
		TotalPages: totalPages,
		HasNext:    pagination.Page < totalPages,
		HasPrev:    pagination.Page > 1,
	}, nil
}

func (r *MockActivityLogRepository) DeleteOldLogs(ctx context.Context, before time.Time) error {
	r.manager.mu.Lock()
	defer r.manager.mu.Unlock()
	for id, log := range r.manager.activityLogs {
		if log.Timestamp.Before(before) {
			delete(r.manager.activityLogs, id)
		}
	}
	return nil
}

// MockCacheRepository implements domain.CacheRepository
type MockCacheRepository struct {
	manager *MockRepositoryManager
}

func (r *MockCacheRepository) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	r.manager.mu.Lock()
	defer r.manager.mu.Unlock()
	r.manager.cache[key] = CacheItem{
		Value:     value,
		ExpiresAt: time.Now().Add(ttl),
	}
	return nil
}

func (r *MockCacheRepository) Get(ctx context.Context, key string) ([]byte, error) {
	r.manager.mu.RLock()
	defer r.manager.mu.RUnlock()
	item, exists := r.manager.cache[key]
	if !exists {
		return nil, fmt.Errorf("key not found")
	}
	if time.Now().After(item.ExpiresAt) {
		delete(r.manager.cache, key)
		return nil, fmt.Errorf("key expired")
	}
	if bytes, ok := item.Value.([]byte); ok {
		return bytes, nil
	}
	return nil, fmt.Errorf("value is not bytes")
}

func (r *MockCacheRepository) Delete(ctx context.Context, key string) error {
	r.manager.mu.Lock()
	defer r.manager.mu.Unlock()
	delete(r.manager.cache, key)
	return nil
}

func (r *MockCacheRepository) SetUserSession(ctx context.Context, sessionID string, user *authDomain.AuthUser, ttl time.Duration) error {
	r.manager.mu.Lock()
	defer r.manager.mu.Unlock()
	r.manager.cache["session:"+sessionID] = CacheItem{
		Value:     user,
		ExpiresAt: time.Now().Add(ttl),
	}
	return nil
}

func (r *MockCacheRepository) GetUserSession(ctx context.Context, sessionID string) (*authDomain.AuthUser, error) {
	r.manager.mu.RLock()
	defer r.manager.mu.RUnlock()
	item, exists := r.manager.cache["session:"+sessionID]
	if !exists {
		return nil, fmt.Errorf("session not found")
	}
	if time.Now().After(item.ExpiresAt) {
		delete(r.manager.cache, "session:"+sessionID)
		return nil, fmt.Errorf("session expired")
	}
	user, ok := item.Value.(*authDomain.AuthUser)
	if !ok {
		return nil, fmt.Errorf("invalid session data")
	}
	return user, nil
}

func (r *MockCacheRepository) DeleteUserSession(ctx context.Context, sessionID string) error {
	return r.Delete(ctx, "session:"+sessionID)
}

func (r *MockCacheRepository) SetLoginAttempts(ctx context.Context, identifier string, count int, ttl time.Duration) error {
	r.manager.mu.Lock()
	defer r.manager.mu.Unlock()
	r.manager.cache["login_attempts:"+identifier] = CacheItem{
		Value:     count,
		ExpiresAt: time.Now().Add(ttl),
	}
	return nil
}

func (r *MockCacheRepository) GetLoginAttempts(ctx context.Context, identifier string) (int, error) {
	r.manager.mu.RLock()
	defer r.manager.mu.RUnlock()
	item, exists := r.manager.cache["login_attempts:"+identifier]
	if !exists {
		return 0, nil
	}
	if time.Now().After(item.ExpiresAt) {
		delete(r.manager.cache, "login_attempts:"+identifier)
		return 0, nil
	}
	count, ok := item.Value.(int)
	if !ok {
		return 0, fmt.Errorf("invalid count data")
	}
	return count, nil
}

func (r *MockCacheRepository) IncrementLoginAttempts(ctx context.Context, identifier string, ttl time.Duration) (int, error) {
	count, _ := r.GetLoginAttempts(ctx, identifier)
	count++
	r.SetLoginAttempts(ctx, identifier, count, ttl)
	return count, nil
}

func (r *MockCacheRepository) ResetLoginAttempts(ctx context.Context, identifier string) error {
	return r.Delete(ctx, "login_attempts:"+identifier)
}

func (r *MockCacheRepository) Health(ctx context.Context) error {
	return nil
}

func (r *MockCacheRepository) Expire(ctx context.Context, key string, expiration time.Duration) error {
	r.manager.mu.Lock()
	defer r.manager.mu.Unlock()
	if item, exists := r.manager.cache[key]; exists {
		item.ExpiresAt = time.Now().Add(expiration)
		r.manager.cache[key] = item
	}
	return nil
}

func (r *MockCacheRepository) Keys(ctx context.Context, pattern string) ([]string, error) {
	r.manager.mu.RLock()
	defer r.manager.mu.RUnlock()
	keys := make([]string, 0, len(r.manager.cache))
	for key := range r.manager.cache {
		keys = append(keys, key)
	}
	return keys, nil
}

func (r *MockCacheRepository) FlushAll(ctx context.Context) error {
	r.manager.mu.Lock()
	defer r.manager.mu.Unlock()
	r.manager.cache = make(map[string]CacheItem)
	return nil
}

func (r *MockCacheRepository) SetNX(ctx context.Context, key string, value []byte, expiration time.Duration) (bool, error) {
	r.manager.mu.Lock()
	defer r.manager.mu.Unlock()
	if _, exists := r.manager.cache[key]; exists {
		return false, nil
	}
	r.manager.cache[key] = CacheItem{
		Value:     value,
		ExpiresAt: time.Now().Add(expiration),
	}
	return true, nil
}

func (r *MockCacheRepository) Close() error {
	return nil
}

func (r *MockCacheRepository) Increment(ctx context.Context, key string) (int64, error) {
	r.manager.mu.Lock()
	defer r.manager.mu.Unlock()
	value, exists := r.manager.cache[key]
	if !exists {
		r.manager.cache[key] = CacheItem{Value: int64(1), ExpiresAt: time.Now().Add(time.Hour)}
		return 1, nil
	}
	if intVal, ok := value.Value.(int64); ok {
		newVal := intVal + 1
		r.manager.cache[key] = CacheItem{Value: newVal, ExpiresAt: value.ExpiresAt}
		return newVal, nil
	}
	return 0, fmt.Errorf("value is not an integer")
}

func (r *MockCacheRepository) Decrement(ctx context.Context, key string) (int64, error) {
	r.manager.mu.Lock()
	defer r.manager.mu.Unlock()
	value, exists := r.manager.cache[key]
	if !exists {
		r.manager.cache[key] = CacheItem{Value: int64(-1), ExpiresAt: time.Now().Add(time.Hour)}
		return -1, nil
	}
	if intVal, ok := value.Value.(int64); ok {
		newVal := intVal - 1
		r.manager.cache[key] = CacheItem{Value: newVal, ExpiresAt: value.ExpiresAt}
		return newVal, nil
	}
	return 0, fmt.Errorf("value is not an integer")
}

func (r *MockCacheRepository) SetPasswordResetToken(ctx context.Context, token string, userID string, ttl time.Duration) error {
	r.manager.mu.Lock()
	defer r.manager.mu.Unlock()
	r.manager.cache["password_reset:"+token] = CacheItem{
		Value:     userID,
		ExpiresAt: time.Now().Add(ttl),
	}
	return nil
}

func (r *MockCacheRepository) GetPasswordResetToken(ctx context.Context, token string) (string, error) {
	r.manager.mu.RLock()
	defer r.manager.mu.RUnlock()
	item, exists := r.manager.cache["password_reset:"+token]
	if !exists {
		return "", fmt.Errorf("token not found")
	}
	if time.Now().After(item.ExpiresAt) {
		delete(r.manager.cache, "password_reset:"+token)
		return "", fmt.Errorf("token expired")
	}
	userID, ok := item.Value.(string)
	if !ok {
		return "", fmt.Errorf("invalid token data")
	}
	return userID, nil
}

func (r *MockCacheRepository) DeletePasswordResetToken(ctx context.Context, token string) error {
	return r.Delete(ctx, "password_reset:"+token)
}

func (r *MockCacheRepository) SetRevokedToken(ctx context.Context, tokenID string, ttl time.Duration) error {
	r.manager.mu.Lock()
	defer r.manager.mu.Unlock()
	r.manager.cache["revoked:"+tokenID] = CacheItem{
		Value:     true,
		ExpiresAt: time.Now().Add(ttl),
	}
	return nil
}

func (r *MockCacheRepository) IsTokenRevoked(ctx context.Context, tokenID string) (bool, error) {
	r.manager.mu.RLock()
	defer r.manager.mu.RUnlock()
	item, exists := r.manager.cache["revoked:"+tokenID]
	if !exists {
		return false, nil
	}
	if time.Now().After(item.ExpiresAt) {
		delete(r.manager.cache, "revoked:"+tokenID)
		return false, nil
	}
	return true, nil
}

func (r *MockCacheRepository) SetUserLockout(ctx context.Context, userID string, ttl time.Duration) error {
	r.manager.mu.Lock()
	defer r.manager.mu.Unlock()
	r.manager.cache["lockout:"+userID] = CacheItem{
		Value:     true,
		ExpiresAt: time.Now().Add(ttl),
	}
	return nil
}

func (r *MockCacheRepository) IsUserLockedOut(ctx context.Context, userID string) (bool, error) {
	r.manager.mu.RLock()
	defer r.manager.mu.RUnlock()
	item, exists := r.manager.cache["lockout:"+userID]
	if !exists {
		return false, nil
	}
	if time.Now().After(item.ExpiresAt) {
		delete(r.manager.cache, "lockout:"+userID)
		return false, nil
	}
	return true, nil
}

func (r *MockCacheRepository) RemoveUserLockout(ctx context.Context, userID string) error {
	return r.Delete(ctx, "lockout:"+userID)
}

func (r *MockCacheRepository) SetRateLimitCounter(ctx context.Context, key string, count int, ttl time.Duration) error {
	r.manager.mu.Lock()
	defer r.manager.mu.Unlock()
	r.manager.cache["rate:"+key] = CacheItem{
		Value:     count,
		ExpiresAt: time.Now().Add(ttl),
	}
	return nil
}

func (r *MockCacheRepository) GetRateLimitCounter(ctx context.Context, key string) (int, error) {
	r.manager.mu.RLock()
	defer r.manager.mu.RUnlock()
	item, exists := r.manager.cache["rate:"+key]
	if !exists {
		return 0, nil
	}
	if time.Now().After(item.ExpiresAt) {
		delete(r.manager.cache, "rate:"+key)
		return 0, nil
	}
	count, ok := item.Value.(int)
	if !ok {
		return 0, fmt.Errorf("invalid count data")
	}
	return count, nil
}

func (r *MockCacheRepository) IncrementRateLimitCounter(ctx context.Context, key string, ttl time.Duration) (int, error) {
	count, _ := r.GetRateLimitCounter(ctx, key)
	count++
	r.SetRateLimitCounter(ctx, key, count, ttl)
	return count, nil
}

func (r *MockCacheRepository) Exists(ctx context.Context, key string) (bool, error) {
	r.manager.mu.RLock()
	defer r.manager.mu.RUnlock()
	item, exists := r.manager.cache[key]
	if !exists {
		return false, nil
	}
	if time.Now().After(item.ExpiresAt) {
		delete(r.manager.cache, key)
		return false, nil
	}
	return true, nil
}