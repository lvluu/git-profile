package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

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

	data, err := ioutil.ReadFile(cm.ConfigPath)
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

	if err := ioutil.WriteFile(cm.ConfigPath, data, 0644); err != nil {
		log.Fatal(err)
	}
}

func main() {
	configManager := NewConfigManager()

	var rootCmd = &cobra.Command{Use: "git-profile"}

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
		Use:   "add [name] [username] [email]",
		Short: "Add a new Git profile",
		Args:  cobra.ExactArgs(3),
		Run: func(cmd *cobra.Command, args []string) {
			profileName := args[0]
			username := args[1]
			email := args[2]

			profile := Profile{
				Name:  username,
				Email: email,
			}

			configManager.Profiles[profileName] = profile
			configManager.save()

			fmt.Printf("Profile '%s' added successfully!\n", profileName)
		},
	}

	var applyCmd = &cobra.Command{
		Use:   "apply [profile-name]",
		Short: "Apply a specific Git profile",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			profileName := args[0]
			profile, exists := configManager.Profiles[profileName]
			if !exists {
				fmt.Printf("Profile '%s' not found.\n", profileName)
				return
			}

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

			fmt.Printf("Profile '%s' applied successfully!\n", profileName)
		},
	}

	rootCmd.AddCommand(listCmd, addCmd, applyCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
