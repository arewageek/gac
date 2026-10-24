package git

import (
	"fmt"
	"os/exec"
	"strings"
)

func IsInsideWorkTree() bool {
	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(output)) == "true"
}

func SetConfig(key, value string, isLocal bool) error {
	scopeFlag := "--global"
	if isLocal {
		scopeFlag = "--local"
	}

	cmd := exec.Command("git", "config", scopeFlag, key, value)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git config %s %s failed: %s (%w)", scopeFlag, key, strings.TrimSpace(string(out)), err)
	}
	return nil
}

func ApplyProfile(name, email, expandedSSHKeyPath string, isLocal bool) error {
	if isLocal && !IsInsideWorkTree() {
		return fmt.Errorf("cannot apply --local: current directory is not a Git repository")
	}

	sshCmd := fmt.Sprintf("ssh -i %s -F /dev/null", expandedSSHKeyPath)

	if err := SetConfig("user.name", name, isLocal); err != nil {
		return err
	}
	if err := SetConfig("user.email", email, isLocal); err != nil {
		return err
	}
	if err := SetConfig("core.sshCommand", sshCmd, isLocal); err != nil {
		return err
	}

	return nil
}
