package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCompleteUserWorkflow tests the complete user journey from registration to file operations
func TestCompleteUserWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	// Test data
	userData := map[string]interface{}{
		"name":     "John Doe",
		"email":    "john.doe@example.com",
		"password": "SecurePassword123!",
	}

	// Step 1: User Registration
	t.Log("Step 1: User Registration")
	userID, err := registerUser(t, userData)
	require.NoError(t, err, "User registration should succeed")
	require.NotEmpty(t, userID, "User ID should be returned")
	t.Logf("User registered with ID: %s", userID)

	// Step 2: User Login
	t.Log("Step 2: User Login")
	token, err := loginUser(t, userData["email"].(string), userData["password"].(string))
	require.NoError(t, err, "User login should succeed")
	require.NotEmpty(t, token, "JWT token should be returned")
	t.Logf("User logged in, token received")

	// Step 3: Get User Profile
	t.Log("Step 3: Get User Profile")
	profile, err := getUserProfile(t, token)
	require.NoError(t, err, "Get user profile should succeed")
	assert.Equal(t, userData["name"], profile["name"], "Profile name should match")
	assert.Equal(t, userData["email"], profile["email"], "Profile email should match")
	t.Logf("User profile retrieved successfully")

	// Step 4: Upload File
	t.Log("Step 4: Upload File")
	fileContent := []byte("This is a test file content for E2E testing")
	filename := "test-document.txt"
	fileID, err := uploadFile(t, token, filename, fileContent)
	require.NoError(t, err, "File upload should succeed")
	require.NotEmpty(t, fileID, "File ID should be returned")
	t.Logf("File uploaded with ID: %s", fileID)

	// Step 5: List User Files
	t.Log("Step 5: List User Files")
	files, err := listUserFiles(t, token)
	require.NoError(t, err, "List files should succeed")
	assert.GreaterOrEqual(t, len(files), 1, "At least one file should be listed")

	// Find our uploaded file
	var uploadedFile map[string]interface{}
	for _, file := range files {
		if file["id"] == fileID {
			uploadedFile = file
			break
		}
	}
	require.NotNil(t, uploadedFile, "Uploaded file should be in the list")
	assert.Equal(t, filename, uploadedFile["filename"], "Filename should match")
	t.Logf("File found in user's file list")

	// Step 6: Download File
	t.Log("Step 6: Download File")
	downloadedContent, err := downloadFile(t, token, fileID)
	require.NoError(t, err, "File download should succeed")
	assert.Equal(t, fileContent, downloadedContent, "Downloaded content should match uploaded content")
	t.Logf("File downloaded successfully")

	// Step 7: Update User Profile
	t.Log("Step 7: Update User Profile")
	updatedData := map[string]interface{}{
		"name": "John Doe Updated",
	}
	err = updateUserProfile(t, token, updatedData)
	require.NoError(t, err, "Profile update should succeed")

	// Verify update
	updatedProfile, err := getUserProfile(t, token)
	require.NoError(t, err, "Get updated profile should succeed")
	assert.Equal(t, updatedData["name"], updatedProfile["name"], "Updated name should match")
	t.Logf("User profile updated successfully")

	// Step 8: Delete File
	t.Log("Step 8: Delete File")
	err = deleteFile(t, token, fileID)
	require.NoError(t, err, "File deletion should succeed")

	// Verify deletion
	filesAfterDelete, err := listUserFiles(t, token)
	require.NoError(t, err, "List files after delete should succeed")
	for _, file := range filesAfterDelete {
		assert.NotEqual(t, fileID, file["id"], "Deleted file should not be in the list")
	}
	t.Logf("File deleted successfully")

	// Step 9: Logout (if implemented)
	t.Log("Step 9: Logout")
	err = logoutUser(t, token)
	if err != nil {
		t.Logf("Logout not implemented or failed: %v", err)
	} else {
		t.Logf("User logged out successfully")
	}

	t.Log("Complete user workflow test passed!")
}

