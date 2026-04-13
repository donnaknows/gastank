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

### Run tests

```bash
go test ./internal/...
```

### Live development

```bash
task dev
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
