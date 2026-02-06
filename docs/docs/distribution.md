---
sidebar_position: 3
---

# Distribution for Developers

Third-party developers can set up their tools for distribution via Anyisland by providing an `anyisland.json` manifest.

## Creating a Manifest

The `anyisland.json` file should be placed in the root of your repository. It tells Anyisland how to build and install your tool.

### Manifest Structure

```json
{
  "name": "your-tool-name",
  "version": "1.0.0",
  "build": {
    "steps": [
      "go build -o your-tool ."
    ],
    "bin": "your-tool"
  }
}
```

- `name`: The name of your tool.
- `version`: The current version of your tool.
- `build`:
    - `steps`: A list of shell commands required to build your tool from source.
    - `bin`: The relative path to the resulting binary after the build steps are executed.

## Official Packages

If you want your tool to be part of the official Anyisland package set, you can contribute a manifest to the `packages/official` directory in the Anyisland repository.

Each official package has its own directory:
`packages/official/{package-name}/anyisland.json`

## Anyisland-Aware Tools

Tools can become "Anyisland-aware" by registering themselves with the Anyisland daemon via UDP heartbeats. This allows for automated discovery and management.

Example registration heartbeat (UDP port 1995):

```json
{
  "op": "REGISTER",
  "name": "aware-tool",
  "source": "github.com/user/repo",
  "version": "v1.0.0",
  "type": "binary"
}
```
