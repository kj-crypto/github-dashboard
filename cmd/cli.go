package main

import (
	"fmt"
	"github-dashboard/pkg/github"
	"github-dashboard/pkg/tui"
	"github-dashboard/pkg/utils"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/spf13/cobra"
)

func setupFileLogger() {
	debug := os.Getenv("GITHUB_DASHBOARD_DEBUG")
	debugEnabled := debug == "on" || debug == "true" || debug == "1"

	if !debugEnabled {
		log.SetOutput(io.Discard)
		return
	}

	// Create logs directory if it doesn't exist
	logsDir := "logs"
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		log.Printf("Failed to create logs directory: %v", err)
		return
	}

	// Create log file with timestamp
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	logFile := filepath.Join(logsDir, fmt.Sprintf("ui_debug_%s.log", timestamp))

	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Printf("Failed to open log file: %v", err)
		return
	}

	log.SetOutput(file)
	log.Printf("UI logging initialized - %s", time.Now().Format(time.RFC3339))
}

func main() {
	token := utils.GetToken()
	if token == "" {
		log.Fatal("GITHUB_TOKEN not set")
	}

	rootCmd := &cobra.Command{
		Use:   "github-cli",
		Short: "GitHub CLI",
		Long:  "GitHub CLI for dashboard browsing and new repo creating",
	}

	dashboardCmd := &cobra.Command{
		Use:   "dashboard [username]",
		Short: "Browse GitHub dashboard for a user",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			username := args[0]
			setupFileLogger()
			m := tui.InitModel(username)
			p := tea.NewProgram(m)
			if _, err := p.Run(); err != nil {
				log.Fatal(err)
			}
		},
	}

	var isPublic bool
	createRepoCmd := &cobra.Command{
		Use:   "create [username] [repoName] [repoDescription]",
		Short: "Create a new GitHub repository",
		Args:  cobra.ExactArgs(3),
		Run: func(cmd *cobra.Command, args []string) {
			username := args[0]
			repoName := args[1]
			repoDesc := args[2]
			fmt.Printf("Creating new repository: %s/%s - %s (public: %v)\n", username, repoName, repoDesc, isPublic)
			resp, err := github.CreateRepository(token, repoName, repoDesc, isPublic)
			if err != nil {
				fmt.Printf("Error creating repository: %v\n", err)
				return
			}
			fmt.Printf("Repository %s created: %s\n", resp.Name, resp.URL)
		},
	}
	createRepoCmd.Flags().BoolVarP(&isPublic, "public", "p", false, "Create a public repository")

	rootCmd.AddCommand(dashboardCmd)
	rootCmd.AddCommand(createRepoCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

}
