package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	fileDomain "github.com/elotusteam/microservice-project/services/file/domain"
	"github.com/elotusteam/microservice-project/shared/data"
	"github.com/google/uuid"
)

// PostgreSQLRepositoryManager implements fileDomain.RepositoryManager
type PostgreSQLRepositoryManager struct {
	db            data.DatabaseConnection
	tx            data.Transaction
	file          *PostgreSQLFileRepository
	uploadSession *PostgreSQLUploadSessionRepository
	fileShare     *PostgreSQLFileShareRepository
	fileVersion   *PostgreSQLFileVersionRepository
	imageVariant  *PostgreSQLImageVariantRepository
}

// NewPostgreSQLRepositoryManager creates a new PostgreSQL repository manager
func NewPostgreSQLRepositoryManager(db data.DatabaseConnection) fileDomain.RepositoryManager {
	return &PostgreSQLRepositoryManager{
		db: db,
	}
}

// File returns the file repository
func (rm *PostgreSQLRepositoryManager) File() fileDomain.FileRepository {
	return &PostgreSQLFileRepository{db: rm.getDB()}
}

// UploadSession returns the upload session repository
func (rm *PostgreSQLRepositoryManager) UploadSession() fileDomain.UploadSessionRepository {
	return &PostgreSQLUploadSessionRepository{db: rm.getDB()}
}

// FileShare returns the file share repository
func (rm *PostgreSQLRepositoryManager) FileShare() fileDomain.FileShareRepository {
	return &PostgreSQLFileShareRepository{db: rm.getDB()}
}

// FileVersion returns the file version repository
func (rm *PostgreSQLRepositoryManager) FileVersion() fileDomain.FileVersionRepository {
	return &PostgreSQLFileVersionRepository{db: rm.getDB()}
}

// ImageVariant returns the image variant repository
func (rm *PostgreSQLRepositoryManager) ImageVariant() fileDomain.ImageVariantRepository {
	return &PostgreSQLImageVariantRepository{db: rm.getDB()}
}

// BeginTx starts a new transaction
func (rm *PostgreSQLRepositoryManager) BeginTx(ctx context.Context) (fileDomain.RepositoryManager, error) {
	tx, err := rm.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	return &PostgreSQLRepositoryManager{
		db: rm.db,
		tx: tx,
	}, nil
}

// Commit commits the transaction
func (rm *PostgreSQLRepositoryManager) Commit() error {
	if rm.tx == nil {
		return fmt.Errorf("no active transaction")
	}
	return rm.tx.Commit()
}

// Rollback rolls back the transaction
func (rm *PostgreSQLRepositoryManager) Rollback() error {
	if rm.tx == nil {
		return fmt.Errorf("no active transaction")
	}
	return rm.tx.Rollback()
}

// getDB returns the appropriate database connection (transaction or regular)
func (rm *PostgreSQLRepositoryManager) getDB() data.DatabaseConnection {
	if rm.tx != nil {
		// Return the transaction which already implements DatabaseConnection
		return rm.tx.(data.DatabaseConnection)
	}
	return rm.db
}

// PostgreSQLFileRepository implements fileDomain.FileRepository
type PostgreSQLFileRepository struct {
	db data.DatabaseConnection
}

