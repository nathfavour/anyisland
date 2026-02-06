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
	prompt := fmt.Sprintf(`You are an expert software engineer and build systems specialist for "Anyisland", an AI-powered package manager.
Analyze the following repository and generate a PRECISE build and installation plan.

Repository: %s
Files:
%s

README:
%s

CRITICAL REQUIREMENTS:
1. Return ONLY a valid JSON object. No markdown, no preamble, no explanation.
2. "steps" must be a list of shell commands to build the project.
3. "bin" must be the path to the resulting executable binary relative to the repository root.
4. If the tool is a script (Python, Node.js), point "bin" to the main entry point.
5. "install_dir" is optional.
6. If the repository is NOT a buildable software project (e.g., just documentation, a profile README, or static assets), DO NOT return a plan. Instead, return an empty object {} or an error-like structure (but ideally, AnalyzeDiscretion should have caught this).

Structure:
{
  "steps": ["command1", "command2"],
  "bin": "path/to/binary",
  "toolchain": "go|rust|node|python|flutter|...",
  "install_dir": "/optional/custom/path"
}
`, repoURL, strings.Join(files, "\n"), readme)

	resp, err := v.query(ctx, prompt, "ask")
// ... (rest of the function remains same but I'll provide the context)
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

	if resp == "{}" {
		return nil, fmt.Errorf("AI could not determine a valid build plan for this repository")
	}

	var plan BuildPlan
// ...
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



func (v *VibeauraSynthesizer) AnalyzeDiscretion(ctx context.Context, files []string, readme string) (*DiscretionResult, error) {



	prompt := fmt.Sprintf(`You are an expert security and software auditor for "Anyisland".



Determine if the following repository is a FUNCTIONAL command-line tool or application that should be installed.







DO NOT allow:



- Personal profile repositories (e.g., nathfavour/nathfavour).



- Pure documentation or tutorial repositories.



- Collections of "awesome" lists or bookmarks.



- Repositories that are just configuration files (dotfiles) without an installer.



- Static websites or asset collections.







ALLOW:



- CLI tools (compilers, linters, utilities).



- GUI applications with clear build steps.



- Libraries that include a CLI component.







Files:



%s







README snippet:



%.1000s







Return ONLY a JSON object: {"allowed": true/false, "reason": "concise explanation"}`, strings.Join(files, "\n"), readme)







	resp, err := v.query(ctx, prompt, "plan")

	if err != nil {

		return nil, err

	}



	jsonStart := strings.Index(resp, "{")

	jsonEnd := strings.LastIndex(resp, "}")

	if jsonStart == -1 || jsonEnd == -1 {

		return nil, fmt.Errorf("invalid AI response: could not find JSON object")

	}

	resp = resp[jsonStart : jsonEnd+1]



	var result DiscretionResult

	if err := json.Unmarshal([]byte(resp), &result); err != nil {

		return nil, err

	}

	return &result, nil

}



func (v *VibeauraSynthesizer) DebugBuildFailure(ctx context.Context, log string, manifest interface{}) (string, error) {

	manifestJSON, _ := json.MarshalIndent(manifest, "", "  ")

	prompt := fmt.Sprintf(`The build for a tool failed. Analyze the error log and the manifest, then suggest a fix or explain the cause.

Manifest:

%s



Error Log:

%s



Provide a concise, helpful explanation and if possible, the corrected command or steps.`, string(manifestJSON), log)



		return v.query(ctx, prompt, "plan")



	}



	



	func (v *VibeauraSynthesizer) DiscoverTool(ctx context.Context, query string) (string, error) {



		prompt := fmt.Sprintf(`The user is looking for a CLI tool for: "%s"



	Find the most reputable and relevant GitHub repository for this purpose. 



	Return ONLY the GitHub URL (e.g., https://github.com/user/repo). 



	If you are unsure, return "NONE".`, query)



	



		return v.query(ctx, prompt, "ask")



	}



	



	func (v *VibeauraSynthesizer) ExplainTool(ctx context.Context, name string, manifest interface{}, readme string) (string, error) {



		manifestJSON, _ := json.MarshalIndent(manifest, "", "  ")



		prompt := fmt.Sprintf(`Provide a concise (max 5 lines) explanation of what the tool "%s" does and how to run it.



	Manifest:



	%s



	



	README snippet:



	%.1000s



	



	Focus on the primary command and its most common use case.`, name, string(manifestJSON), readme)



	



		return v.query(ctx, prompt, "ask")



	}



	
