# Changelog

All notable changes to `ms-teams-tui` are documented here. This project uses semantic versioning.

## 0.2.0 - 2026-07-10

### Added

- Power-user chat and channel layout with stable navigation, message selection, long-message expansion, and responsive rendering.
- Bangla and complex-Unicode-safe wrapping based on terminal cell widths.
- Automatic inline image previews through the Kitty Graphics Protocol, including cmux clipboard image pasting.
- Attachment preview, download, local file upload, and resumable uploads up to 50 MB.
- Structured Adaptive Card rendering for Workflow and bot messages, including facts and open-URL actions.
- cmux-native notifications with message previews and workspace unread state.
- Direct Saved Messages discovery with the `s` shortcut.
- Official Teams audio/video call handoff with uppercase `C` and `V`.
- `VERSION` as the release source of truth, `teams --version`, version bump helpers, release checks, and reproducible multi-platform release automation.
- Checksum-verifying release installer that installs the command as `teams`.
- Security policy, private vulnerability reporting, dependency alerts, and automated security fixes.

### Changed

- Renamed the project and Go module to `ms-teams-tui`; the installed command remains `teams`.
- Config and cache directories now use `ms-teams-tui`. Existing `teams-tui-go` data migrates automatically on first launch.
- SQLite history now uses `ms-teams-tui.db` and migrates the legacy filename.
- Release builds require Go 1.26.5 or newer and embed the exact semantic version.
- GitHub Actions are pinned to full commit SHAs and run race, vet, govulncheck, and gosec gates.

### Security

- Removed an opaque prebuilt Linux executable from the repository.
- Restricted automatic attachment transfers to HTTPS Microsoft Graph, OneDrive, and SharePoint hosts with redirect revalidation.
- Fixed a bearer-token host validation weakness caused by substring matching.
- Prevented attachment filename path traversal outside the Downloads directory.
- Hardened config/cache migration against symlink path escapes.
- Tightened local state and download directory permissions.

### Attribution

- Built from the original [`nospor/teams-tui-go`](https://github.com/nospor/teams-tui-go) project and its contributors.