// Create creates a new file record
func (r *PostgreSQLFileRepository) Create(ctx context.Context, file *fileDomain.File) error {
	metadataJSON, err := json.Marshal(file.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	query := `
		INSERT INTO files (id, user_id, filename, original_name, mime_type, size, path, url, checksum, status, metadata, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`

	_, err = r.db.Exec(ctx, query,
		file.ID, file.UserID, file.Filename, file.OriginalName, file.MimeType,
		file.Size, file.Path, file.URL, file.Checksum, file.Status,
		metadataJSON, file.CreatedAt, file.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}

	return nil
}

// GetByID retrieves a file by ID
func (r *PostgreSQLFileRepository) GetByID(ctx context.Context, id uuid.UUID) (*fileDomain.File, error) {
	query := `
		SELECT id, user_id, filename, original_name, mime_type, size, path, url, checksum, status, metadata, created_at, updated_at, deleted_at
		FROM files
		WHERE id = $1 AND deleted_at IS NULL
	`

	row := r.db.QueryRow(ctx, query, id)
	return r.scanFile(row)
}

// GetByUserID retrieves files by user ID with pagination
func (r *PostgreSQLFileRepository) GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*fileDomain.File, error) {
	query := `
		SELECT id, user_id, filename, original_name, mime_type, size, path, url, checksum, status, metadata, created_at, updated_at, deleted_at
		FROM files
		WHERE user_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query files: %w", err)
	}
	defer rows.Close()

	var files []*fileDomain.File
	for rows.Next() {
		file, err := r.scanFile(rows)
		if err != nil {
			return nil, err
		}
		files = append(files, file)
	}

	return files, nil
}

// Update updates a file record
func (r *PostgreSQLFileRepository) Update(ctx context.Context, file *fileDomain.File) error {
	metadataJSON, err := json.Marshal(file.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	query := `
		UPDATE files
		SET filename = $2, original_name = $3, mime_type = $4, size = $5, path = $6, url = $7,
		    checksum = $8, status = $9, metadata = $10, updated_at = $11
		WHERE id = $1 AND deleted_at IS NULL
	`

	_, err = r.db.Exec(ctx, query,
		file.ID, file.Filename, file.OriginalName, file.MimeType, file.Size,
		file.Path, file.URL, file.Checksum, file.Status, metadataJSON, time.Now(),
	)

	if err != nil {
		return fmt.Errorf("failed to update file: %w", err)
	}

	return nil
}

// Delete soft deletes a file record
func (r *PostgreSQLFileRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE files SET deleted_at = $1 WHERE id = $2 AND deleted_at IS NULL`
	_, err := r.db.Exec(ctx, query, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}

// GetByChecksum retrieves a file by checksum
func (r *PostgreSQLFileRepository) GetByChecksum(ctx context.Context, checksum string) (*fileDomain.File, error) {
	query := `
		SELECT id, user_id, filename, original_name, mime_type, size, path, url, checksum, status, metadata, created_at, updated_at, deleted_at
		FROM files
		WHERE checksum = $1 AND deleted_at IS NULL
		LIMIT 1
	`

	row := r.db.QueryRow(ctx, query, checksum)
	return r.scanFile(row)
}

// Search searches files by query
func (r *PostgreSQLFileRepository) Search(ctx context.Context, query string, userID uuid.UUID, limit, offset int) ([]*fileDomain.File, error) {
	sqlQuery := `
		SELECT id, user_id, filename, original_name, mime_type, size, path, url, checksum, status, metadata, created_at, updated_at, deleted_at
		FROM files
		WHERE user_id = $1 AND deleted_at IS NULL
		  AND (filename ILIKE $2 OR original_name ILIKE $2)
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4
	`

	searchPattern := "%" + query + "%"
	rows, err := r.db.Query(ctx, sqlQuery, userID, searchPattern, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to search files: %w", err)
	}
	defer rows.Close()

	var files []*fileDomain.File
	for rows.Next() {
		file, err := r.scanFile(rows)
		if err != nil {
			return nil, err
		}
		files = append(files, file)
	}

	return files, nil
}

// GetTotalSize gets total size of files for a user
func (r *PostgreSQLFileRepository) GetTotalSize(ctx context.Context, userID uuid.UUID) (int64, error) {
	query := `SELECT COALESCE(SUM(size), 0) FROM files WHERE user_id = $1 AND deleted_at IS NULL`
	var totalSize int64
	err := r.db.QueryRow(ctx, query, userID).Scan(&totalSize)
	if err != nil {
		return 0, fmt.Errorf("failed to get total size: %w", err)
	}
	return totalSize, nil
}

// GetFileCount gets file count for a user
func (r *PostgreSQLFileRepository) GetFileCount(ctx context.Context, userID uuid.UUID) (int64, error) {
	query := `SELECT COUNT(*) FROM files WHERE user_id = $1 AND deleted_at IS NULL`
	var count int64
	err := r.db.QueryRow(ctx, query, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get file count: %w", err)
	}
	return count, nil
}

// scanFile scans a database row into a File struct
func (r *PostgreSQLFileRepository) scanFile(row data.Row) (*fileDomain.File, error) {
	var file fileDomain.File
	var metadataJSON []byte
	var deletedAt sql.NullTime

	err := row.Scan(
		&file.ID, &file.UserID, &file.Filename, &file.OriginalName, &file.MimeType,
		&file.Size, &file.Path, &file.URL, &file.Checksum, &file.Status,
		&metadataJSON, &file.CreatedAt, &file.UpdatedAt, &deletedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("file not found")
		}
		return nil, fmt.Errorf("failed to scan file: %w", err)
	}

	if deletedAt.Valid {
		file.DeletedAt = &deletedAt.Time
	}

	if len(metadataJSON) > 0 {
		err = json.Unmarshal(metadataJSON, &file.Metadata)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
	}

	return &file, nil
}

