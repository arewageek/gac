package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"gac/internal/config"
	"gac/internal/git"
)

const (
	colorReset = "\033[0m"
	colorGreen = "\033[32m"
	colorBold  = "\033[1m"
)

func main() {
	if err := run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string, stdin io.Reader, stdout, stderr io.Writer) error {
	if len(args) == 0 {
		printUsage(stdout)
		return nil
	}

	subcommand := args[0]
	subArgs := args[1:]

	switch subcommand {
	case "add":
		return runAdd(subArgs, stdin, stdout)
	case "use":
		return runUse(subArgs, stdout)
	case "ls", "list":
		return runList(subArgs, stdout)
	case "rm", "remove":
		return runRemove(subArgs, stdout)
	case "help", "-h", "--help":
		printUsage(stdout)
		return nil
	default:
		return fmt.Errorf("unknown command %q. Run 'gac help' for usage", subcommand)
	}
}

func printUsage(w io.Writer) {
	usage := `gac - Git Account Control

Usage:
  gac add <profile>          Prompt for identity details and save a new profile
  gac use <profile> [--local] Switch active Git identity (globally by default)
  gac ls                     List all saved profiles and show the active profile
  gac rm <profile>           Remove a profile from ~/.gac.json
  gac help                   Show this help message

Options:
  --local                    Scope identity and SSH key changes to current repository only
`
	fmt.Fprint(w, usage)
}

func promptInput(scanner *bufio.Scanner, w io.Writer, prompt string) (string, error) {
	fmt.Fprint(w, prompt)
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return "", err
		}
		return "", fmt.Errorf("input aborted")
	}
	return strings.TrimSpace(scanner.Text()), nil
}

func runAdd(args []string, r io.Reader, w io.Writer) error {
	if len(args) < 1 {
		return fmt.Errorf("missing profile name. Usage: gac add <profile>")
	}
	profileKey := strings.TrimSpace(args[0])
	if profileKey == "" {
		return fmt.Errorf("profile name cannot be empty")
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	if _, exists := cfg.Profiles[profileKey]; exists {
		fmt.Fprintf(w, "Note: Profile %q already exists and will be updated.\n", profileKey)
	}

	scanner := bufio.NewScanner(r)

	name, err := promptInput(scanner, w, "Enter Name: ")
	if err != nil {
		return err
	}
	if name == "" {
		return fmt.Errorf("name cannot be empty")
	}

	email, err := promptInput(scanner, w, "Enter Email: ")
	if err != nil {
		return err
	}
	if email == "" {
		return fmt.Errorf("email cannot be empty")
	}

	sshKey, err := promptInput(scanner, w, "Enter Path to SSH Key (e.g. ~/.ssh/id_ed25519): ")
	if err != nil {
		return err
	}
	if sshKey == "" {
		return fmt.Errorf("ssh key path cannot be empty")
	}

	expandedKey, err := config.ExpandPath(sshKey)
	if err != nil {
		return fmt.Errorf("invalid SSH key path: %w", err)
	}

	if info, err := os.Stat(expandedKey); err != nil {
		if os.IsNotExist(err) {
			fmt.Fprintf(w, "warning: SSH key file %q does not exist yet at %s\n", sshKey, expandedKey)
		} else {
			return fmt.Errorf("cannot verify SSH key path %q: %w", sshKey, err)
		}
	} else if info.IsDir() {
		return fmt.Errorf("specified SSH key path %q is a directory, not a key file", sshKey)
	}

	cfg.Profiles[profileKey] = config.Profile{
		Name:   name,
		Email:  email,
		SSHKey: sshKey,
	}

	if cfg.Active == "" {
		cfg.Active = profileKey
	}

	if err := cfg.Save(); err != nil {
		return err
	}

	fmt.Fprintf(w, "Saved profile %q.\n", profileKey)
	return nil
}

func runUse(args []string, w io.Writer) error {
	var profileKey string
	isLocal := false

	for _, arg := range args {
		if arg == "--local" {
			isLocal = true
		} else if strings.HasPrefix(arg, "-") {
			return fmt.Errorf("unknown flag %q. Usage: gac use <profile> [--local]", arg)
		} else if profileKey == "" {
			profileKey = arg
		} else {
			return fmt.Errorf("unexpected argument %q. Usage: gac use <profile> [--local]", arg)
		}
	}

	if profileKey == "" {
		return fmt.Errorf("missing profile name. Usage: gac use <profile> [--local]")
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	profile, ok := cfg.Profiles[profileKey]
	if !ok {
		return fmt.Errorf("profile %q not found. Run 'gac ls' to see available profiles", profileKey)
	}

	expandedKey, err := config.ExpandPath(profile.SSHKey)
	if err != nil {
		return fmt.Errorf("failed to expand SSH key path %q: %w", profile.SSHKey, err)
	}

	if err := git.ApplyProfile(profile.Name, profile.Email, expandedKey, isLocal); err != nil {
		return err
	}

	if isLocal {
		fmt.Fprintf(w, "Switched local repository git identity to %q (%s <%s>).\n", profileKey, profile.Name, profile.Email)
	} else {
		cfg.Active = profileKey
		if err := cfg.Save(); err != nil {
			return fmt.Errorf("profile applied to git config, but failed to update ~/.gac.json active state: %w", err)
		}
		fmt.Fprintf(w, "Switched global git identity to %q (%s <%s>).\n", profileKey, profile.Name, profile.Email)
	}

	return nil
}

func runList(args []string, w io.Writer) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	if len(cfg.Profiles) == 0 {
		fmt.Fprintln(w, "No profiles configured. Run 'gac add <profile>' to create one.")
		return nil
	}

	keys := make([]string, 0, len(cfg.Profiles))
	for k := range cfg.Profiles {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		p := cfg.Profiles[k]
		if k == cfg.Active {
			fmt.Fprintf(w, "%s* %s%s (%s <%s>) [key: %s]\n", colorGreen, colorBold, k, p.Name, p.Email, p.SSHKey)
		} else {
			fmt.Fprintf(w, "  %s (%s <%s>) [key: %s]\n", k, p.Name, p.Email, p.SSHKey)
		}
	}

	return nil
}

func runRemove(args []string, w io.Writer) error {
	if len(args) < 1 {
		return fmt.Errorf("missing profile name. Usage: gac rm <profile>")
	}
	profileKey := strings.TrimSpace(args[0])

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	if _, ok := cfg.Profiles[profileKey]; !ok {
		return fmt.Errorf("profile %q not found", profileKey)
	}

	delete(cfg.Profiles, profileKey)
	if cfg.Active == profileKey {
		cfg.Active = ""
	}

	if err := cfg.Save(); err != nil {
		return err
	}

	fmt.Fprintf(w, "Removed profile %q.\n", profileKey)
	return nil
}