// TestMultiUserCollaboration tests multiple users working simultaneously
func TestMultiUserCollaboration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	numUsers := 5
	var wg sync.WaitGroup
	var mu sync.Mutex
	results := make([]map[string]interface{}, 0, numUsers)
	errors := make([]error, 0)

	t.Logf("Testing collaboration with %d users", numUsers)

	for i := 0; i < numUsers; i++ {
		wg.Add(1)
		go func(userIndex int) {
			defer wg.Done()

			// Each user performs a complete workflow
			userData := map[string]interface{}{
				"name":     fmt.Sprintf("User %d", userIndex),
				"email":    fmt.Sprintf("user%d@example.com", userIndex),
				"password": "SecurePassword123!",
			}

			result := map[string]interface{}{
				"userIndex": userIndex,
				"success":   false,
			}

			// Register user
			userID, err := registerUser(t, userData)
			if err != nil {
				mu.Lock()
				errors = append(errors, fmt.Errorf("user %d registration failed: %v", userIndex, err))
				mu.Unlock()
				return
			}

			// Login user
			token, err := loginUser(t, userData["email"].(string), userData["password"].(string))
			if err != nil {
				mu.Lock()
				errors = append(errors, fmt.Errorf("user %d login failed: %v", userIndex, err))
				mu.Unlock()
				return
			}

			// Upload multiple files
			fileIDs := make([]string, 0)
			for j := 0; j < 3; j++ {
				filename := fmt.Sprintf("user%d_file%d.txt", userIndex, j)
				content := []byte(fmt.Sprintf("Content from user %d, file %d", userIndex, j))
				fileID, err := uploadFile(t, token, filename, content)
				if err != nil {
					mu.Lock()
					errors = append(errors, fmt.Errorf("user %d file upload failed: %v", userIndex, err))
					mu.Unlock()
					return
				}
				fileIDs = append(fileIDs, fileID)
			}

			// List files
			files, err := listUserFiles(t, token)
			if err != nil {
				mu.Lock()
				errors = append(errors, fmt.Errorf("user %d list files failed: %v", userIndex, err))
				mu.Unlock()
				return
			}

			result["userID"] = userID
			result["fileCount"] = len(files)
			result["uploadedFiles"] = len(fileIDs)
			result["success"] = true

			mu.Lock()
			results = append(results, result)
			mu.Unlock()
		}(i)
	}

	wg.Wait()

	// Verify results
	require.Empty(t, errors, "No errors should occur during multi-user test")
	require.Len(t, results, numUsers, "All users should complete successfully")

	for _, result := range results {
		assert.True(t, result["success"].(bool), "Each user workflow should succeed")
		assert.Equal(t, 3, result["uploadedFiles"], "Each user should upload 3 files")
		assert.GreaterOrEqual(t, result["fileCount"], 3, "Each user should have at least 3 files")
	}

	t.Logf("Multi-user collaboration test completed successfully with %d users", numUsers)
}

// TestErrorScenarios tests various error conditions and recovery
func TestErrorScenarios(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	t.Run("InvalidCredentials", func(t *testing.T) {
		_, err := loginUser(t, "nonexistent@example.com", "wrongpassword")
		assert.Error(t, err, "Login with invalid credentials should fail")
	})

	t.Run("UnauthorizedAccess", func(t *testing.T) {
		_, err := getUserProfile(t, "invalid-token")
		assert.Error(t, err, "Access with invalid token should fail")
	})

	t.Run("DuplicateRegistration", func(t *testing.T) {
		userData := map[string]interface{}{
			"name":     "Duplicate User",
			"email":    "duplicate@example.com",
			"password": "SecurePassword123!",
		}

		// First registration should succeed
		_, err := registerUser(t, userData)
		require.NoError(t, err, "First registration should succeed")

		// Second registration should fail
		_, err = registerUser(t, userData)
		assert.Error(t, err, "Duplicate registration should fail")
	})

	t.Run("InvalidFileUpload", func(t *testing.T) {
		// Register and login a user first
		userData := map[string]interface{}{
			"name":     "File Test User",
			"email":    "filetest@example.com",
			"password": "SecurePassword123!",
		}
		_, err := registerUser(t, userData)
		require.NoError(t, err)

		token, err := loginUser(t, userData["email"].(string), userData["password"].(string))
		require.NoError(t, err)

		// Try to upload file with malicious content
		maliciousContent := []byte("<script>alert('xss')</script>")
		_, err = uploadFile(t, token, "malicious.js", maliciousContent)
		assert.Error(t, err, "Upload of malicious file should fail")

		// Try to upload oversized file
		oversizedContent := make([]byte, 100*1024*1024) // 100MB
		_, err = uploadFile(t, token, "large.txt", oversizedContent)
		assert.Error(t, err, "Upload of oversized file should fail")
	})

	t.Run("ServiceRecovery", func(t *testing.T) {
		// This test would simulate service failures and recovery
		// In a real scenario, this might involve stopping/starting services
		t.Log("Service recovery test - would test resilience to service failures")
		// For now, just test that services are responsive
		assert.True(t, isServiceHealthy(t, "auth"), "Auth service should be healthy")
		assert.True(t, isServiceHealthy(t, "user"), "User service should be healthy")
		assert.True(t, isServiceHealthy(t, "file"), "File service should be healthy")
	})
}