// PostgreSQLUploadSessionRepository implements fileDomain.UploadSessionRepository
type PostgreSQLUploadSessionRepository struct {
	db data.DatabaseConnection
}

// Create creates a new upload session
func (r *PostgreSQLUploadSessionRepository) Create(ctx context.Context, session *fileDomain.UploadSession) error {
	query := `
		INSERT INTO upload_sessions (id, user_id, filename, total_size, chunk_size, uploaded_size, status, expires_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err := r.db.Exec(ctx, query,
		session.ID, session.UserID, session.Filename, session.TotalSize, session.ChunkSize,
		session.UploadedSize, session.Status, session.ExpiresAt, session.CreatedAt, session.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create upload session: %w", err)
	}

	return nil
}

// GetByID retrieves an upload session by ID
func (r *PostgreSQLUploadSessionRepository) GetByID(ctx context.Context, id uuid.UUID) (*fileDomain.UploadSession, error) {
	query := `
		SELECT id, user_id, filename, total_size, chunk_size, uploaded_size, status, expires_at, created_at, updated_at
		FROM upload_sessions
		WHERE id = $1
	`

	row := r.db.QueryRow(ctx, query, id)
	return r.scanUploadSession(row)
}

// GetByUserID retrieves upload sessions by user ID
func (r *PostgreSQLUploadSessionRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*fileDomain.UploadSession, error) {
	query := `
		SELECT id, user_id, filename, total_size, chunk_size, uploaded_size, status, expires_at, created_at, updated_at
		FROM upload_sessions
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query upload sessions: %w", err)
	}
	defer rows.Close()

	var sessions []*fileDomain.UploadSession
	for rows.Next() {
		session, err := r.scanUploadSession(rows)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, session)
	}

	return sessions, nil
}

// Update updates an upload session
func (r *PostgreSQLUploadSessionRepository) Update(ctx context.Context, session *fileDomain.UploadSession) error {
	query := `
		UPDATE upload_sessions
		SET uploaded_size = $2, status = $3, updated_at = $4
		WHERE id = $1
	`

	_, err := r.db.Exec(ctx, query, session.ID, session.UploadedSize, session.Status, time.Now())
	if err != nil {
		return fmt.Errorf("failed to update upload session: %w", err)
	}

	return nil
}

// Delete deletes an upload session
func (r *PostgreSQLUploadSessionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM upload_sessions WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete upload session: %w", err)
	}
	return nil
}

// DeleteExpired deletes expired upload sessions
func (r *PostgreSQLUploadSessionRepository) DeleteExpired(ctx context.Context) error {
	query := `DELETE FROM upload_sessions WHERE expires_at < $1`
	_, err := r.db.Exec(ctx, query, time.Now())
	if err != nil {
		return fmt.Errorf("failed to delete expired sessions: %w", err)
	}
	return nil
}

