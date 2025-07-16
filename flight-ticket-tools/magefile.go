//go:build mage

package main

import (
	"fmt"
	"time"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

const (
	// Docker configuration
	registryURL = ""
	imageName   = "flight-ticket-tools"
	dockerFile  = "Dockerfile"

	// Google Cloud configuration
	projectID      = "a3rlabs-sandbox"
	region         = "us-east1"
	serviceName    = "flight-ticket-tools"
	serviceAccount = "<SERVICE_ACCOUNT_EMAIL>"
)

// Docker namespace contains Docker-related build targets
type Docker mg.Namespace

// Build builds the Docker image
func (Docker) Build() error {
	fmt.Println("Building Docker image...")

	tag := getImageTag()
	fullImageName := fmt.Sprintf("%s/%s:%s", registryURL, imageName, tag)

	return sh.Run("docker", "build",
		"-t", fullImageName,
		"-t", fmt.Sprintf("%s/%s:latest", registryURL, imageName),
		"-f", dockerFile,
		".")
}

// Push pushes the Docker image to Artifact Registry
func (Docker) Push() error {
	mg.Deps(Docker.Build)

	fmt.Println("Configuring Docker for Artifact Registry...")
	if err := sh.Run("gcloud", "auth", "configure-docker", fmt.Sprintf("%s-docker.pkg.dev", region)); err != nil {
		return fmt.Errorf("failed to configure docker auth: %w", err)
	}

	fmt.Println("Pushing Docker image to Artifact Registry...")

	tag := getImageTag()
	fullImageName := fmt.Sprintf("%s/%s:%s", registryURL, imageName, tag)
	latestImageName := fmt.Sprintf("%s/%s:latest", registryURL, imageName)

	// Push both tagged and latest versions
	if err := sh.Run("docker", "push", fullImageName); err != nil {
		return fmt.Errorf("failed to push tagged image: %w", err)
	}

	if err := sh.Run("docker", "push", latestImageName); err != nil {
		return fmt.Errorf("failed to push latest image: %w", err)
	}

	fmt.Printf("Successfully pushed image: %s\n", fullImageName)
	return nil
}

// Run runs the Docker image locally for testing
func (Docker) Run() error {
	mg.Deps(Docker.Build)

	fmt.Println("Running Docker image locally...")

	tag := getImageTag()
	fullImageName := fmt.Sprintf("%s/%s:%s", registryURL, imageName, tag)

	return sh.Run("docker", "run",
		"-p", "8080:8080",
		"-e", "ENVIRONMENT=cloudrun",
		"-e", "PORT=8080",
		"--rm",
		fullImageName)
}

// CloudRun namespace contains Cloud Run deployment targets
type CloudRun mg.Namespace

// Deploy deploys the service to Cloud Run with service account
func (CloudRun) Deploy() error {
	mg.Deps(Docker.Push)

	fmt.Println("Deploying to Cloud Run...")

	tag := getImageTag()
	fullImageName := fmt.Sprintf("%s/%s:%s", registryURL, imageName, tag)

	return sh.Run("gcloud", "run", "deploy", serviceName,
		"--image", fullImageName,
		"--platform", "managed",
		"--region", region,
		"--project", projectID,
		"--service-account", serviceAccount,
		"--allow-unauthenticated",
		"--port", "8080",
		"--memory", "512Mi",
		"--cpu", "1",
		"--concurrency", "100",
		"--max-instances", "10",
		"--timeout", "300",
		"--set-env-vars", "ENVIRONMENT=cloudrun,PORT=8080")
}

// Update updates an existing Cloud Run service
func (CloudRun) Update() error {
	mg.Deps(Docker.Push)

	fmt.Println("Updating Cloud Run service...")

	tag := getImageTag()
	fullImageName := fmt.Sprintf("%s/%s:%s", registryURL, imageName, tag)

	return sh.Run("gcloud", "run", "services", "update", serviceName,
		"--image", fullImageName,
		"--region", region,
		"--project", projectID)
}

// Logs shows Cloud Run service logs
func (CloudRun) Logs() error {
	fmt.Println("Fetching Cloud Run logs...")
	return sh.Run("gcloud", "logs", "tail",
		fmt.Sprintf("projects/%s/logs/run.googleapis.com%%2Fstdout", projectID),
		"--filter", fmt.Sprintf(`resource.labels.service_name="%s"`, serviceName))
}

// Status shows Cloud Run service status
func (CloudRun) Status() error {
	fmt.Println("Getting Cloud Run service status...")
	return sh.Run("gcloud", "run", "services", "describe", serviceName,
		"--region", region,
		"--project", projectID)
}

// Delete deletes the Cloud Run service
func (CloudRun) Delete() error {
	fmt.Println("Deleting Cloud Run service...")
	return sh.Run("gcloud", "run", "services", "delete", serviceName,
		"--region", region,
		"--project", projectID,
		"--quiet")
}

// Dev namespace contains development targets
type Dev mg.Namespace

// Local runs the application locally in stdio mode
func (Dev) Local() error {
	fmt.Println("Running application locally in stdio mode...")
	return sh.Run("uv", "run", "python", "main.py")
}

// Http runs the application locally in HTTP mode for testing Cloud Run behavior
func (Dev) Http() error {
	fmt.Println("Running application locally in HTTP mode...")
	env := map[string]string{
		"ENVIRONMENT": "cloudrun",
		"PORT":        "8080",
	}
	return sh.RunWith(env, "uv", "run", "python", "main.py")
}

// Test runs local tests (placeholder for future tests)
func (Dev) Test() error {
	fmt.Println("Running tests...")
	// Add test commands here when tests are implemented
	fmt.Println("No tests implemented yet")
	return nil
}

// Clean removes local Docker images
func Clean() error {
	fmt.Println("Cleaning up local Docker images...")

	// Remove local images (ignore errors if images don't exist)
	tag := getImageTag()
	fullImageName := fmt.Sprintf("%s/%s:%s", registryURL, imageName, tag)
	latestImageName := fmt.Sprintf("%s/%s:latest", registryURL, imageName)

	sh.Run("docker", "rmi", fullImageName)
	sh.Run("docker", "rmi", latestImageName)

	// Clean up dangling images
	return sh.Run("docker", "image", "prune", "-f")
}

// Setup sets up the development environment
func Setup() error {
	fmt.Println("Setting up development environment...")

	// Install uv if not present
	if err := sh.Run("which", "uv"); err != nil {
		fmt.Println("Installing uv...")
		if err := sh.Run("curl", "-LsSf", "https://astral.sh/uv/install.sh", "|", "sh"); err != nil {
			return fmt.Errorf("failed to install uv: %w", err)
		}
	}

	// Sync dependencies
	fmt.Println("Syncing dependencies...")
	if err := sh.Run("uv", "sync"); err != nil {
		return fmt.Errorf("failed to sync dependencies: %w", err)
	}

	// Configure gcloud if needed
	fmt.Println("Checking gcloud configuration...")
	if err := sh.Run("gcloud", "config", "get-value", "project"); err != nil {
		fmt.Printf("Please run 'gcloud auth login' and 'gcloud config set project %s'\n", projectID)
	}

	return nil
}

// getImageTag generates a tag based on current timestamp and git commit (if available)
func getImageTag() string {
	// Try to get git commit hash
	if gitHash, err := sh.Output("git", "rev-parse", "--short", "HEAD"); err == nil {
		return fmt.Sprintf("%s-%s", time.Now().Format("20060102-150405"), gitHash)
	}

	// Fallback to timestamp only
	return time.Now().Format("20060102-150405")
}