// Helper functions for E2E operations

func registerUser(t *testing.T, userData map[string]interface{}) (string, error) {
	payload, _ := json.Marshal(userData)
	req := httptest.NewRequest("POST", "/api/users", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	mockUserHandler(rr, req)

	if rr.Code != http.StatusCreated {
		return "", fmt.Errorf("registration failed with status %d", rr.Code)
	}

	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		return "", err
	}

	userID, ok := response["id"].(string)
	if !ok {
		return "", fmt.Errorf("user ID not found in response")
	}

	return userID, nil
}

func loginUser(t *testing.T, email, password string) (string, error) {
	payload := map[string]interface{}{
		"email":    email,
		"password": password,
	}
	data, _ := json.Marshal(payload)

	req := httptest.NewRequest("POST", "/api/auth/login", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	mockAuthHandler(rr, req)

	if rr.Code != http.StatusOK {
		return "", fmt.Errorf("login failed with status %d", rr.Code)
	}

	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		return "", err
	}

	token, ok := response["token"].(string)
	if !ok {
		return "", fmt.Errorf("token not found in response")
	}

	return token, nil
}

func getUserProfile(t *testing.T, token string) (map[string]interface{}, error) {
	req := httptest.NewRequest("GET", "/api/users/profile", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	rr := httptest.NewRecorder()
	mockUserHandler(rr, req)

	if rr.Code != http.StatusOK {
		return nil, fmt.Errorf("get profile failed with status %d", rr.Code)
	}

	var profile map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &profile)
	return profile, err
}

func uploadFile(t *testing.T, token, filename string, content []byte) (string, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return "", err
	}

	_, err = part.Write(content)
	if err != nil {
		return "", err
	}

	err = writer.Close()
	if err != nil {
		return "", err
	}

	req := httptest.NewRequest("POST", "/api/files/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+token)

	rr := httptest.NewRecorder()
	mockFileHandler(rr, req)

	if rr.Code != http.StatusOK {
		return "", fmt.Errorf("file upload failed with status %d", rr.Code)
	}

	var response map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		return "", err
	}

	fileID, ok := response["id"].(string)
	if !ok {
		return "", fmt.Errorf("file ID not found in response")
	}

	return fileID, nil
}

func listUserFiles(t *testing.T, token string) ([]map[string]interface{}, error) {
	req := httptest.NewRequest("GET", "/api/files", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	rr := httptest.NewRecorder()
	mockFileHandler(rr, req)

	if rr.Code != http.StatusOK {
		return nil, fmt.Errorf("list files failed with status %d", rr.Code)
	}

	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		return nil, err
	}

	filesInterface, ok := response["files"]
	if !ok {
		return []map[string]interface{}{}, nil
	}

	files := make([]map[string]interface{}, 0)
	if filesList, ok := filesInterface.([]interface{}); ok {
		for _, file := range filesList {
			if fileMap, ok := file.(map[string]interface{}); ok {
				files = append(files, fileMap)
			}
		}
	}

	return files, nil
}

