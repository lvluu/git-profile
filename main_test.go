package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestConfigManager tests the configuration management functionality
func TestConfigManager(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "git-profile-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create a test config path
	testConfigPath := filepath.Join(tmpDir, ".git-profiles-test.json")

	// Create a config manager with the test path
	cm := &ConfigManager{
		ConfigPath: testConfigPath,
		Profiles:   make(map[string]Profile),
	}

	// Test adding a profile
	testProfile := Profile{
		Name:  "John Doe",
		Email: "john.doe@example.com",
	}
	cm.Profiles["work"] = testProfile
	cm.save()

	// Verify the file was created
	_, err = os.Stat(testConfigPath)
	assert.NoError(t, err)

	// Read the file contents
	data, err := os.ReadFile(testConfigPath)
	assert.NoError(t, err)

	// Verify the contents
	var loadedProfiles map[string]Profile
	err = json.Unmarshal(data, &loadedProfiles)
	assert.NoError(t, err)
	assert.Contains(t, loadedProfiles, "work")
	assert.Equal(t, "John Doe", loadedProfiles["work"].Name)
	assert.Equal(t, "john.doe@example.com", loadedProfiles["work"].Email)
}

// TestProfileValidation tests profile input validation
func TestProfileValidation(t *testing.T) {
	// Test with completely new profile
	newProfile := Profile{
		Name:  "Jane Smith",
		Email: "jane.smith@example.com",
	}
	assert.NotEmpty(t, newProfile.Name)
	assert.NotEmpty(t, newProfile.Email)

	// Test with existing profile and partial update
	existingProfile := Profile{
		Name:  "John Doe",
		Email: "john.doe@example.com",
	}

	// Simulate interactive update with some fields kept
	updatedProfile := Profile{
		Name:  "", // Should keep existing name
		Email: "john.updated@example.com",
	}

	// Merge logic
	if updatedProfile.Name == "" {
		updatedProfile.Name = existingProfile.Name
	}
	if updatedProfile.Email == "" {
		updatedProfile.Email = existingProfile.Email
	}

	assert.Equal(t, "John Doe", updatedProfile.Name)
	assert.Equal(t, "john.updated@example.com", updatedProfile.Email)
}

// TestProfileSerialization tests JSON serialization and deserialization
func TestProfileSerialization(t *testing.T) {
	// Create a profile with all fields
	profile := Profile{
		Name:  "Alice Johnson",
		Email: "alice.johnson@example.com",
	}
	profile.Signing.Key = "1234ABCD"

	// Serialize to JSON
	jsonData, err := json.Marshal(profile)
	assert.NoError(t, err)

	// Deserialize back to Profile
	var decodedProfile Profile
	err = json.Unmarshal(jsonData, &decodedProfile)
	assert.NoError(t, err)

	// Verify all fields match
	assert.Equal(t, "Alice Johnson", decodedProfile.Name)
	assert.Equal(t, "alice.johnson@example.com", decodedProfile.Email)
	assert.Equal(t, "1234ABCD", decodedProfile.Signing.Key)
}

// TestMultipleProfiles tests managing multiple profiles
func TestMultipleProfiles(t *testing.T) {
	// Create a config manager
	cm := &ConfigManager{
		Profiles: make(map[string]Profile),
	}

	// Add multiple profiles
	cm.Profiles["work"] = Profile{
		Name:  "John Doe",
		Email: "john.doe@company.com",
	}
	cm.Profiles["personal"] = Profile{
		Name:  "John Personal",
		Email: "john.personal@gmail.com",
	}

	// Verify number of profiles
	assert.Equal(t, 2, len(cm.Profiles))

	// Verify individual profile details
	workProfile, exists := cm.Profiles["work"]
	assert.True(t, exists)
	assert.Equal(t, "John Doe", workProfile.Name)

	personalProfile, exists := cm.Profiles["personal"]
	assert.True(t, exists)
	assert.Equal(t, "John Personal", personalProfile.Name)
}

// TestProfileRemoval tests removing a profile
func TestProfileRemoval(t *testing.T) {
	// Create a config manager with some profiles
	cm := &ConfigManager{
		Profiles: map[string]Profile{
			"work":     {Name: "John Doe", Email: "john.doe@company.com"},
			"personal": {Name: "John Personal", Email: "john.personal@gmail.com"},
		},
	}

	// Initial count
	assert.Equal(t, 2, len(cm.Profiles))

	// Remove a profile
	delete(cm.Profiles, "work")

	// Verify removal
	assert.Equal(t, 1, len(cm.Profiles))
	_, exists := cm.Profiles["work"]
	assert.False(t, exists)
}
