package main

import (
	"log"
	"os"
	google_builds_gitlab "peekmoon.org/google-builds-gitlab"

	"github.com/GoogleCloudPlatform/functions-framework-go/funcframework"
)

func main() {
	funcframework.RegisterHTTPFunction("/", google_builds_gitlab.GitHookHandler)
	funcframework.RegisterHTTPFunction("/config", google_builds_gitlab.GitHookConfigHandler)
	// Use PORT environment variable, or default to 8080.
	port := "8080"
	if envPort := os.Getenv("PORT"); envPort != "" {
		port = envPort
	}

	if err := funcframework.Start(port); err != nil {
		log.Fatalf("funcframework.Start: %v\n", err)
	}
}
