package cmd

import (
	"bufio"
	"fmt"
	"os"
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
	ui.PrintHeader("AI CODE EDITOR")

	fmt.Print("  1. Enter filename path (or 'back'): ")
	if !scanner.Scan() {
		return
	}
	filePath := strings.TrimSpace(scanner.Text())

	if filePath == "" || filePath == "back" {
		fmt.Print("\033[H\033[2J")
		ui.ShowStartupBanner()
		return
	}

	fmt.Print("  2. Enter Instruction: ")
	if !scanner.Scan() {
		return
	}
	instruction := strings.TrimSpace(scanner.Text())

	if instruction == "" {
		return
	}

	runEditLogic(prov, filePath, instruction, scanner)

	color.New(color.Faint).Print("\n  [Press Enter to return]")
	scanner.Scan()
	fmt.Print("\033[H\033[2J")
	ui.ShowStartupBanner()
}

func runEditLogic(prov ai.Provider, filePath, instruction string, scanner *bufio.Scanner) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		color.Red("  ✖ File error: %v", err)
		return
	}

	fmt.Println("  Processing...")
	prompt := fmt.Sprintf("Rewrite code. Instruction: %s. Return ONLY code.\nCode:\n%s", instruction, string(content))

	newCode, err := prov.Send(prompt)
	if err != nil {
		color.Red("Error: %v", err)
		return
	}

	newCode = cleanMarkdown(newCode)

	diff := difflib.UnifiedDiff{
		A:        difflib.SplitLines(string(content)),
		B:        difflib.SplitLines(newCode),
		FromFile: "Original",
		ToFile:   "Proposed",
		Context:  3,
	}
	text, _ := difflib.GetUnifiedDiffString(diff)

	if strings.TrimSpace(text) == "" {
		color.Yellow("  No changes proposed.")
		return
	}

	fmt.Println("\n" + color.New(color.Bold).Sprint("DIFF PREVIEW:"))
	printDiff(text)

	if confirm(scanner, "\n  Apply changes?") {
		os.WriteFile(filePath, []byte(newCode), 0644)
		color.Green("  ✔ Saved.")
	} else {
		color.Yellow("  ✖ Discarded.")
	}
}

func cleanMarkdown(code string) string {
	lines := strings.Split(code, "\n")
	var out []string
	for _, l := range lines {
		if !strings.HasPrefix(strings.TrimSpace(l), "```") {
			out = append(out, l)
		}
	}
	return strings.Join(out, "\n")
}

func printDiff(diff string) {
	lines := strings.Split(diff, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++") {
			color.Green(line)
		} else if strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "---") {
			color.Red(line)
		} else {
			fmt.Println(line)
		}
	}
}

func confirm(scanner *bufio.Scanner, q string) bool {
	fmt.Print(q + " [y/N]: ")
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
