package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/waiiphyo/prometheus-target-generator/internal/generator"
)

// Function to read an existing JSON file and append new targets if necessary
func appendNewTargets(outputFile string, newTargets []generator.TargetConfig) error {
	// Check if the file already exists
	if _, err := os.Stat(outputFile); err == nil {
		// Read existing data
		existingData, err := os.ReadFile(outputFile)
		if err != nil {
			return fmt.Errorf("failed to read existing file: %w", err)
		}

		// Parse the existing data
		var existingTargets []generator.TargetConfig
		if err := json.Unmarshal(existingData, &existingTargets); err != nil {
			return fmt.Errorf("failed to unmarshal existing JSON: %w", err)
		}

		// Create a map to track existing targets
		existingTargetsMap := make(map[string]bool)
		for _, target := range existingTargets {
			for _, targetIP := range target.Targets {
				existingTargetsMap[targetIP] = true
			}
		}

		// Append only new targets
		for _, newTarget := range newTargets {
			for _, targetIP := range newTarget.Targets {
				if !existingTargetsMap[targetIP] {
					existingTargets = append(existingTargets, newTarget)
					existingTargetsMap[targetIP] = true
				}
			}
		}

		// Write updated JSON
		updatedData, err := json.MarshalIndent(existingTargets, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal updated JSON: %w", err)
		}

		if err := os.WriteFile(outputFile, updatedData, 0644); err != nil {
			return fmt.Errorf("failed to write updated file: %w", err)
		}

		log.Printf("Updated: %s\n", outputFile)
	} else {
		// Create new JSON file with new targets
		updatedData, err := json.MarshalIndent(newTargets, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal new JSON: %w", err)
		}

		if err := os.WriteFile(outputFile, updatedData, 0644); err != nil {
			return fmt.Errorf("failed to write new file: %w", err)
		}

		log.Printf("Created: %s\n", outputFile)
	}

	return nil
}

// Main function to handle input, generate targets, and write to the output file
func main() {
	inputFile := flag.String("input", "input.txt", "Path to the input file")
	outputDir := flag.String("output-dir", "/etc/prometheus/targets", "Directory to save JSON files")
	flag.Parse()

	// Open the input file
	file, err := os.Open(*inputFile)
	if err != nil {
		log.Fatalf("Failed to open input file: %v", err)
	}
	defer file.Close()

	var groupName string
	var ipData string

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		// If a new group starts, generate the targets and append to file
		if line[0] == '[' && line[len(line)-1] == ']' {
			if groupName != "" && ipData != "" {
				newTargets, err := generator.GenerateTargets(groupName, ipData)
				if err != nil {
					log.Fatalf("Failed to generate targets: %v", err)
				}

				// Append the new targets to the JSON file
				outputFile := filepath.Join(*outputDir, fmt.Sprintf("%s.json", groupName))
				if err := appendNewTargets(outputFile, newTargets); err != nil {
					log.Fatalf("Failed to append targets: %v", err)
				}
			}

			// Reset for the new group
			groupName = line[1 : len(line)-1]
			ipData = ""
		} else {
			// Add IPs to ipData for the current group
			ipData += line + ","
		}
	}

	// Handle last group
	if groupName != "" && ipData != "" {
		newTargets, err := generator.GenerateTargets(groupName, ipData)
		if err != nil {
			log.Fatalf("Failed to generate targets: %v", err)
		}

		outputFile := filepath.Join(*outputDir, fmt.Sprintf("%s.json", groupName))
		if err := appendNewTargets(outputFile, newTargets); err != nil {
			log.Fatalf("Failed to append targets: %v", err)
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("Error reading input file: %v", err)
	}

	log.Println("Prometheus target JSON files updated successfully!")
}