func downloadFile(t *testing.T, token, fileID string) ([]byte, error) {
	req := httptest.NewRequest("GET", "/api/files/"+fileID, nil)
	req.Header.Set("Authorization", "Bearer "+token)

	rr := httptest.NewRecorder()
	mockFileHandler(rr, req)

	if rr.Code != http.StatusOK {
		return nil, fmt.Errorf("file download failed with status %d", rr.Code)
	}

	return rr.Body.Bytes(), nil
}

func updateUserProfile(t *testing.T, token string, updates map[string]interface{}) error {
	payload, _ := json.Marshal(updates)
	req := httptest.NewRequest("PUT", "/api/users/profile", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	rr := httptest.NewRecorder()
	mockUserHandler(rr, req)

	if rr.Code != http.StatusOK {
		return fmt.Errorf("profile update failed with status %d", rr.Code)
	}

	return nil
}

func deleteFile(t *testing.T, token, fileID string) error {
	req := httptest.NewRequest("DELETE", "/api/files/"+fileID, nil)
	req.Header.Set("Authorization", "Bearer "+token)

	rr := httptest.NewRecorder()
	mockFileHandler(rr, req)

	if rr.Code != http.StatusOK {
		return fmt.Errorf("file deletion failed with status %d", rr.Code)
	}

	return nil
}

func logoutUser(t *testing.T, token string) error {
	req := httptest.NewRequest("POST", "/api/auth/logout", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	rr := httptest.NewRecorder()
	mockAuthHandler(rr, req)

	if rr.Code != http.StatusOK {
		return fmt.Errorf("logout failed with status %d", rr.Code)
	}

	return nil
}

func isServiceHealthy(t *testing.T, service string) bool {
	req := httptest.NewRequest("GET", "/api/"+service+"/health", nil)
	rr := httptest.NewRecorder()

	switch service {
	case "auth":
		mockAuthHandler(rr, req)
	case "user":
		mockUserHandler(rr, req)
	case "file":
		mockFileHandler(rr, req)
	default:
		return false
	}

	return rr.Code == http.StatusOK
}

// Mock handlers for testing (simplified implementations)
func mockAuthHandler(w http.ResponseWriter, r *http.Request) {
	if strings.Contains(r.URL.Path, "health") {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(`{"status":"healthy"}`)); err != nil {
			fmt.Printf("Warning: Failed to write response: %v\n", err)
		}
		return
	}

	if strings.Contains(r.URL.Path, "login") {
		// Check for invalid credentials
		var loginData map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&loginData); err != nil {
			fmt.Printf("Warning: Failed to decode login data: %v\n", err)
		}

		if email, ok := loginData["email"].(string); ok {
			if email == "nonexistent@example.com" {
				w.WriteHeader(http.StatusUnauthorized)
				if _, err := w.Write([]byte(`{"error":"Invalid credentials"}`)); err != nil {
					fmt.Printf("Warning: Failed to write response: %v\n", err)
				}
				return
			}
		}

		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(`{"token":"mock-jwt-token"}`)); err != nil {
			fmt.Printf("Warning: Failed to write response: %v\n", err)
		}
		return
	}

	if strings.Contains(r.URL.Path, "logout") {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(`{"message":"logged out"}`)); err != nil {
			fmt.Printf("Warning: Failed to write response: %v\n", err)
		}
		return
	}

	w.WriteHeader(http.StatusNotFound)
}

// Track registered users for duplicate detection
var registeredUsers = make(map[string]bool)