// GetActiveByUserID retrieves active upload sessions by user ID
func (r *PostgreSQLUploadSessionRepository) GetActiveByUserID(ctx context.Context, userID uuid.UUID) ([]*fileDomain.UploadSession, error) {
	query := `
		SELECT id, user_id, filename, total_size, chunk_size, uploaded_size, status, expires_at, created_at, updated_at
		FROM upload_sessions
		WHERE user_id = $1 AND status = 'active' AND expires_at > $2
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query, userID, time.Now())
	if err != nil {
		return nil, fmt.Errorf("failed to query active sessions: %w", err)
	}
	defer rows.Close()

	var sessions []*fileDomain.UploadSession
	for rows.Next() {
		session, err := r.scanUploadSession(rows)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, session)
	}

	return sessions, nil
}

// scanUploadSession scans a database row into an UploadSession struct
func (r *PostgreSQLUploadSessionRepository) scanUploadSession(row data.Row) (*fileDomain.UploadSession, error) {
	var session fileDomain.UploadSession

	err := row.Scan(
		&session.ID, &session.UserID, &session.Filename, &session.TotalSize, &session.ChunkSize,
		&session.UploadedSize, &session.Status, &session.ExpiresAt, &session.CreatedAt, &session.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("upload session not found")
		}
		return nil, fmt.Errorf("failed to scan upload session: %w", err)
	}

	return &session, nil
}

// PostgreSQLFileShareRepository implements fileDomain.FileShareRepository
type PostgreSQLFileShareRepository struct {
	db data.DatabaseConnection
}

// Create creates a new file share
func (r *PostgreSQLFileShareRepository) Create(ctx context.Context, share *fileDomain.FileShare) error {
	permissionsJSON, err := json.Marshal(share.Permissions)
	if err != nil {
		return fmt.Errorf("failed to marshal permissions: %w", err)
	}

	query := `
		INSERT INTO file_shares (id, file_id, shared_by, shared_with, share_token, permissions, expires_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err = r.db.Exec(ctx, query,
		share.ID, share.FileID, share.SharedBy, share.SharedWith, share.ShareToken,
		permissionsJSON, share.ExpiresAt, share.CreatedAt, share.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create file share: %w", err)
	}

	return nil
}

// GetByID retrieves a file share by ID
func (r *PostgreSQLFileShareRepository) GetByID(ctx context.Context, id uuid.UUID) (*fileDomain.FileShare, error) {
	query := `
		SELECT id, file_id, shared_by, shared_with, share_token, permissions, expires_at, created_at, updated_at
		FROM file_shares
		WHERE id = $1
	`

	row := r.db.QueryRow(ctx, query, id)
	return r.scanFileShare(row)
}

// GetByToken retrieves a file share by token
func (r *PostgreSQLFileShareRepository) GetByToken(ctx context.Context, token string) (*fileDomain.FileShare, error) {
	query := `
		SELECT id, file_id, shared_by, shared_with, share_token, permissions, expires_at, created_at, updated_at
		FROM file_shares
		WHERE share_token = $1 AND (expires_at IS NULL OR expires_at > $2)
	`

	row := r.db.QueryRow(ctx, query, token, time.Now())
	return r.scanFileShare(row)
}

// GetByFileID retrieves file shares by file ID
func (r *PostgreSQLFileShareRepository) GetByFileID(ctx context.Context, fileID uuid.UUID) ([]*fileDomain.FileShare, error) {
	query := `
		SELECT id, file_id, shared_by, shared_with, share_token, permissions, expires_at, created_at, updated_at
		FROM file_shares
		WHERE file_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query, fileID)
	if err != nil {
		return nil, fmt.Errorf("failed to query file shares: %w", err)
	}
	defer rows.Close()

	var shares []*fileDomain.FileShare
	for rows.Next() {
		share, err := r.scanFileShare(rows)
		if err != nil {
			return nil, err
		}
		shares = append(shares, share)
	}

	return shares, nil
}

// GetBySharedBy retrieves file shares by shared by user ID
func (r *PostgreSQLFileShareRepository) GetBySharedBy(ctx context.Context, userID uuid.UUID) ([]*fileDomain.FileShare, error) {
	query := `
		SELECT id, file_id, shared_by, shared_with, share_token, permissions, expires_at, created_at, updated_at
		FROM file_shares
		WHERE shared_by = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query shared files: %w", err)
	}
	defer rows.Close()

	var shares []*fileDomain.FileShare
	for rows.Next() {
		share, err := r.scanFileShare(rows)
		if err != nil {
			return nil, err
		}
		shares = append(shares, share)
	}

	return shares, nil
}

