# Visual Export Service

Anyisland provides a centralized **Visual Export Service** that allows any managed TUI or CLI application to generate high-fidelity screenshots and high-quality screen recordings. 

By offloading the rendering and encoding to the Anyisland daemon, applications can provide professional visual exports with **zero dependencies** (no graphics libraries or FFmpeg required) and **minimal CPU impact**.

## Overview

The service operates over the Anyisland Unix Domain Socket (UDS). Managed applications ship raw ANSI strings representing their visual state, and the daemon returns paths to the generated files.

### Key Features
*   **Intelligent Trimming**: Automatically detects border characters (like `│`) and trims empty space to the right, ensuring tight, clean frames.
*   **Modern Aesthetic**: Renders output into a sleek container with Mac-style window controls, rounded corners, and multi-layer soft shadows.
*   **High Fidelity**: Accurate mapping of ANSI colors (16, 256, and TrueColor) using an embedded `GoMono` font for cross-platform consistency.
*   **Hardware Accelerated Recording**: Detects and leverages GPU encoders (`h264_nvenc`, `h264_videotoolbox`) for efficient video generation.

---

## IPC Protocol

Applications communicate with the service by connecting to the socket at `~/.anyisland/anyisland.sock`.

### 1. Screenshots (`VISUAL_SHOT`)

Captures a single terminal state into a PNG image.

**Request:**
```json
{
  "op": "VISUAL_SHOT",
  "tool": "my-app",
  "payload": "\u001b[31mHello\u001b[0m World..."
}
```

**Response:**
```json
{
  "status": "SUCCESS",
  "path": "/home/user/.anyisland/visual/my-app_shot_2026-02-10_120000.png"
}
```

### 2. Screen Recording (`VISUAL_RECORD`)

Captures a stream of terminal states into an MP4 video.

#### Step A: Start Session
**Request:**
```json
{
  "op": "VISUAL_RECORD_START",
  "tool": "my-app"
}
```
**Response:**
```json
{
  "status": "SUCCESS",
  "tool_id": "rec-123456789" 
}
```
*Note: `tool_id` in the response is the Session ID required for subsequent calls.*

#### Step B: Push Frames
Send this every time the application view changes.
**Request:**
```json
{
  "op": "VISUAL_RECORD_FRAME",
  "session": "rec-123456789",
  "payload": "\u001b[31mUpdate\u001b[0m state..."
}
```

#### Step C: Stop & Process
**Request:**
```json
{
  "op": "VISUAL_RECORD_STOP",
  "session": "rec-123456789"
}
```
**Response:**
```json
{
  "status": "SUCCESS",
  "path": "/home/user/.anyisland/visual/anyisland_rec_2026-02-10_120500.mp4"
}
```

---

## Implementation Guide

### Environment Discovery
Managed applications should check for the `ANYISLAND_IPC_SOCK` environment variable to locate the socket.

### Performance Considerations
*   **Frame Deduplication**: The service automatically deduplicates identical frames. Applications should only push frames when the view actually changes (`isDirty`).
*   **Backgrounding**: In Anyisland, rendering and encoding happen in a worker pool. Pushing frames to the socket is extremely fast and will not cause lag in your TUI.

### Storage
All exported assets are stored in the `visual/` directory within the island:
`~/.anyisland/visual/`
