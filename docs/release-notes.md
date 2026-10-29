- Added native darwin/arm64 (Apple Silicon) RMK binaries to the release build, alongside linux/amd64 and darwin/amd64.
- Updated the s3-installer script to detect the CPU architecture and download the matching macOS binary (including when the shell runs under Rosetta).
- Added the .Arch placeholder for the inventory.tools section of the project file, with the optional arch-amd64 and arch-arm64 fields to override the architecture name in tool download URLs (by default amd64 or arm64).
- Updated the project file examples to use .Arch instead of the hardcoded amd64.
- Changed rmk update to install the binary for the hardware architecture, so RMK run under Rosetta gets the native arm64 version.
- Changed the tools cache to download tools again when the CPU architecture changes.
- Updated Go to 1.27.3 and switched the CI to the Go version from go.mod.
- Updated go-git, kiota-http-go and xpath to fix vulnerabilities (GO-2026-6214, GO-2026-6213, GO-2026-5224, GO-2026-4526).
- Fixed download error handling and error formatting in rmk secret.

Additional information

Starting with v0.49.0, rmk update installs the binary for the hardware architecture. RMK versions earlier than v0.49.0 update to the same architecture, so to switch from amd64 to the native arm64 version on Apple Silicon, run rmk update twice (the first run installs v0.49.0, the second one installs arm64) or reinstall RMK with the installer.
