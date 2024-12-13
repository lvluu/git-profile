package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

const version = "0.2.0"

// Profile represents a Git profile with name, email, and optional additional config
type Profile struct {
	Name    string `json:"name"`
	Email   string `json:"email"`
	Signing struct {
		Key string `json:"key,omitempty"`
	} `json:"signing,omitempty"`
}

// ConfigManager handles loading and saving profiles
type ConfigManager struct {
	ConfigPath string
	Profiles   map[string]Profile
}

// NewConfigManager creates a new config manager
func NewConfigManager() *ConfigManager {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}

	configPath := filepath.Join(homeDir, ".git-profiles.json")

	cm := &ConfigManager{
		ConfigPath: configPath,
		Profiles:   make(map[string]Profile),
	}

	cm.load()
	return cm
}

// load reads existing profiles from config file
func (cm *ConfigManager) load() {
	if _, err := os.Stat(cm.ConfigPath); os.IsNotExist(err) {
		return
	}

	data, err := os.ReadFile(cm.ConfigPath)
	if err != nil {
		log.Fatal(err)
	}

	if len(data) > 0 {
		if err := json.Unmarshal(data, &cm.Profiles); err != nil {
			log.Fatal(err)
		}
	}
}

// save writes profiles to config file
func (cm *ConfigManager) save() {
	data, err := json.MarshalIndent(cm.Profiles, "", "  ")
	if err != nil {
		log.Fatal(err)
	}

	if err := os.WriteFile(cm.ConfigPath, data, 0644); err != nil {
		log.Fatal(err)
	}
}

// interactiveProfileInput prompts user for profile details
func interactiveProfileInput(existing *Profile) Profile {
	reader := bufio.NewReader(os.Stdin)
	profile := Profile{}

	// Name input
	if existing != nil && existing.Name != "" {
		fmt.Printf("Enter name [current: %s, press Enter to keep]: ", existing.Name)
	} else {
		fmt.Print("Enter name: ")
	}
	name, _ := reader.ReadString('\n')
	name = strings.TrimSpace(name)
	if name == "" && existing != nil {
		profile.Name = existing.Name
	} else {
		profile.Name = name
	}

	// Email input
	if existing != nil && existing.Email != "" {
		fmt.Printf("Enter email [current: %s, press Enter to keep]: ", existing.Email)
	} else {
		fmt.Print("Enter email: ")
	}
	email, _ := reader.ReadString('\n')
	email = strings.TrimSpace(email)
	if email == "" && existing != nil {
		profile.Email = existing.Email
	} else {
		profile.Email = email
	}

	// Optional signing key
	fmt.Print("Enter signing key (optional, press Enter to skip): ")
	signingKey, _ := reader.ReadString('\n')
	signingKey = strings.TrimSpace(signingKey)
	if signingKey != "" {
		profile.Signing.Key = signingKey
	} else if existing != nil {
		profile.Signing.Key = existing.Signing.Key
	}

	return profile
}

