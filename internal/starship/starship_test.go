package starship

import (
	"fmt"
	"os/exec"
	"strings"
	"testing"
)

func TestInstall(t *testing.T) {
	// Backup and restore original variables
	origExecCommand := execCommand
	origRunCommand := runCommand
	origLookPath := lookPath
	defer func() {
		execCommand = origExecCommand
		runCommand = origRunCommand
		lookPath = origLookPath
	}()

	tests := []struct {
		name        string
		hasStarship bool
		hasBrew     bool
		commandFail bool
		wantErr     bool
		wantCmd     string
	}{
		{
			name:        "Already installed",
			hasStarship: true,
			hasBrew:     false,
			commandFail: false,
			wantErr:     false,
			wantCmd:     "", // Should not run any command
		},
		{
			name:        "Install via Brew",
			hasStarship: false,
			hasBrew:     true,
			commandFail: false,
			wantErr:     false,
			wantCmd:     "brew install starship",
		},
		{
			name:        "Brew install fails",
			hasStarship: false,
			hasBrew:     true,
			commandFail: true,
			wantErr:     true,
			wantCmd:     "brew install starship",
		},
		{
			name:        "Install via Shell",
			hasStarship: false,
			hasBrew:     false,
			commandFail: false,
			wantErr:     false,
			wantCmd:     "curl -sS https://starship.rs/install.sh",
		},
		{
			name:        "Shell install fails",
			hasStarship: false,
			hasBrew:     false,
			commandFail: true,
			wantErr:     true,
			wantCmd:     "curl -sS https://starship.rs/install.sh",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock LookPath
			lookPath = func(file string) (string, error) {
				if file == "starship" && tt.hasStarship {
					return "/usr/bin/starship", nil
				}
				if file == "brew" && tt.hasBrew {
					return "/usr/bin/brew", nil
				}
				if file == "starship" || file == "brew" {
					return "", fmt.Errorf("not found")
				}
				// Allow other lookups to pass if needed (e.g. sh, curl)
				return "/usr/bin/" + file, nil
			}

			// Mock execCommand to capture the command
			var capturedCmd *exec.Cmd
			execCommand = func(name string, arg ...string) *exec.Cmd {
				cmd := exec.Command("echo", "mock") // Dummy command
				// Store the intended command parts for verification
				// Reconstruct command string approx
				fullCmd := name
				if len(arg) > 0 {
					fullCmd += " " + strings.Join(arg, " ")
				}

				// For verification in runCommand if needed, or just captured here
				// We attach the command string to the cmd struct via specific field is hard
				// So we capture it in the outer scope
				capturedCmd = &exec.Cmd{
					Path: name,
					Args: append([]string{name}, arg...),
				}
				return cmd
			}

			// Mock runCommand to simulate success/failure
			runCommand = func(cmd *exec.Cmd) error {
				if tt.commandFail {
					return fmt.Errorf("mock error")
				}
				return nil
			}

			err := Install()

			if (err != nil) != tt.wantErr {
				t.Errorf("Install() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantCmd != "" {
				if capturedCmd == nil {
					t.Errorf("Expected command %q, but no command was executed", tt.wantCmd)
				} else {
					// Reconstruct command line for comparison checks
					// Note: args[0] is the command name in capturedCmd.Args
					cmdStr := strings.Join(capturedCmd.Args, " ")
					// The curl command is complex: sh -c curl ... | sh ...
					// We just check if it contains key parts
					if !strings.Contains(cmdStr, tt.wantCmd) {
						// Special handling for curl command because of pipe matching
						// If we expect "curl ...", we check if the actual command (sh -c ...) contains it
						if strings.HasPrefix(tt.wantCmd, "curl") {
							if !strings.Contains(cmdStr, "curl") {
								t.Errorf("Expected command containing 'curl', got %q", cmdStr)
							}
						} else {
							// For brew, it should be exact match or close
							// capturedCmd.Args will be [brew install starship]
							// cmdStr "brew install starship"
							if cmdStr != tt.wantCmd && !strings.Contains(cmdStr, tt.wantCmd) {
								t.Errorf("Expected command %q, got %q", tt.wantCmd, cmdStr)
							}
						}
					}
				}
			} else {
				if capturedCmd != nil {
					t.Errorf("Expected no command, but %v was executed", capturedCmd.Args)
				}
			}
		})
	}
}

func TestApplyTheme(t *testing.T) {
	// Backup and restore original variables
	origExecCommand := execCommand
	origRunCommand := runCommand
	defer func() {
		execCommand = origExecCommand
		runCommand = origRunCommand
	}()

	tests := []struct {
		name        string
		theme       string
		commandFail bool
		wantErr     bool
	}{
		{"Apply Tokyo Night", "tokyo-night", false, false},
		{"Apply Pastel Powerline", "pastel-powerline", false, false},
		{"Command fails", "broken", true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock execCommand
			var capturedCmd *exec.Cmd
			execCommand = func(name string, arg ...string) *exec.Cmd {
				cmd := exec.Command("echo", "mock")
				capturedCmd = &exec.Cmd{
					Path: name,
					Args: append([]string{name}, arg...),
				}
				return cmd
			}

			// Mock runCommand
			runCommand = func(cmd *exec.Cmd) error {
				if tt.commandFail {
					return fmt.Errorf("mock error")
				}
				return nil
			}

			err := ApplyTheme(tt.theme)

			if (err != nil) != tt.wantErr {
				t.Errorf("ApplyTheme() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				if capturedCmd == nil {
					t.Fatal("Expected command execution, got none")
				}

				// Verify args
				args := capturedCmd.Args
				if args[0] != "starship" || args[1] != "preset" || args[2] != tt.theme {
					t.Errorf("Unexpected command args: %v", args)
				}

				// Check for -o flag and config path
				foundOutputFlag := false
				for i, arg := range args {
					if arg == "-o" {
						foundOutputFlag = true
						if i+1 >= len(args) || !strings.Contains(args[i+1], "starship.toml") {
							t.Error("Expected -o flag followed by starship.toml path")
						}
						break
					}
				}
				if !foundOutputFlag {
					t.Error("Expected -o flag in command")
				}
			}
		})
	}
}
