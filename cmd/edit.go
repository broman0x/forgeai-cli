package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/broman0x/forgeai-cli/internal/ai"
	"github.com/broman0x/forgeai-cli/internal/ui"
	"github.com/fatih/color"
	"github.com/pmezard/go-difflib/difflib"
	"github.com/spf13/cobra"
)

var editCmd = &cobra.Command{
	Use:   "edit [file] \"instruction\"",
	Short: "AI Code Editor with Diff",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 2 {
			cmd.Help()
			return
		}
		prov, _ := ai.NewProvider()
		runEditLogic(prov, args[0], args[1], nil)
	},
}

func init() {
	rootCmd.AddCommand(editCmd)
}

func StartEditModeInteractive(scanner *bufio.Scanner, prov ai.Provider) {
	cTitle := color.New(color.FgHiCyan, color.Bold).SprintFunc()
	cPrompt := color.New(color.FgWhite).SprintFunc()
	cSubtle := color.New(color.FgHiBlack).SprintFunc()

	fmt.Println()
	fmt.Println(cTitle("  CODE EDITOR"))
	fmt.Println(cSubtle("  ───────────────────────────────────────────"))
	fmt.Println()

	fmt.Print(cPrompt("  File path: "))
	if !scanner.Scan() {
		return
	}
	filePath := strings.TrimSpace(scanner.Text())

	if filePath == "" || filePath == "back" {
		fmt.Print("\033[H\033[2J")
		ui.ShowStartupBanner()
		return
	}

	fmt.Print(cPrompt("  Instruction: "))
	if !scanner.Scan() {
		return
	}
	instruction := strings.TrimSpace(scanner.Text())

	if instruction == "" {
		return
	}

	runEditLogic(prov, filePath, instruction, scanner)

	fmt.Print(cSubtle("\n  Press Enter to continue..."))
	scanner.Scan()
	fmt.Print("\033[H\033[2J")
	ui.ShowStartupBanner()
}

func runEditLogic(prov ai.Provider, filePath, instruction string, scanner *bufio.Scanner) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		color.Red("  Error: %v", err)
		return
	}

	ext := filepath.Ext(filePath)
	lang := detectLanguage(ext)

	spinner := ui.NewSpinner("Processing request")
	spinner.Start()

	prompt := fmt.Sprintf(`You are an expert programmer. Modify the following %s code according to this instruction: "%s"

IMPORTANT: Return ONLY the modified code without any explanations, comments, or markdown formatting.

File: %s

Code:
%s`, lang, instruction, filePath, string(content))

	newCode, err := prov.Send(prompt)
	spinner.Stop()

	if err != nil {
		color.Red("  Error: %v", err)
		return
	}

	newCode = cleanMarkdown(newCode)

	diff := difflib.UnifiedDiff{
		A:        difflib.SplitLines(string(content)),
		B:        difflib.SplitLines(newCode),
		FromFile: "Original",
		ToFile:   "Modified",
		Context:  3,
	}
	diffText, _ := difflib.GetUnifiedDiffString(diff)

	if strings.TrimSpace(diffText) == "" {
		color.Yellow("\n  No changes detected")
		return
	}

	cTitle := color.New(color.FgHiCyan, color.Bold).SprintFunc()
	cSubtle := color.New(color.FgHiBlack).SprintFunc()
	fmt.Printf("\n  %s\n", cTitle("DIFF PREVIEW"))
	fmt.Println(cSubtle("  ───────────────────────────────────────────"))
	printDiff(diffText)
	fmt.Println(cSubtle("  ───────────────────────────────────────────"))

	if confirm(scanner, "\n  Apply changes? [y/N]") {
		os.WriteFile(filePath, []byte(newCode), 0644)
		color.Green("  Changes applied successfully")
	} else {
		color.Yellow("  Changes discarded")
	}
}

func detectLanguage(ext string) string {
	langMap := map[string]string{
		".js":    "JavaScript",
		".ts":    "TypeScript",
		".py":    "Python",
		".go":    "Go",
		".java":  "Java",
		".cpp":   "C++",
		".c":     "C",
		".cs":    "C#",
		".rb":    "Ruby",
		".php":   "PHP",
		".rs":    "Rust",
		".kt":    "Kotlin",
		".swift": "Swift",
		".jsx":   "React JSX",
		".tsx":   "React TSX",
		".vue":   "Vue",
		".html":  "HTML",
		".css":   "CSS",
		".scss":  "SCSS",
		".sql":   "SQL",
		".sh":    "Shell",
	}

	if lang, ok := langMap[strings.ToLower(ext)]; ok {
		return lang
	}
	return "code"
}

func cleanMarkdown(code string) string {
	lines := strings.Split(code, "\n")
	var out []string
	inCodeBlock := false

	for _, l := range lines {
		trimmed := strings.TrimSpace(l)
		if strings.HasPrefix(trimmed, "```") {
			inCodeBlock = !inCodeBlock
			continue
		}
		if !inCodeBlock || trimmed != "" {
			out = append(out, l)
		}
	}
	return strings.Join(out, "\n")
}

func printDiff(diff string) {
	cAdd := color.New(color.FgGreen)
	cDel := color.New(color.FgRed)
	cInfo := color.New(color.FgCyan)
	cNormal := color.New(color.FgWhite)

	lines := strings.Split(diff, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "+++") || strings.HasPrefix(line, "---") {
			cInfo.Printf("  %s\n", line)
		} else if strings.HasPrefix(line, "+") {
			cAdd.Printf("  %s\n", line)
		} else if strings.HasPrefix(line, "-") {
			cDel.Printf("  %s\n", line)
		} else if strings.HasPrefix(line, "@@") {
			cInfo.Printf("  %s\n", line)
		} else {
			cNormal.Printf("  %s\n", line)
		}
	}
}

func confirm(scanner *bufio.Scanner, q string) bool {
	fmt.Print(q + " ")
	if scanner != nil {
		if !scanner.Scan() {
			return false
		}
		return strings.ToLower(strings.TrimSpace(scanner.Text())) == "y"
	}
	cliScanner := bufio.NewScanner(os.Stdin)
	if !cliScanner.Scan() {
		return false
	}
	return strings.ToLower(strings.TrimSpace(cliScanner.Text())) == "y"
}
