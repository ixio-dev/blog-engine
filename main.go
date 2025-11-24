package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"tech-blog/build"
	"tech-blog/config"
	"tech-blog/deploy"
	"tech-blog/server"
	"tech-blog/validate"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "blog",
	Short: "Static blog generator",
	Long:  "A minimal, high-performance static blog system built from Markdown files",
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func init() {
	rootCmd.AddCommand(buildCmd)
	rootCmd.AddCommand(previewCmd)
	rootCmd.AddCommand(deployCmd)
	rootCmd.AddCommand(validateCmd)
	rootCmd.AddCommand(cleanCmd)
	rootCmd.AddCommand(newCmd)
}

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build the static site",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load()
		if err != nil {
			log.Fatal("Failed to load config:", err)
		}

		builder, err := build.New(cfg)
		if err != nil {
			log.Fatal("Failed to create builder:", err)
		}

		if err := builder.Build(cfg.Build.Drafts, false); err != nil {
			log.Fatal("Build failed:", err)
		}

		log.Println("Build completed successfully")
	},
}

var previewCmd = &cobra.Command{
	Use:   "preview",
	Short: "Start local preview server",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load()
		if err != nil {
			log.Fatal("Failed to load config:", err)
		}

		server := server.New(cfg)
		if err := server.Start(); err != nil {
			log.Fatal("Server failed:", err)
		}
	},
}

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy site via SCP",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load()
		if err != nil {
			log.Fatal("Failed to load config:", err)
		}

		deployer := deploy.New(cfg)
		if err := deployer.Deploy(); err != nil {
			log.Fatal("Deployment failed:", err)
		}
	},
}

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate configuration and content",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load()
		if err != nil {
			log.Fatal("Failed to load config:", err)
		}

		if err := validate.Config(cfg); err != nil {
			log.Fatal("Config validation failed:", err)
		}

		if err := validate.Content(cfg.ContentDir()); err != nil {
			log.Fatal("Content validation failed:", err)
		}

		log.Println("Validation completed successfully")
	},
}

var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Clean build artifacts",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load()
		if err != nil {
			log.Fatal("Failed to load config:", err)
		}

		if err := os.RemoveAll(cfg.Build.OutputDir); err != nil {
			log.Fatal("Failed to clean build directory:", err)
		}

		log.Println("Clean completed successfully")
	},
}

var newCmd = &cobra.Command{
	Use:   "new [title]",
	Short: "Create a new blog post",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		title := args[0]
		filename := createSlug(title) + ".md"
		filepath := fmt.Sprintf("content/%s", filename)

		// Create the content directory if it doesn't exist
		if err := os.MkdirAll("content", 0755); err != nil {
			log.Fatal("Failed to create content directory:", err)
		}

		// Check if file already exists
		if _, err := os.Stat(filepath); err == nil {
			log.Fatalf("Post already exists: %s", filepath)
		}

		// Create the new post template
		today := time.Now().Format("2006-01-02")
		content := fmt.Sprintf(`---
title: "%s"
date: "%s"
tags: []
draft: true
---

# %s

Write your post content here.

`, title, today, title)

		if err := os.WriteFile(filepath, []byte(content), 0644); err != nil {
			log.Fatal("Failed to create post file:", err)
		}

		log.Printf("Created new post: %s", filepath)
	},
}

func createSlug(title string) string {
	// Simple slugification - lowercase and replace spaces with hyphens
	slug := strings.ToLower(title)
	slug = strings.ReplaceAll(slug, " ", "-")
	// Remove special characters
	slug = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			return r
		}
		return -1
	}, slug)

	// Remove multiple consecutive hyphens
	for strings.Contains(slug, "--") {
		slug = strings.ReplaceAll(slug, "--", "-")
	}

	// Remove leading/trailing hyphens
	return strings.Trim(slug, "-")
}