// GetBySharedWith retrieves file shares by shared with user ID
func (r *PostgreSQLFileShareRepository) GetBySharedWith(ctx context.Context, userID uuid.UUID) ([]*fileDomain.FileShare, error) {
	query := `
		SELECT id, file_id, shared_by, shared_with, share_token, permissions, expires_at, created_at, updated_at
		FROM file_shares
		WHERE shared_with = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query files shared with user: %w", err)
	}
	defer rows.Close()

	var shares []*fileDomain.FileShare
	for rows.Next() {
		share, err := r.scanFileShare(rows)
		if err != nil {
			return nil, err
		}
		shares = append(shares, share)
	}

	return shares, nil
}

// Update updates a file share
func (r *PostgreSQLFileShareRepository) Update(ctx context.Context, share *fileDomain.FileShare) error {
	permissionsJSON, err := json.Marshal(share.Permissions)
	if err != nil {
		return fmt.Errorf("failed to marshal permissions: %w", err)
	}

	query := `
		UPDATE file_shares
		SET permissions = $2, expires_at = $3, updated_at = $4
		WHERE id = $1
	`

	_, err = r.db.Exec(ctx, query, share.ID, permissionsJSON, share.ExpiresAt, time.Now())
	if err != nil {
		return fmt.Errorf("failed to update file share: %w", err)
	}

	return nil
}

// Delete deletes a file share
func (r *PostgreSQLFileShareRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM file_shares WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete file share: %w", err)
	}
	return nil
}

// DeleteExpired deletes expired file shares
func (r *PostgreSQLFileShareRepository) DeleteExpired(ctx context.Context) error {
	query := `DELETE FROM file_shares WHERE expires_at IS NOT NULL AND expires_at < $1`
	_, err := r.db.Exec(ctx, query, time.Now())
	if err != nil {
		return fmt.Errorf("failed to delete expired shares: %w", err)
	}
	return nil
}

// scanFileShare scans a database row into a FileShare struct
func (r *PostgreSQLFileShareRepository) scanFileShare(row data.Row) (*fileDomain.FileShare, error) {
	var share fileDomain.FileShare
	var permissionsJSON []byte
	var sharedWith sql.NullString
	var expiresAt sql.NullTime

	err := row.Scan(
		&share.ID, &share.FileID, &share.SharedBy, &sharedWith, &share.ShareToken,
		&permissionsJSON, &expiresAt, &share.CreatedAt, &share.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("file share not found")
		}
		return nil, fmt.Errorf("failed to scan file share: %w", err)
	}

	if sharedWith.Valid {
		uuid, err := uuid.Parse(sharedWith.String)
		if err != nil {
			return nil, fmt.Errorf("failed to parse shared_with UUID: %w", err)
		}
		share.SharedWith = &uuid
	}

	if expiresAt.Valid {
		share.ExpiresAt = &expiresAt.Time
	}

	if len(permissionsJSON) > 0 {
		err = json.Unmarshal(permissionsJSON, &share.Permissions)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal permissions: %w", err)
		}
	}

	return &share, nil
}

// PostgreSQLFileVersionRepository implements fileDomain.FileVersionRepository
type PostgreSQLFileVersionRepository struct {
	db data.DatabaseConnection
}

// Create creates a new file version
func (r *PostgreSQLFileVersionRepository) Create(ctx context.Context, version *fileDomain.FileVersion) error {
	query := `
		INSERT INTO file_versions (id, file_id, version, path, size, checksum, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := r.db.Exec(ctx, query,
		version.ID, version.FileID, version.Version, version.Path,
		version.Size, version.Checksum, version.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create file version: %w", err)
	}

	return nil
}

// GetByID retrieves a file version by ID
func (r *PostgreSQLFileVersionRepository) GetByID(ctx context.Context, id uuid.UUID) (*fileDomain.FileVersion, error) {
	query := `
		SELECT id, file_id, version, path, size, checksum, created_at
		FROM file_versions
		WHERE id = $1
	`

	row := r.db.QueryRow(ctx, query, id)
	return r.scanFileVersion(row)
}

// GetByFileID retrieves file versions by file ID
func (r *PostgreSQLFileVersionRepository) GetByFileID(ctx context.Context, fileID uuid.UUID) ([]*fileDomain.FileVersion, error) {
	query := `
		SELECT id, file_id, version, path, size, checksum, created_at
		FROM file_versions
		WHERE file_id = $1
		ORDER BY version DESC
	`

	rows, err := r.db.Query(ctx, query, fileID)
	if err != nil {
		return nil, fmt.Errorf("failed to query file versions: %w", err)
	}
	defer rows.Close()

	var versions []*fileDomain.FileVersion
	for rows.Next() {
		version, err := r.scanFileVersion(rows)
		if err != nil {
			return nil, err
		}
		versions = append(versions, version)
	}

	return versions, nil
}

// GetLatestByFileID retrieves the latest file version by file ID
func (r *PostgreSQLFileVersionRepository) GetLatestByFileID(ctx context.Context, fileID uuid.UUID) (*fileDomain.FileVersion, error) {
	query := `
		SELECT id, file_id, version, path, size, checksum, created_at
		FROM file_versions
		WHERE file_id = $1
		ORDER BY version DESC
		LIMIT 1
	`

	row := r.db.QueryRow(ctx, query, fileID)
	return r.scanFileVersion(row)
}

// Delete deletes a file version
func (r *PostgreSQLFileVersionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM file_versions WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete file version: %w", err)
	}
	return nil
}

