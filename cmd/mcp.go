package cmd

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/urfave/cli/v2"

	"rmk/logger"
	"rmk/util"
)

const (
	mcpServerName     = "rmk"
	mcpDefaultTimeout = 5 * time.Minute
)

type mcpWorkDirArgs struct {
	WorkDir string `json:"workdir,omitempty" jsonschema:"optional absolute path to the RMK project root; defaults to the MCP server process working directory"`
}

type mcpReleaseArgs struct {
	WorkDir  string   `json:"workdir,omitempty" jsonschema:"optional absolute path to the RMK project root; defaults to the MCP server process working directory"`
	Selector []string `json:"selector,omitempty" jsonschema:"optional Helmfile release label selectors (foo=bar or foo!=bar)"`
}

type mcpProjectGenerateArgs struct {
	WorkDir           string   `json:"workdir,omitempty" jsonschema:"optional absolute path to the RMK project root; defaults to the MCP server process working directory"`
	Scope             []string `json:"scope,omitempty" jsonschema:"optional list of project scopes"`
	Environment       []string `json:"environment,omitempty" jsonschema:"optional list of project environments (requires root-domain config)"`
	CreateSopsAgeKeys bool     `json:"create_sops_age_keys,omitempty" jsonschema:"when true, create SOPS age keys during project generate"`
	Confirm           bool     `json:"confirm" jsonschema:"must be true to run project generate (explicit gate)"`
}

func mcpCommands() *cli.Command {
	return &cli.Command{
		Name:  "mcp",
		Usage: "Model Context Protocol (MCP) server for AI assistants",
		Subcommands: []*cli.Command{
			{
				Name:         "serve",
				Usage:        "Start MCP server over stdio (for Cursor, Claude Desktop, and other MCP clients)",
				Aliases:      []string{"s"},
				Category:     "mcp",
				BashComplete: util.ShellCompleteCustomOutput,
				Action:       mcpServeAction(),
			},
			{
				Name:         "status",
				Usage:        "Show MCP readiness (binary path, project detection) and client config snippets",
				Category:     "mcp",
				BashComplete: util.ShellCompleteCustomOutput,
				Action:       mcpStatusAction(),
			},
		},
	}
}

func mcpServeAction() cli.ActionFunc {
	return func(c *cli.Context) error {
		if err := util.ValidateNArg(c, 0); err != nil {
			return err
		}

		// MCP stdio uses stdout for JSON-RPC; keep all RMK logs on stderr.
		logger.Init(c.String("log-format"), "error")

		version := mcpAppVersion(c)
		server := mcp.NewServer(&mcp.Implementation{Name: mcpServerName, Version: version}, nil)
		registerMCPTools(server)

		fmt.Fprintf(os.Stderr, "rmk mcp serve started (version %s, transport stdio)\n", version)
		return server.Run(c.Context, &mcp.StdioTransport{})
	}
}

func mcpStatusAction() cli.ActionFunc {
	return func(c *cli.Context) error {
		if err := util.ValidateNArg(c, 0); err != nil {
			return err
		}

		exe, err := os.Executable()
		if err != nil {
			return err
		}
		exe, err = filepath.EvalSymlinks(exe)
		if err != nil {
			exe, _ = os.Executable()
		}

		cwd, err := os.Getwd()
		if err != nil {
			return err
		}

		projectFile := filepath.Join(cwd, util.TenantProjectFile)
		hasProject := util.IsExists(projectFile, true)

		fmt.Printf("RMK MCP status\n")
		fmt.Printf("==============\n")
		fmt.Printf("binary:          %s\n", exe)
		fmt.Printf("version:         %s\n", mcpAppVersion(c))
		fmt.Printf("cwd:             %s\n", cwd)
		fmt.Printf("project.yaml:    %s\n", boolStatus(hasProject))
		fmt.Printf("transport:       stdio\n")
		fmt.Printf("serve command:   %s mcp serve\n", exe)
		fmt.Printf("\n")
		fmt.Printf("Cursor (~/.cursor/mcp.json or project .cursor/mcp.json):\n")
		fmt.Printf("%s\n", mcpClientSnippet(exe, cwd))
		fmt.Printf("\nClaude Desktop (claude_desktop_config.json):\n")
		fmt.Printf("%s\n", mcpClientSnippet(exe, cwd))

		if !hasProject {
			fmt.Printf("\nNote: no %s in cwd; start the MCP client with cwd set to an RMK project root,\n", util.TenantProjectFile)
			fmt.Printf("or pass workdir to tools that support it.\n")
		}

		return nil
	}
}

