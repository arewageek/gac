# gac (Git Account Control)

A zero-dependency, ultra-fast CLI tool for developers to switch Git identities (`user.name`, `user.email`) and authentication (`core.sshCommand`) with a single command.

No manual editing of `~/.gitconfig` or messing with complex `~/.ssh/config` host aliases.

---

## The Problem

Developers managing multiple GitHub/GitLab accounts (personal, open-source, full-time job, freelance clients) constantly hit two friction points:
1. **Wrong commit author:** Committing to a work repo with your personal email, or polluting an open-source project with your corporate identity.
2. **Permission Denied (`403` / `Permission denied (publickey)`):** SSH agent trying the wrong key for the host, failing pushes.

`gac` solves this by directly injecting the identity and exact SSH key into the Git configuration at either the **global** or **repository** level.

---

## Prerequisites

1. **Git** installed on your system.
2. **Distinct SSH Keys** for your accounts. If you don't have separate keys yet, generate them:
   ```bash
   # Personal key
   ssh-keygen -t ed25519 -C "hello@arewageek.com" -f ~/.ssh/id_ed25519_personal

   # Work key
   ssh-keygen -t ed25519 -C "work@company.com" -f ~/.ssh/id_ed25519_work
   ```
   *(Don't forget to add the corresponding `.pub` keys to your GitHub/GitLab account settings).*

---

## Installation

### Method 1: Direct Download (No Go or Compiler Required)

Download the pre-compiled standalone binary for your operating system directly from the [Latest Release](https://github.com/arewageek/gac/releases/latest):

#### Linux (x86_64 / amd64)
```bash
curl -fsSL https://github.com/arewageek/gac/releases/latest/download/gac-linux-amd64 -o gac
chmod +x gac
sudo mv gac /usr/local/bin/
```

#### Linux (ARM64)
```bash
curl -fsSL https://github.com/arewageek/gac/releases/latest/download/gac-linux-arm64 -o gac
chmod +x gac
sudo mv gac /usr/local/bin/
```

#### macOS (Apple Silicon M1 / M2 / M3 / M4)
```bash
curl -fsSL https://github.com/arewageek/gac/releases/latest/download/gac-darwin-arm64 -o gac
chmod +x gac
sudo mv gac /usr/local/bin/
```

#### macOS (Intel)
```bash
curl -fsSL https://github.com/arewageek/gac/releases/latest/download/gac-darwin-amd64 -o gac
chmod +x gac
sudo mv gac /usr/local/bin/
```

#### Windows (PowerShell)
```powershell
Invoke-WebRequest -Uri "https://github.com/arewageek/gac/releases/latest/download/gac-windows-amd64.exe" -OutFile "$HOME\bin\gac.exe"
# Ensure $HOME\bin is in your PATH
```

---

### Method 2: Using `go install` (If Go is installed)
```bash
go install github.com/arewageek/gac@latest
```
> **Tip:** Ensure your Go bin directory is in your `$PATH`:
> ```bash
> export PATH="$PATH:$(go env GOPATH)/bin"
> ```

---

### Method 3: Build From Source
```bash
git clone https://github.com/arewageek/gac.git
cd gac
go build -o gac .
sudo mv gac /usr/local/bin/
```

---

## Quickstart (60-Second Setup)

### 1. Add your profiles

Add your personal account:
```bash
gac add personal
```
```text
Enter Name: Jane Doe
Enter Email: jane@gmail.com
Enter Path to SSH Key (e.g. ~/.ssh/id_ed25519): ~/.ssh/id_ed25519_personal
```

Add your work account:
```bash
gac add work
```
```text
Enter Name: Jane Doe
Enter Email: jane@company.com
Enter Path to SSH Key (e.g. ~/.ssh/id_ed25519): ~/.ssh/id_ed25519_work
```

### 2. View your profiles
```bash
gac ls
```
Output:
```text
* personal (Jane Doe <jane@gmail.com>) [key: ~/.ssh/id_ed25519_personal]
  work (Jane Doe <jane@company.com>) [key: ~/.ssh/id_ed25519_work]
```
The active global profile is marked with `*`.

---

## Day-to-Day Usage

### Switch Identity Globally
Set your default identity across your entire machine:
```bash
gac use work
```
This updates `~/.gitconfig` and sets `work` as the active global profile in `~/.gac.json`.

### Switch Identity for Current Repository Only (`--local`)
Keep your global identity as personal, but configure a specific corporate repository to use your work identity:
```bash
cd ~/dev/company-project
gac use work --local
```
This scopes the change strictly to `.git/config` inside the current repository. Your global Git settings remain completely untouched.

### Remove a Profile
```bash
gac rm work
```

---

## How It Works

Instead of modifying your OS-level `~/.ssh/config` file, `gac` uses Git's built-in `core.sshCommand`:

```bash
git config core.sshCommand "ssh -i <path-to-key> -F /dev/null"
```

* `-i <path-to-key>`: Tells SSH exactly which private key to use for git operations.
* `-F /dev/null`: Ignores conflicting `~/.ssh/config` host directives, guaranteeing the specified key is always used.

---

## Troubleshooting

### Verify which profile is currently applied in a repo:
```bash
git config user.name
git config user.email
git config core.sshCommand
```

### Test SSH authentication directly:
```bash
ssh -T -i ~/.ssh/id_ed25519_work -F /dev/null git@github.com
# Expected output: Hi <username>! You've successfully authenticated...
```

---

## License

MIT
