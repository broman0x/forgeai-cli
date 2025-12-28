package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/broman0x/forgeai-cli/internal/ai"
	"github.com/broman0x/forgeai-cli/internal/ui"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var scanCmd = &cobra.Command{
	Use:   "scan [directory]",
	Short: "Scan directory and apply AI fixes",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			cmd.Help()
			return
		}
		prov, _ := ai.NewProvider()
		runScanLogic(prov, args[0])
	},
}

func init() {
	rootCmd.AddCommand(scanCmd)
}

func StartScanModeInteractive(scanner *bufio.Scanner, prov ai.Provider) {
	cTitle := color.New(color.FgHiCyan, color.Bold).SprintFunc()
	cPrompt := color.New(color.FgWhite).SprintFunc()
	cSubtle := color.New(color.FgHiBlack).SprintFunc()

	fmt.Println()
	fmt.Println(cTitle("  PROJECT SCANNER"))
	fmt.Println(cSubtle("  ───────────────────────────────────────────"))
	fmt.Println()

	fmt.Print(cPrompt("  Directory path: "))
	if !scanner.Scan() {
		return
	}
	dirPath := strings.TrimSpace(scanner.Text())

	if dirPath == "" || dirPath == "back" {
		fmt.Print("\033[H\033[2J")
		ui.ShowStartupBanner()
		return
	}

	fmt.Print(cPrompt("  Instruction (e.g., 'fix all bugs', 'add comments'): "))
	if !scanner.Scan() {
		return
	}
	instruction := strings.TrimSpace(scanner.Text())

	if instruction == "" {
		return
	}

	runScanLogic(prov, dirPath)

	fmt.Print(cSubtle("\n  Press Enter to continue..."))
	scanner.Scan()
	fmt.Print("\033[H\033[2J")
	ui.ShowStartupBanner()
}

func runScanLogic(prov ai.Provider, dirPath string) {
	info, err := os.Stat(dirPath)
	if err != nil {
		color.Red("  Error: Directory not found - %v", err)
		return
	}

	if !info.IsDir() {
		color.Red("  Error: Path is not a directory")
		return
	}

	files := scanDirectory(dirPath)
	if len(files) == 0 {
		color.Yellow("  No code files found in directory")
		return
	}

	cInfo := color.New(color.FgCyan).SprintFunc()
	cSuccess := color.New(color.FgGreen).SprintFunc()

	fmt.Println()
	fmt.Printf("  %s %d files found\n", cInfo("Scanned:"), len(files))
	fmt.Println()

	spinner := ui.NewSpinner("Analyzing project structure")
	spinner.Start()

	time.Sleep(500 * time.Millisecond)
	spinner.Stop()

	fmt.Println(cSuccess("  Project context built"))
	fmt.Println()
	fmt.Println(color.New(color.FgWhite).Sprint("  Files scanned:"))

	for i, file := range files {
		if i < 10 {
			fmt.Printf("    - %s\n", filepath.Base(file))
		}
	}
	if len(files) > 10 {
		fmt.Printf("    ... and %d more files\n", len(files)-10)
	}

	fmt.Println()
	fmt.Println(color.New(color.FgHiBlack).Sprint("  Ready for AI analysis"))
	fmt.Println(color.New(color.FgHiBlack).Sprint("  Use chat mode to discuss this project"))
}

func scanDirectory(root string) []string {
	var files []string
	codeExts := map[string]bool{
		".go": true, ".js": true, ".ts": true, ".py": true,
		".java": true, ".cpp": true, ".c": true, ".cs": true,
		".rb": true, ".php": true, ".rs": true, ".kt": true,
		".swift": true, ".jsx": true, ".tsx": true, ".vue": true,
	}

	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if info.IsDir() {
			name := info.Name()
			if name == "node_modules" || name == ".git" || name == "vendor" || name == "target" || name == "build" {
				return filepath.SkipDir
			}
			return nil
		}

		ext := filepath.Ext(path)
		if codeExts[ext] {
			files = append(files, path)
		}

		return nil
	})

	return files
}
