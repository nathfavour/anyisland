# Agent Guidelines

## Build Management
- **Never** create build artifacts or binaries directly in the source directories.
- All build outputs must be directed to a `build/` directory at the project root.
- The `build/` directory is explicitly ignored by git to prevent accidental commits of binary data.
- If the `build/` directory does not exist, create it before running build commands.
- Clean up any temporary build files after verification.
