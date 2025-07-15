//go:build mage
// +build mage

package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

const (
	// Docker and GCP configuration
	ImageName     = "flight-ticket-service"
	LocalPort     = "8080"
	ContainerPort = "8080"

	// Google Cloud configuration
	ProjectID   = ""                      // GCP project ID
	Region      = "us-east1"              // GCP region
	Repository  = ""                      // Artifact Registry repository name
	ServiceName = "flight-ticket-service" // Cloud Run service name
)

// Default target to run when none is specified
var Default = Build

// Build Go application locally
func Build() error {
	fmt.Println("Building Go application...")
	cmd := exec.Command("go", "build", "-o", "server", "src/cmd/server/server.go")
	return cmd.Run()
}

// Run Go application locally
func Run() error {
	fmt.Println("Running Go application locally on port 6000...")
	cmd := exec.Command("go", "run", "src/cmd/server/server.go")
	cmd.Env = append(os.Environ(), "PORT=6000")
	return cmd.Run()
}

// Clean up build artifacts
func Clean() error {
	fmt.Println("Cleaning up...")
	return os.RemoveAll("server")
}

// Docker build - Build Docker image
func DockerBuild() error {
	fmt.Printf("Building Docker image: %s\n", ImageName)
	cmd := exec.Command("docker", "build", "-t", ImageName, ".")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Docker run - Run Docker container locally
func DockerRun() error {
	fmt.Printf("Running Docker container locally on port %s\n", LocalPort)
	cmd := exec.Command("docker", "run", "--rm", "-p",
		fmt.Sprintf("%s:%s", LocalPort, ContainerPort),
		"-e", fmt.Sprintf("PORT=%s", ContainerPort),
		ImageName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Docker stop - Stop running containers
func DockerStop() error {
	fmt.Println("Stopping Docker containers...")
	cmd := exec.Command("docker", "ps", "-q", "--filter", fmt.Sprintf("ancestor=%s", ImageName))
	output, err := cmd.Output()
	if err != nil {
		return err
	}

	containerIDs := strings.TrimSpace(string(output))
	if containerIDs == "" {
		fmt.Println("No running containers found")
		return nil
	}

	for _, containerID := range strings.Split(containerIDs, "\n") {
		if containerID != "" {
			stopCmd := exec.Command("docker", "stop", containerID)
			if err := stopCmd.Run(); err != nil {
				fmt.Printf("Error stopping container %s: %v\n", containerID, err)
			} else {
				fmt.Printf("Stopped container: %s\n", containerID)
			}
		}
	}
	return nil
}

// Docker push - Tag and push image to Google Artifact Registry
func DockerPush() error {
	artifactRegistryURL := fmt.Sprintf("%s-docker.pkg.dev/%s/%s/%s", Region, ProjectID, Repository, ImageName)

	fmt.Printf("Tagging image for Artifact Registry: %s\n", artifactRegistryURL)
	tagCmd := exec.Command("docker", "tag", ImageName, artifactRegistryURL)
	tagCmd.Stdout = os.Stdout
	tagCmd.Stderr = os.Stderr
	if err := tagCmd.Run(); err != nil {
		return fmt.Errorf("failed to tag image: %v", err)
	}

	fmt.Printf("Pushing image to Artifact Registry: %s\n", artifactRegistryURL)
	pushCmd := exec.Command("docker", "push", artifactRegistryURL)
	pushCmd.Stdout = os.Stdout
	pushCmd.Stderr = os.Stderr
	return pushCmd.Run()
}

// Deploy - Deploy to Google Cloud Run
func Deploy() error {
	artifactRegistryURL := fmt.Sprintf("%s-docker.pkg.dev/%s/%s/%s", Region, ProjectID, Repository, ImageName)

	fmt.Printf("Deploying %s to Cloud Run service: %s\n", artifactRegistryURL, ServiceName)

	args := []string{
		"run", "deploy", ServiceName,
		"--image", artifactRegistryURL,
		"--platform", "managed",
		"--region", Region,
		"--allow-unauthenticated",
		"--port", ContainerPort,
		"--project", ProjectID,
		"--memory", "512Mi",
		"--cpu", "1",
		"--timeout", "300",
		"--concurrency", "100",
		"--max-instances", "10",
		"--set-env-vars", fmt.Sprintf("GOOGLE_CLOUD_PROJECT=%s", ProjectID),
		"--set-env-vars", "GIN_MODE=release",
	}

	cmd := exec.Command("gcloud", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// DeployWithServiceAccount - Deploy to Cloud Run with service account for Firestore access
func DeployWithServiceAccount() error {
	serviceAccountEmail := fmt.Sprintf("%s@%s.iam.gserviceaccount.com", ServiceName, ProjectID)
	artifactRegistryURL := fmt.Sprintf("%s-docker.pkg.dev/%s/%s/%s", Region, ProjectID, Repository, ImageName)

	fmt.Printf("Deploying %s to Cloud Run with service account: %s\n", artifactRegistryURL, serviceAccountEmail)

	args := []string{
		"run", "deploy", ServiceName,
		"--image", artifactRegistryURL,
		"--platform", "managed",
		"--region", Region,
		"--allow-unauthenticated",
		"--port", ContainerPort,
		"--project", ProjectID,
		"--memory", "512Mi",
		"--cpu", "1",
		"--timeout", "300",
		"--concurrency", "100",
		"--max-instances", "10",
		"--service-account", serviceAccountEmail,
		"--set-env-vars", fmt.Sprintf("GOOGLE_CLOUD_PROJECT=%s", ProjectID),
		"--set-env-vars", "GIN_MODE=release",
	}

	cmd := exec.Command("gcloud", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// SetupServiceAccount - Create and configure service account for Firestore access
func SetupServiceAccount() error {
	serviceAccountEmail := fmt.Sprintf("%s@%s.iam.gserviceaccount.com", ServiceName, ProjectID)

	fmt.Printf("Creating service account: %s\n", serviceAccountEmail)

	// Create service account
	createCmd := exec.Command("gcloud", "iam", "service-accounts", "create", ServiceName,
		"--display-name", "Flight Ticket Service",
		"--description", "Service account for Flight Ticket Service Cloud Run deployment",
		"--project", ProjectID)
	createCmd.Stdout = os.Stdout
	createCmd.Stderr = os.Stderr

	if err := createCmd.Run(); err != nil {
		fmt.Printf("Note: Service account creation failed (might already exist): %v\n", err)
	}

	// Grant Firestore permissions
	fmt.Println("Granting Firestore permissions...")
	roles := []string{
		"roles/datastore.user",
		"roles/firebase.admin",
	}

	for _, role := range roles {
		bindCmd := exec.Command("gcloud", "projects", "add-iam-policy-binding", ProjectID,
			"--member", fmt.Sprintf("serviceAccount:%s", serviceAccountEmail),
			"--role", role)
		bindCmd.Stdout = os.Stdout
		bindCmd.Stderr = os.Stderr

		if err := bindCmd.Run(); err != nil {
			fmt.Printf("Warning: Failed to bind role %s: %v\n", role, err)
		}
	}

	fmt.Println("Service account setup completed!")
	return nil
}

// Full pipeline - Build, push, and deploy
func Pipeline() error {
	fmt.Println("Running full pipeline: Build -> Push -> Deploy")

	if err := DockerBuild(); err != nil {
		return fmt.Errorf("docker build failed: %v", err)
	}

	if err := DockerPush(); err != nil {
		return fmt.Errorf("docker push failed: %v", err)
	}

	if err := Deploy(); err != nil {
		return fmt.Errorf("deployment failed: %v", err)
	}

	fmt.Println("Pipeline completed successfully!")
	return nil
}

// FullPipeline - Complete pipeline with service account setup
func FullPipeline() error {
	fmt.Println("Running full pipeline with service account: Setup -> Build -> Push -> Deploy")

	if err := SetupServiceAccount(); err != nil {
		return fmt.Errorf("service account setup failed: %v", err)
	}

	if err := DockerBuild(); err != nil {
		return fmt.Errorf("docker build failed: %v", err)
	}

	if err := DockerPush(); err != nil {
		return fmt.Errorf("docker push failed: %v", err)
	}

	if err := DeployWithServiceAccount(); err != nil {
		return fmt.Errorf("deployment failed: %v", err)
	}

	fmt.Println("Full pipeline completed successfully!")
	return nil
}

// Setup - Create Artifact Registry repository (run once)
func Setup() error {
	fmt.Printf("Creating Artifact Registry repository: %s\n", Repository)

	args := []string{
		"artifacts", "repositories", "create", Repository,
		"--repository-format", "docker",
		"--location", Region,
		"--project", ProjectID,
	}

	cmd := exec.Command("gcloud", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Repository might already exist, so we don't treat this as a fatal error
	if err := cmd.Run(); err != nil {
		fmt.Printf("Note: Repository creation failed (might already exist): %v\n", err)
	}

	// Configure Docker to use gcloud as credential helper
	fmt.Println("Configuring Docker authentication for Artifact Registry...")
	authCmd := exec.Command("gcloud", "auth", "configure-docker",
		fmt.Sprintf("%s-docker.pkg.dev", Region), "--project", ProjectID)
	authCmd.Stdout = os.Stdout
	authCmd.Stderr = os.Stderr
	return authCmd.Run()
}

// Logs - View Cloud Run service logs
func Logs() error {
	fmt.Printf("Fetching logs for Cloud Run service: %s\n", ServiceName)

	cmd := exec.Command("gcloud", "logs", "tail",
		fmt.Sprintf("projects/%s/logs/run.googleapis.com%%2Fstdout", ProjectID),
		"--filter", fmt.Sprintf("resource.labels.service_name=%s", ServiceName),
		"--project", ProjectID)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Status - Get Cloud Run service status and URL
func Status() error {
	fmt.Printf("Getting status for Cloud Run service: %s\n", ServiceName)

	cmd := exec.Command("gcloud", "run", "services", "describe", ServiceName,
		"--region", Region,
		"--project", ProjectID,
		"--format", "value(status.url)")

	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get service status: %v", err)
	}

	serviceURL := strings.TrimSpace(string(output))
	if serviceURL != "" {
		fmt.Printf("🌐 Service URL: %s\n", serviceURL)
		fmt.Printf("📖 API Documentation: %s/swagger/\n", serviceURL)
		fmt.Printf("❤️  Health Check: %s/health\n", serviceURL)
	}

	return nil
}