func registerMCPTools(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "project_status",
		Description: "Report RMK project detection for a working directory (project.yaml presence, cwd).",
	}, mcpToolProjectStatus)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "config_view",
		Description: "View RMK configuration for the selected environment (wraps `rmk config view`). Does not expose decrypted secrets.",
	}, mcpToolConfigView)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "release_list",
		Description: "List Helmfile releases for the project (wraps `rmk release list --skip-context-switch`). Prefer output=json for structured results.",
	}, mcpToolReleaseList)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "release_template",
		Description: "Render Helmfile templates for releases (wraps `rmk release template --skip-context-switch`). Read-only inspect/validate.",
	}, mcpToolReleaseTemplate)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "release_show_dag",
		Description: "Show simplified release DAGs (wraps `rmk release show-dag --skip-context-switch`). Helmfile v1.1.x only.",
	}, mcpToolReleaseShowDag)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "project_generate",
		Description: "Generate project directories and file structure (wraps `rmk project generate`). Requires confirm=true. Does not manage clusters.",
	}, mcpToolProjectGenerate)
}

func mcpToolProjectStatus(_ context.Context, _ *mcp.CallToolRequest, args mcpWorkDirArgs) (*mcp.CallToolResult, any, error) {
	workDir, err := resolveMCPWorkDir(args.WorkDir)
	if err != nil {
		return mcpToolErr(err.Error()), nil, nil
	}

	projectFile := filepath.Join(workDir, util.TenantProjectFile)
	hasProject := util.IsExists(projectFile, true)

	var b strings.Builder
	fmt.Fprintf(&b, "workdir: %s\n", workDir)
	fmt.Fprintf(&b, "project.yaml: %s\n", boolStatus(hasProject))
	if hasProject {
		fmt.Fprintf(&b, "project_file: %s\n", projectFile)
	}

	return mcpToolOK(b.String()), nil, nil
}

func mcpToolConfigView(ctx context.Context, _ *mcp.CallToolRequest, args mcpWorkDirArgs) (*mcp.CallToolResult, any, error) {
	workDir, err := resolveMCPWorkDir(args.WorkDir)
	if err != nil {
		return mcpToolErr(err.Error()), nil, nil
	}
	return runRMKTool(ctx, workDir, "config", "view")
}

type mcpReleaseListArgs struct {
	mcpReleaseArgs
	Output string `json:"output,omitempty" jsonschema:"optional output format: short or json (default json)"`
}

func mcpToolReleaseList(ctx context.Context, _ *mcp.CallToolRequest, args mcpReleaseListArgs) (*mcp.CallToolResult, any, error) {
	workDir, err := resolveMCPWorkDir(args.WorkDir)
	if err != nil {
		return mcpToolErr(err.Error()), nil, nil
	}

	output := args.Output
	if output == "" {
		output = "json"
	}

	cmdArgs := []string{"release", "list", "--skip-context-switch", "--output", output}
	cmdArgs = append(cmdArgs, selectorFlags(args.Selector)...)
	return runRMKTool(ctx, workDir, cmdArgs...)
}

func mcpToolReleaseTemplate(ctx context.Context, _ *mcp.CallToolRequest, args mcpReleaseArgs) (*mcp.CallToolResult, any, error) {
	workDir, err := resolveMCPWorkDir(args.WorkDir)
	if err != nil {
		return mcpToolErr(err.Error()), nil, nil
	}

	cmdArgs := []string{"release", "template", "--skip-context-switch"}
	cmdArgs = append(cmdArgs, selectorFlags(args.Selector)...)
	return runRMKTool(ctx, workDir, cmdArgs...)
}

