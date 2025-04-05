package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

func main() {
	// Create a new Docker client
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		fmt.Printf("Error creating Docker client: %v\n", err)
		os.Exit(1)
	}
	defer cli.Close()

	// Get a list of all containers
	containers, err := cli.ContainerList(ctx, types.ContainerListOptions{All: true})
	if err != nil {
		fmt.Printf("Error listing containers: %v\n", err)
		os.Exit(1)
	}

	// If no containers found, display a message
	if len(containers) == 0 {
		fmt.Println("No containers found")
		return
	}

	fmt.Println("=== woki - Simple Container Log Scraper ===")
	fmt.Printf("Found %d containers\n\n", len(containers))

	// Iterate through each container and fetch logs
	for _, container := range containers {
		// Only show logs for running containers
		if container.State != "running" {
			continue
		}

		// Print container info
		fmt.Printf("Container: %s (ID: %.12s)\n", container.Names[0], container.ID)

		// Options for container logs
		options := types.ContainerLogsOptions{
			ShowStdout: true,
			ShowStderr: true,
			Timestamps: true,
			Tail:       "10", // Get only last 10 lines
		}

		// Get container logs
		logs, err := cli.ContainerLogs(ctx, container.ID, options)
		if err != nil {
			fmt.Printf("Error fetching logs for container %s: %v\n", container.ID, err)
			continue
		}

		// Print logs with a 2-second timeout
		logCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		defer cancel()

		// Read and print logs
		fmt.Println("--- Recent Logs ---")
		// Container logs have a header for each line, this is a simple implementation
		// that just copies all data to stdout
		_, err = io.Copy(os.Stdout, logs)
		if err != nil {
			fmt.Printf("Error reading logs: %v\n", err)
		}
		logs.Close()

		fmt.Println("\n--------------------------------")

		// Check if the context has timed out
		select {
		case <-logCtx.Done():
			if logCtx.Err() == context.DeadlineExceeded {
				fmt.Println("Timeout reached when reading logs")
			}
		default:
		}
	}

	fmt.Println("Log scraping completed")
} 