// DeleteByFileID deletes all versions of a file
func (r *PostgreSQLFileVersionRepository) DeleteByFileID(ctx context.Context, fileID uuid.UUID) error {
	query := `DELETE FROM file_versions WHERE file_id = $1`
	_, err := r.db.Exec(ctx, query, fileID)
	if err != nil {
		return fmt.Errorf("failed to delete file versions: %w", err)
	}
	return nil
}

// scanFileVersion scans a database row into a FileVersion struct
func (r *PostgreSQLFileVersionRepository) scanFileVersion(row data.Row) (*fileDomain.FileVersion, error) {
	var version fileDomain.FileVersion

	err := row.Scan(
		&version.ID, &version.FileID, &version.Version, &version.Path,
		&version.Size, &version.Checksum, &version.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("file version not found")
		}
		return nil, fmt.Errorf("failed to scan file version: %w", err)
	}

	return &version, nil
}

// Database connection wrappers to implement data.DatabaseConnection interface

type dbWrapper struct {
	db *sql.DB
}

func (w *dbWrapper) Query(ctx context.Context, query string, args ...interface{}) (data.Rows, error) {
	rows, err := w.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return &rowsWrapper{rows: rows}, nil
}

func (w *dbWrapper) QueryRow(ctx context.Context, query string, args ...interface{}) data.Row {
	return &rowWrapper{row: w.db.QueryRowContext(ctx, query, args...)}
}

func (w *dbWrapper) Exec(ctx context.Context, query string, args ...interface{}) (data.Result, error) {
	result, err := w.db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return &resultWrapper{result: result}, nil
}

