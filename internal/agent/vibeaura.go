package agent

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type VibeauraSynthesizer struct {
	socketPath string
}

func NewVibeauraSynthesizer() *VibeauraSynthesizer {
	home, _ := os.UserHomeDir()
	return &VibeauraSynthesizer{
		socketPath: filepath.Join(home, ".vibeauracle", "vibeaura.sock"),
	}
}

type ipcRequest struct {
	Type    string      `json:"type"`
	Method  string      `json:"method"`
	ID      string      `json:"id"`
	Payload interface{} `json:"payload"`
}

type ipcResponse struct {
	Type    string          `json:"type"`
	ID      string          `json:"id"`
	Payload json.RawMessage `json:"payload"`
}

func (v *VibeauraSynthesizer) query(ctx context.Context, content string, intent string) (string, error) {
	conn, err := net.Dial("unix", v.socketPath)
	if err != nil {
		return "", fmt.Errorf("failed to connect to vibeaura socket at %s: %w. Is vibeaura running?", v.socketPath, err)
	}
	defer conn.Close()

	id := fmt.Sprintf("anyisland-%d", time.Now().UnixNano())
	req := ipcRequest{
		Type:   "request",
		Method: "query",
		ID:     id,
		Payload: map[string]string{
			"content": content,
			"intent":  intent,
		},
	}

	data, err := json.Marshal(req)
	if err != nil {
		return "", err
	}

	if _, err := fmt.Fprintln(conn, string(data)); err != nil {
		return "", err
	}

	scanner := bufio.NewScanner(conn)
	if scanner.Scan() {
		var resp ipcResponse
		if err := json.Unmarshal(scanner.Bytes(), &resp); err != nil {
			return "", err
		}

		if resp.Type == "error" {
			var errPayload struct {
				Message string `json:"message"`
			}
			json.Unmarshal(resp.Payload, &errPayload)
			return "", fmt.Errorf("vibeaura error: %s", errPayload.Message)
		}

		var payload struct {
			Content string `json:"content"`
		}
		if err := json.Unmarshal(resp.Payload, &payload); err != nil {
			return "", err
		}

		return payload.Content, nil
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	return "", fmt.Errorf("no response from vibeaura")
}

func (v *VibeauraSynthesizer) GenerateBuildPlan(ctx context.Context, repoURL string, files []string, readme string) (*BuildPlan, error) {
	prompt := fmt.Sprintf(`Analyze the following repository and generate a build plan.
Repository: %s
Files:
%s

README:
%s

Return ONLY a JSON object with the following structure:
{
  "steps": ["command1", "command2"],
  "bin": "binary_name"
}
Ensure the commands are safe and platform-agnostic where possible, or tailored for the current environment.`, repoURL, strings.Join(files, "\n"), readme)

	resp, err := v.query(ctx, prompt, "ask")
	if err != nil {
		return nil, err
	}

	// Try to find JSON in response in case there's markdown fluff
	jsonStart := strings.Index(resp, "{")
	jsonEnd := strings.LastIndex(resp, "}")
	if jsonStart == -1 || jsonEnd == -1 {
		return nil, fmt.Errorf("invalid AI response: could not find JSON object")
	}
	resp = resp[jsonStart : jsonEnd+1]

	var plan BuildPlan
	if err := json.Unmarshal([]byte(resp), &plan); err != nil {
		return nil, fmt.Errorf("failed to parse build plan: %w\nRaw response: %s", err, resp)
	}

	return &plan, nil
}

func (v *VibeauraSynthesizer) SummarizeUpdates(ctx context.Context, commits []string) (string, error) {
	prompt := fmt.Sprintf("Summarize the following git commits into a human-readable changelog:\n%s", strings.Join(commits, "\n"))
	return v.query(ctx, prompt, "ask")
}

func (v *VibeauraSynthesizer) RedactCommand(ctx context.Context, command string) (string, error) {
	prompt := fmt.Sprintf("Redact any sensitive information (API keys, passwords, PII) from this shell command. Return ONLY the redacted command:\n%s", command)
	return v.query(ctx, prompt, "ask")
}