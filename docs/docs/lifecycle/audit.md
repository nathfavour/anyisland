# Audit & Observability

Anyisland keeps a detailed record of its lifecycle to ensure transparency and reliability.

## Lifecycle Audit Log

All lifecycle events are appended to a JSONL file:
`~/.anyisland/audit/lifecycle.jsonl`

## Event Schema

| Field | Description |
| :--- | :--- |
| `timestamp` | RFC3339 timestamp. |
| `type` | `install`, `update`, `rollback`, `uninstall`, `self_heal`. |
| `action` | Specific action (e.g., `checksum_verify`, `binary_move`). |
| `status` | `success` or `error`. |
| `message` | Human-readable details. |
| `version` | Target version. |
| `commit` | Git SHA of the binary. |

## AI Agent Integration

The `anyisland` CLI and `anyisland daemon` daemon provide these logs to the AI Synthesizer. This allows the AI to:
*   Identify recurring update failures.
*   Suggest rollbacks if a specific version is known to be buggy.
*   Automate recovery steps during self-healing.
