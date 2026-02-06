# Contributing to Anyisland

## Adding Official Packages

Official packages are stored in the `packages/official` directory. To add a new official package:

1. Create a new directory under `packages/official/{package-name}`.
2. Add an `anyisland.json` manifest file in that directory.
3. The manifest should follow this structure:

```json
{
  "name": "package-name",
  "version": "version-string",
  "build": {
    "steps": ["command1", "command2"],
    "bin": "relative/path/to/binary"
  }
}
```

## Development

- **CLI:** `cmd/anyisland`
- **Daemon:** `cmd/anyislandd`
- **Internal Logic:** `internal/`
