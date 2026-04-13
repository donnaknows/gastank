# Gastank

Gastank is a cross-platform desktop tray app for tracking AI token usage across providers, built with Wails v3 + React.

- Native system tray app (macOS, Windows, Linux)
- GitHub Copilot usage tracking with auto-refresh
- Built-in CLI for terminal use (`gastank usage`)
- OAuth device flow authentication (no env vars needed)

## Install

### macOS / Linux

```bash
curl -fsSL https://raw.githubusercontent.com/donnaknows/gastank/main/scripts/install.sh | bash
```

**macOS**: Installs `gastank.app` to `/Applications`. Since the app is not notarized, you may need to run:
```bash
xattr -cr /Applications/gastank.app
```

**Linux**: Installs an AppImage to `~/.local/bin/gastank-app`.

### Windows

```powershell
iwr -useb https://raw.githubusercontent.com/donnaknows/gastank/main/scripts/install.ps1 | iex
```

Downloads and runs the NSIS installer.

### Manual download

Grab the latest release from [GitHub Releases](https://github.com/donnaknows/gastank/releases).

## Releases

Releases are managed with [`googleapis/release-please-action`](https://github.com/googleapis/release-please-action).

- Conventional commits on `main` update a Release PR
- Merging that Release PR creates the version tag and GitHub release
- The same workflow then builds and uploads macOS, Windows, and Linux artifacts

No manual tagging is required in the normal flow.

## Authentication

Gastank authenticates via the GitHub OAuth device flow — no environment variables or external CLI required.

On first launch, open the app and click **Sign in with GitHub**. You'll be shown a short code and a URL. Open the URL in your browser, enter the code, and approve. The app polls for approval and stores the resulting token at:

- **Linux:** `~/.config/gastank/credentials.json`
- **macOS:** `~/Library/Application Support/gastank/credentials.json`
- **Windows:** `%AppData%\gastank\credentials.json`

Credentials are shared between the GUI and CLI — once logged in via the app, the CLI works without any further setup.

If the token does not have the right access, or the account is not Copilot-enabled, GitHub will typically answer with `401`, `403`, or `404` and the adapter surfaces that response back to the caller.

## CLI Usage

The binary includes a built-in CLI. Log in once via the GUI, then:

```bash
gastank usage                  # fetch Copilot usage (JSON)
gastank usage github-copilot   # explicit provider name
gastank --version              # print version
gastank --help                 # show help
```

## Development

### Prerequisites

| Tool | Version | Install |
|---|---|---|
| Go | 1.25+ | [go.dev/dl](https://go.dev/dl/) |
| Node.js | 20+ | [nodejs.org](https://nodejs.org/) |
| Task | 3.x | [taskfile.dev/installation](https://taskfile.dev/installation/) |
| Wails CLI | v3 alpha | `go install github.com/wailsapp/wails/v3/cmd/wails3@v3.0.0-alpha.74` |

**Platform-specific dependencies:**

<details>
<summary>macOS</summary>

Xcode Command Line Tools (ships with most setups):
```bash
xcode-select --install
```
</details>

<details>
<summary>Linux (Ubuntu/Debian)</summary>

```bash
sudo apt-get install -y build-essential pkg-config libgtk-3-dev libwebkit2gtk-4.1-dev libayatana-appindicator3-dev
```
</details>

<details>
<summary>Windows</summary>

- [MSYS2](https://www.msys2.org/) or [Git for Windows](https://gitforwindows.org/) (provides bash for scripts)
- [NSIS](https://nsis.sourceforge.io/) (only needed for building the installer)
  ```
  choco install nsis
  ```
- WebView2 runtime (usually pre-installed on Windows 10/11)
</details>

### Quick start

```bash
# Install frontend dependencies
cd frontend && npm install && cd ..

# Run in dev mode (hot-reload)
task dev
```

### Run tests

```bash
go test ./internal/...
```

### Build

```bash
task build
```

### Package (macOS .app)

```bash
task package
open bin/gastank.app
```
