# Changelog

All notable changes to `ms-teams-tui` are documented here. This project uses semantic versioning.

## 0.3.0 - 2026-07-10

### Added

- Type-first fuzzy filtering in the file picker, scoped to the current directory and ranked with `sahilm/fuzzy`.
- Drag-and-drop and pasted-path attachment support for regular files, multiple newline-separated files, and mixed image/file selections.
- Space-triggered cached quick preview for PDF, Office, image, and other registered document formats; macOS uses Quick Look.

### Changed

- Printable keys now write into the picker filter by default; arrow, paging, and Enter keys provide secondary list navigation.
- Dotfiles and dot-directories are visible in the file picker by default.
- Local attachment size is checked before reading the file into memory.

### Fixed

- Prevented dropped or pasted file paths from leaking into the outgoing message body as plain text.
- Kept the valid authenticated access token in session memory when token-cache persistence or migration is unavailable, preventing `Send error: no cached token found` during the active session.
- Matched Kitty image transmission, ID bounds, continuation chunks, and placement sequences to cmux's Ghostty renderer so cached popup images render instead of leaving an empty panel.

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