func main() {
	configManager := NewConfigManager()

	var rootCmd = &cobra.Command{
		Use:     "git-profile",
		Version: version,
	}

	// Version flag
	rootCmd.PersistentFlags().BoolP("version", "v", false, "Show version information")

	var listCmd = &cobra.Command{
		Use:   "ls",
		Short: "List all saved Git profiles",
		Run: func(cmd *cobra.Command, args []string) {
			if len(configManager.Profiles) == 0 {
				fmt.Println("No profiles found. Use 'git profile add' to create a profile.")
				return
			}

			for name, profile := range configManager.Profiles {
				fmt.Printf("Profile: %s\n", name)
				fmt.Printf("  Name:  %s\n", profile.Name)
				fmt.Printf("  Email: %s\n", profile.Email)
				if profile.Signing.Key != "" {
					fmt.Printf("  Signing Key: %s\n", profile.Signing.Key)
				}
				fmt.Println()
			}
		},
	}

	var addCmd = &cobra.Command{
		Use:   "add",
		Short: "Add a new Git profile (interactive)",
		Run: func(cmd *cobra.Command, args []string) {
			// Interactive profile name selection
			prompt := promptui.Prompt{
				Label: "Enter profile name",
				Validate: func(input string) error {
					if input == "" {
						return fmt.Errorf("profile name cannot be empty")
					}
					if _, exists := configManager.Profiles[input]; exists {
						return fmt.Errorf("profile '%s' already exists", input)
					}
					return nil
				},
			}

			profileName, err := prompt.Run()
			if err != nil {
				fmt.Println("Cancelled.")
				return
			}

			// Interactive profile details input
			profile := interactiveProfileInput(nil)

			// Save the profile
			configManager.Profiles[profileName] = profile
			configManager.save()

			fmt.Printf("Profile '%s' added successfully!\n", profileName)
		},
	}

	var editCmd = &cobra.Command{
		Use:   "edit",
		Short: "Edit an existing Git profile (interactive)",
		Run: func(cmd *cobra.Command, args []string) {
			// Select profile to edit
			var profileNames []string
			for name := range configManager.Profiles {
				profileNames = append(profileNames, name)
			}

			prompt := promptui.Select{
				Label: "Select profile to edit",
				Items: profileNames,
			}

			_, selectedProfile, err := prompt.Run()
			if err != nil {
				fmt.Println("Cancelled.")
				return
			}

			// Existing profile
			existingProfile := configManager.Profiles[selectedProfile]

			// Interactive edit
			updatedProfile := interactiveProfileInput(&existingProfile)

			// Save updated profile
			configManager.Profiles[selectedProfile] = updatedProfile
			configManager.save()

			fmt.Printf("Profile '%s' updated successfully!\n", selectedProfile)
		},
	}

	var removeCmd = &cobra.Command{
		Use:   "rm",
		Short: "Remove a Git profile (interactive)",
		Run: func(cmd *cobra.Command, args []string) {
			// Select profile to remove
			var profileNames []string
			for name := range configManager.Profiles {
				profileNames = append(profileNames, name)
			}

			prompt := promptui.Select{
				Label: "Select profile to remove",
				Items: profileNames,
			}

			_, selectedProfile, err := prompt.Run()
			if err != nil {
				fmt.Println("Cancelled.")
				return
			}

			// Confirmation prompt
			confirmPrompt := promptui.Prompt{
				Label:     fmt.Sprintf("Are you sure you want to remove profile '%s'", selectedProfile),
				IsConfirm: true,
			}

			_, confirmErr := confirmPrompt.Run()
			if confirmErr != nil {
				fmt.Println("Removal cancelled.")
				return
			}

			// Remove profile
			delete(configManager.Profiles, selectedProfile)
			configManager.save()

			fmt.Printf("Profile '%s' removed successfully!\n", selectedProfile)
		},
	}

	var applyCmd = &cobra.Command{
		Use:   "apply",
		Short: "Apply a specific Git profile (interactive)",
		Run: func(cmd *cobra.Command, args []string) {
			// Select profile to apply
			var profileNames []string
			for name := range configManager.Profiles {
				profileNames = append(profileNames, name)
			}

			prompt := promptui.Select{
				Label: "Select profile to apply",
				Items: profileNames,
			}

			_, selectedProfile, err := prompt.Run()
			if err != nil {
				fmt.Println("Cancelled.")
				return
			}

			profile := configManager.Profiles[selectedProfile]

			gitCommands := [][]string{
				{"config", "--global", "user.name", profile.Name},
				{"config", "--global", "user.email", profile.Email},
			}

			for _, gitCmd := range gitCommands {
				cmd := exec.Command("git", gitCmd...)
				if err := cmd.Run(); err != nil {
					fmt.Printf("Error applying profile: %v\n", err)
					return
				}
			}

			fmt.Printf("Profile '%s' applied successfully!\n", selectedProfile)
		},
	}

	rootCmd.AddCommand(listCmd, addCmd, editCmd, removeCmd, applyCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
