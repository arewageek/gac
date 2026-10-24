# gac (Git Account Control)

A lightweight CLI tool to switch your Git identity (`user.name`, `user.email`) and authentication (`core.sshCommand`) with a single command.

If you juggle work, open-source, or freelance repositories on the same machine, you've likely hit the common pain points: commits showing up under the wrong email, or Git push failing with `Permission denied (publickey)` because the wrong SSH key was offered to the host.

`gac` fixes this by setting the identity and exact private key directly inside Git's config, either globally or scoped to a specific repository.

---

## Installation

`gac` ships as a standalone binary with zero dependencies. Pick the command for your operating system:

### Linux (x86_64 / amd64)
```bash
curl -fsSL https://github.com/arewageek/gac/releases/latest/download/gac-linux-amd64 -o gac
chmod +x gac
sudo mv gac /usr/local/bin/
```

### Linux (ARM64)
```bash
curl -fsSL https://github.com/arewageek/gac/releases/latest/download/gac-linux-arm64 -o gac
chmod +x gac
sudo mv gac /usr/local/bin/
```

### macOS (Apple Silicon: M1, M2, M3, M4)
```bash
curl -fsSL https://github.com/arewageek/gac/releases/latest/download/gac-darwin-arm64 -o gac
chmod +x gac
sudo mv gac /usr/local/bin/
```

### macOS (Intel)
```bash
curl -fsSL https://github.com/arewageek/gac/releases/latest/download/gac-darwin-amd64 -o gac
chmod +x gac
sudo mv gac /usr/local/bin/
```

### Windows (PowerShell)
```powershell
Invoke-WebRequest -Uri "https://github.com/arewageek/gac/releases/latest/download/gac-windows-amd64.exe" -OutFile "$HOME\bin\gac.exe"
```

---

## Quickstart

### 1. Add your profiles
Run `gac add <name>` to store an identity. It prompts for your commit name, email, and the path to your SSH private key:

```bash
gac add personal
```
```text
Enter Name: Jane Doe
Enter Email: jane@gmail.com
Enter Path to SSH Key: ~/.ssh/id_ed25519_personal
```

Add your other accounts the same way:

```bash
gac add work
```
```text
Enter Name: Jane Doe
Enter Email: jane@company.com
Enter Path to SSH Key: ~/.ssh/id_ed25519_work
```

### 2. Check your profiles
```bash
gac ls
```
```text
* personal (Jane Doe <jane@gmail.com>) [key: ~/.ssh/id_ed25519_personal]
  work (Jane Doe <jane@company.com>) [key: ~/.ssh/id_ed25519_work]
```
The active global profile is marked with an asterisk (`*`).

### 3. Switch profiles

**Globally (Default)**  
Updates your default `~/.gitconfig` machine-wide:
```bash
gac use work
```

**For the current repository only (`--local`)**  
Keeps your global default intact, but binds the current repository's `.git/config` to this profile:
```bash
cd ~/projects/work-repo
gac use work --local
```

### 4. Remove a profile
```bash
gac rm work
```

---

## How It Works

Rather than rewriting `~/.ssh/config` or managing complex host aliases (`git@github-work:org/repo`), `gac` configures Git's native `core.sshCommand`:

```bash
git config core.sshCommand "ssh -i <path-to-key> -F /dev/null"
```

Passing `-F /dev/null` tells SSH to ignore existing host directives in your user config, ensuring only the explicit private key is used for authentication.

---

## Verification & Troubleshooting

Check the active Git config for the current directory:
```bash
git config user.name
git config user.email
git config core.sshCommand
```

Test SSH authentication directly against GitHub:
```bash
ssh -T -i ~/.ssh/id_ed25519_work -F /dev/null git@github.com
```

---

## License

MIT