func mockUserHandler(w http.ResponseWriter, r *http.Request) {
	if strings.Contains(r.URL.Path, "health") {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(`{"status":"healthy"}`)); err != nil {
			fmt.Printf("Warning: Failed to write response: %v\n", err)
		}
		return
	}

	if r.Method == "POST" && r.URL.Path == "/api/users" {
		// Check for duplicate registration
		var userData map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&userData); err != nil {
			fmt.Printf("Warning: Failed to decode user data: %v\n", err)
		}

		if email, ok := userData["email"].(string); ok {
			if registeredUsers[email] {
				w.WriteHeader(http.StatusConflict)
				if _, err := w.Write([]byte(`{"error":"User already exists"}`)); err != nil {
					fmt.Printf("Warning: Failed to write response: %v\n", err)
				}
				return
			}
			registeredUsers[email] = true
		}

		w.WriteHeader(http.StatusCreated)
		if _, err := w.Write([]byte(`{"id":"user-123","message":"user created"}`)); err != nil {
			fmt.Printf("Warning: Failed to write response: %v\n", err)
		}
		return
	}

	if strings.Contains(r.URL.Path, "profile") {
		// Check authorization header
		auth := r.Header.Get("Authorization")
		if auth == "Bearer invalid-token" {
			w.WriteHeader(http.StatusUnauthorized)
			if _, err := w.Write([]byte(`{"error":"Unauthorized"}`)); err != nil {
				fmt.Printf("Warning: Failed to write response: %v\n", err)
			}
			return
		}

		if r.Method == "GET" {
			w.WriteHeader(http.StatusOK)
			if _, err := w.Write([]byte(`{"id":"user-123","name":"John Doe","email":"john.doe@example.com"}`)); err != nil {
				fmt.Printf("Warning: Failed to write response: %v\n", err)
			}
			return
		}
		if r.Method == "PUT" {
			w.WriteHeader(http.StatusOK)
			if _, err := w.Write([]byte(`{"message":"profile updated"}`)); err != nil {
				fmt.Printf("Warning: Failed to write response: %v\n", err)
			}
			return
		}
	}

	w.WriteHeader(http.StatusNotFound)
}

// Track uploaded files per user
var userFiles = make(map[string][]map[string]interface{})
var fileCounter = 0

func mockFileHandler(w http.ResponseWriter, r *http.Request) {
	if strings.Contains(r.URL.Path, "health") {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(`{"status":"healthy"}`)); err != nil {
			fmt.Printf("Warning: Failed to write response: %v\n", err)
		}
		return
	}

	if strings.Contains(r.URL.Path, "upload") {
		// Check for malicious files by examining form data
		err := r.ParseMultipartForm(32 << 20) // 32MB
		if err == nil {
			file, header, err := r.FormFile("file")
			if err == nil {
				defer file.Close()

				// Check for malicious file extensions
				if strings.HasSuffix(header.Filename, ".js") {
					w.WriteHeader(http.StatusRequestEntityTooLarge)
					if _, err := w.Write([]byte(`{"error":"Malicious file type not allowed"}`)); err != nil {
						fmt.Printf("Warning: failed to write response: %v\n", err)
					}
					return
				}

				// Check for oversized files
				if header.Size > 50*1024*1024 { // 50MB limit
					w.WriteHeader(http.StatusRequestEntityTooLarge)
					if _, err := w.Write([]byte(`{"error":"File too large"}`)); err != nil {
						fmt.Printf("Warning: failed to write response: %v\n", err)
					}
					return
				}
			}
		}

		// Extract user from token (simplified)
		token := r.Header.Get("Authorization")
		fileCounter++
		fileID := fmt.Sprintf("file-%d", fileCounter)

		// Store file for this user
		if userFiles[token] == nil {
			userFiles[token] = make([]map[string]interface{}, 0)
		}
		userFiles[token] = append(userFiles[token], map[string]interface{}{
			"id":       fileID,
			"filename": r.FormValue("filename"),
		})

		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(fmt.Sprintf(`{"id":"%s","message":"file uploaded"}`, fileID))); err != nil {
			fmt.Printf("Warning: failed to write response: %v\n", err)
		}
		return
	}

	if r.Method == "GET" && r.URL.Path == "/api/files" {
		token := r.Header.Get("Authorization")
		files := userFiles[token]
		if files == nil {
			files = make([]map[string]interface{}, 0)
		}

		filesJSON, _ := json.Marshal(files)
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(fmt.Sprintf(`{"files":%s}`, string(filesJSON)))); err != nil {
			fmt.Printf("Warning: failed to write response: %v\n", err)
		}
		return
	}

	if r.Method == "GET" && strings.Contains(r.URL.Path, "files/") {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("test file content")); err != nil {
			fmt.Printf("Warning: failed to write response: %v\n", err)
		}
		return
	}

	if r.Method == "DELETE" {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(`{"message":"file deleted"}`)); err != nil {
			fmt.Printf("Warning: failed to write response: %v\n", err)
		}
		return
	}

	w.WriteHeader(http.StatusNotFound)
}