func (w *dbWrapper) Prepare(ctx context.Context, query string) (data.Statement, error) {
	stmt, err := w.db.PrepareContext(ctx, query)
	if err != nil {
		return nil, err
	}
	return &stmtWrapper{stmt: stmt}, nil
}

func (w *dbWrapper) Begin(ctx context.Context) (data.Transaction, error) {
	tx, err := w.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	return &txWrapper{tx: tx}, nil
}

func (w *dbWrapper) Ping(ctx context.Context) error {
	return w.db.PingContext(ctx)
}

func (w *dbWrapper) Close() error {
	return w.db.Close()
}

type txWrapper struct {
	tx *sql.Tx
}

func (w *txWrapper) Query(ctx context.Context, query string, args ...interface{}) (data.Rows, error) {
	rows, err := w.tx.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return &rowsWrapper{rows: rows}, nil
}

func (w *txWrapper) QueryRow(ctx context.Context, query string, args ...interface{}) data.Row {
	return &rowWrapper{row: w.tx.QueryRowContext(ctx, query, args...)}
}

func (w *txWrapper) Exec(ctx context.Context, query string, args ...interface{}) (data.Result, error) {
	result, err := w.tx.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return &resultWrapper{result: result}, nil
}

func (w *txWrapper) Prepare(ctx context.Context, query string) (data.Statement, error) {
	stmt, err := w.tx.PrepareContext(ctx, query)
	if err != nil {
		return nil, err
	}
	return &stmtWrapper{stmt: stmt}, nil
}

func (w *txWrapper) Begin(ctx context.Context) (data.Transaction, error) {
	return nil, fmt.Errorf("cannot begin transaction within transaction")
}

func (w *txWrapper) Ping(ctx context.Context) error {
	return fmt.Errorf("ping not supported on transaction")
}

func (w *txWrapper) Close() error {
	return fmt.Errorf("close not supported on transaction")
}

func (w *txWrapper) Commit() error {
	return w.tx.Commit()
}

func (w *txWrapper) Rollback() error {
	return w.tx.Rollback()
}

func (w *txWrapper) Context() context.Context {
	return context.Background()
}

type rowsWrapper struct {
	rows *sql.Rows
}

func (w *rowsWrapper) Next() bool {
	return w.rows.Next()
}

func (w *rowsWrapper) Scan(dest ...interface{}) error {
	return w.rows.Scan(dest...)
}

func (w *rowsWrapper) Close() error {
	return w.rows.Close()
}

func (w *rowsWrapper) Err() error {
	return w.rows.Err()
}

func (w *rowsWrapper) Columns() ([]string, error) {
	return w.rows.Columns()
}

type rowWrapper struct {
	row *sql.Row
}

func (w *rowWrapper) Scan(dest ...interface{}) error {
	return w.row.Scan(dest...)
}

type resultWrapper struct {
	result sql.Result
}

func (w *resultWrapper) LastInsertId() (int64, error) {
	return w.result.LastInsertId()
}

func (w *resultWrapper) RowsAffected() (int64, error) {
	return w.result.RowsAffected()
}

type stmtWrapper struct {
	stmt *sql.Stmt
}

func (w *stmtWrapper) Exec(ctx context.Context, args ...interface{}) (data.Result, error) {
	result, err := w.stmt.ExecContext(ctx, args...)
	if err != nil {
		return nil, err
	}
	return &resultWrapper{result: result}, nil
}

func (w *stmtWrapper) Query(ctx context.Context, args ...interface{}) (data.Rows, error) {
	rows, err := w.stmt.QueryContext(ctx, args...)
	if err != nil {
		return nil, err
	}
	return &rowsWrapper{rows: rows}, nil
}

func (w *stmtWrapper) QueryRow(ctx context.Context, args ...interface{}) data.Row {
	return &rowWrapper{row: w.stmt.QueryRowContext(ctx, args...)}
}