func mcpToolReleaseShowDag(ctx context.Context, _ *mcp.CallToolRequest, args mcpReleaseArgs) (*mcp.CallToolResult, any, error) {
	workDir, err := resolveMCPWorkDir(args.WorkDir)
	if err != nil {
		return mcpToolErr(err.Error()), nil, nil
	}

	cmdArgs := []string{"release", "show-dag", "--skip-context-switch"}
	cmdArgs = append(cmdArgs, selectorFlags(args.Selector)...)
	return runRMKTool(ctx, workDir, cmdArgs...)
}

func mcpToolProjectGenerate(ctx context.Context, _ *mcp.CallToolRequest, args mcpProjectGenerateArgs) (*mcp.CallToolResult, any, error) {
	if !args.Confirm {
		return mcpToolErr("project_generate requires confirm=true"), nil, nil
	}

	workDir, err := resolveMCPWorkDir(args.WorkDir)
	if err != nil {
		return mcpToolErr(err.Error()), nil, nil
	}

	cmdArgs := []string{"project", "generate"}
	for _, scope := range args.Scope {
		cmdArgs = append(cmdArgs, "--scope", scope)
	}
	for _, env := range args.Environment {
		cmdArgs = append(cmdArgs, "--environment", env)
	}
	if args.CreateSopsAgeKeys {
		cmdArgs = append(cmdArgs, "--create-sops-age-keys")
	}

	return runRMKTool(ctx, workDir, cmdArgs...)
}

func runRMKTool(ctx context.Context, workDir string, args ...string) (*mcp.CallToolResult, any, error) {
	stdout, stderr, err := execRMK(ctx, workDir, args...)
	out := strings.TrimSpace(stdout)
	errOut := strings.TrimSpace(stderr)

	if err != nil {
		msg := fmt.Sprintf("command failed: rmk %s\n%s", strings.Join(args, " "), err.Error())
		if errOut != "" {
			msg += "\n" + errOut
		}
		if out != "" {
			msg += "\n" + out
		}
		return mcpToolErr(msg), nil, nil
	}

	if out == "" && errOut != "" {
		out = errOut
	}
	if out == "" {
		out = "(no output)"
	}
	return mcpToolOK(out), nil, nil
}

func execRMK(ctx context.Context, workDir string, args ...string) (string, string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", "", err
	}

	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, mcpDefaultTimeout)
		defer cancel()
	}

	// Keep child process logs off stdout so captured command output stays clean.
	fullArgs := append([]string{"--log-level", "error"}, args...)
	cmd := exec.CommandContext(ctx, exe, fullArgs...)
	cmd.Dir = workDir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	return stdout.String(), stderr.String(), err
}

func resolveMCPWorkDir(requested string) (string, error) {
	if strings.TrimSpace(requested) == "" {
		return os.Getwd()
	}
	return filepath.Abs(requested)
}

func selectorFlags(selectors []string) []string {
	var flags []string
	for _, s := range selectors {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		flags = append(flags, "--selector", s)
	}
	return flags
}

func mcpToolOK(text string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: text}},
	}
}

func mcpToolErr(text string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		IsError: true,
		Content: []mcp.Content{&mcp.TextContent{Text: text}},
	}
}

func mcpAppVersion(c *cli.Context) string {
	if c != nil && c.App != nil && c.App.Metadata != nil {
		if v, ok := c.App.Metadata["version"]; ok && fmt.Sprint(v) != "" {
			return fmt.Sprint(v)
		}
	}
	if c != nil && c.App != nil && c.App.Version != "" {
		return c.App.Version
	}
	return "dev"
}

func boolStatus(v bool) string {
	if v {
		return "yes"
	}
	return "no"
}

func mcpClientSnippet(exe, cwd string) string {
	return fmt.Sprintf(`{
  "mcpServers": {
    "rmk": {
      "command": %q,
      "args": ["mcp", "serve"],
      "cwd": %q
    }
  }
}`, exe, cwd)
}