func (w *stmtWrapper) Close() error {
	return w.stmt.Close()
}

// PostgreSQLImageVariantRepository implements fileDomain.ImageVariantRepository
type PostgreSQLImageVariantRepository struct {
	db data.DatabaseConnection
}

// Create creates a new image variant
func (r *PostgreSQLImageVariantRepository) Create(ctx context.Context, variant *fileDomain.ImageVariant) error {
	query := `
		INSERT INTO image_variants (id, file_id, variant_type, width, height, size, path, format, quality, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`

	_, err := r.db.Exec(ctx, query,
		variant.ID, variant.FileID, variant.VariantType, variant.Width, variant.Height,
		variant.Size, variant.Path, variant.Format, variant.Quality, variant.Status, variant.CreatedAt, variant.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create image variant: %w", err)
	}

	return nil
}

// GetByID retrieves an image variant by ID
func (r *PostgreSQLImageVariantRepository) GetByID(ctx context.Context, id string) (*fileDomain.ImageVariant, error) {
	query := `
		SELECT id, file_id, variant_type, width, height, size, path, format, quality, status, created_at, updated_at
		FROM image_variants
		WHERE id = $1
	`

	row := r.db.QueryRow(ctx, query, id)
	return r.scanImageVariant(row)
}

// GetByFileID retrieves image variants by file ID
func (r *PostgreSQLImageVariantRepository) GetByFileID(ctx context.Context, fileID uuid.UUID) ([]*fileDomain.ImageVariant, error) {
	query := `
		SELECT id, file_id, variant_type, width, height, size, path, format, quality, status, created_at, updated_at
		FROM image_variants
		WHERE file_id = $1
		ORDER BY created_at ASC
	`

	rows, err := r.db.Query(ctx, query, fileID)
	if err != nil {
		return nil, fmt.Errorf("failed to query image variants: %w", err)
	}
	defer rows.Close()

	var variants []*fileDomain.ImageVariant
	for rows.Next() {
		variant, err := r.scanImageVariant(rows)
		if err != nil {
			return nil, err
		}
		variants = append(variants, variant)
	}

	return variants, nil
}

// Update updates an image variant
func (r *PostgreSQLImageVariantRepository) Update(ctx context.Context, variant *fileDomain.ImageVariant) error {
	query := `
		UPDATE image_variants 
		SET variant_type = $2, width = $3, height = $4, size = $5, path = $6, format = $7, quality = $8, status = $9, updated_at = $10
		WHERE id = $1
	`

	_, err := r.db.Exec(ctx, query,
		variant.ID, variant.VariantType, variant.Width, variant.Height,
		variant.Size, variant.Path, variant.Format, variant.Quality, variant.Status, variant.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update image variant: %w", err)
	}

	return nil
}

// Delete deletes an image variant
func (r *PostgreSQLImageVariantRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM image_variants WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete image variant: %w", err)
	}
	return nil
}

// DeleteByFileID deletes all variants of a file
func (r *PostgreSQLImageVariantRepository) DeleteByFileID(ctx context.Context, fileID uuid.UUID) error {
	query := `DELETE FROM image_variants WHERE file_id = $1`
	_, err := r.db.Exec(ctx, query, fileID)
	if err != nil {
		return fmt.Errorf("failed to delete image variants: %w", err)
	}
	return nil
}

// scanImageVariant scans a database row into an ImageVariant struct
func (r *PostgreSQLImageVariantRepository) scanImageVariant(row data.Row) (*fileDomain.ImageVariant, error) {
	var variant fileDomain.ImageVariant

	err := row.Scan(
		&variant.ID, &variant.FileID, &variant.VariantType, &variant.Width, &variant.Height,
		&variant.Size, &variant.Path, &variant.Format, &variant.Quality, &variant.Status, &variant.CreatedAt, &variant.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("image variant not found")
		}
		return nil, fmt.Errorf("failed to scan image variant: %w", err)
	}

	return &variant, nil
}
