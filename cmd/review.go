package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/broman0x/forgeai-cli/internal/ai"
	"github.com/broman0x/forgeai-cli/internal/ui"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var reviewCmd = &cobra.Command{
	Use:   "review [file]",
	Short: "Review a source code file",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			cmd.Help()
			return
		}
		prov, _ := ai.NewProvider()
		runReviewLogic(prov, args[0])
	},
}

func init() {
	rootCmd.AddCommand(reviewCmd)
}

func StartReviewModeInteractive(scanner *bufio.Scanner, prov ai.Provider) {
	ui.PrintHeader("CODE SECURITY REVIEW")

	fmt.Print("  Target File (e.g., main.go): ")
	if !scanner.Scan() {
		return
	}
	filePath := strings.TrimSpace(scanner.Text())

	if filePath == "" || filePath == "back" {
		fmt.Print("\033[H\033[2J")
		ui.ShowStartupBanner()
		return
	}

	runReviewLogic(prov, filePath)

	color.New(color.Faint).Print("\n  [Press Enter to return]")
	scanner.Scan()
	fmt.Print("\033[H\033[2J")
	ui.ShowStartupBanner()
}

func runReviewLogic(prov ai.Provider, filePath string) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		color.Red("  [ERROR] File Access: %v", err)
		return
	}

	fmt.Printf("\n  %s %s\n", color.YellowString("Scanning:"), filePath)
	fmt.Printf("  %s %s\n", color.CyanString("Engine:  "), prov.Name())

	prompt := fmt.Sprintf(`
You are a Senior Engineer. Review this code: "%s".
Structure your answer exactly like this:

### Summary
(Brief overview)

### 1. Bugs
* (Point 1)
* (Point 2)

### 2. Security
* (Point 1)

### 3. Improvements
* (Point 1)

Code:
%s
`, filePath, string(content))

	fmt.Print(color.New(color.Faint).Sprint("\n  processing analysis..."))
	resp, err := prov.Send(prompt)
	fmt.Print("\r                      \r")

	if err != nil {
		color.Red("  [ERROR] Analysis Failed: %v", err)
		return
	}

	printFormattedReport(resp)
}
func printFormattedReport(markdown string) {
	fmt.Println()
	color.New(color.BgHiBlue, color.FgWhite, color.Bold).Println("  FORGE AI REPORT  ")
	fmt.Println()
	cHeader := color.New(color.FgHiCyan, color.Bold).SprintFunc()
	cBold := color.New(color.FgWhite, color.Bold).SprintFunc()
	cList := color.New(color.FgHiWhite).SprintFunc()
	cNormal := color.New(color.FgWhite).SprintFunc()

	lines := strings.Split(markdown, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			fmt.Println()
			continue
		}

		if strings.HasPrefix(line, "###") || strings.HasPrefix(line, "**") {
			clean := strings.ReplaceAll(line, "#", "")
			clean = strings.ReplaceAll(clean, "*", "")
			fmt.Printf("  %s\n", cHeader(strings.ToUpper(strings.TrimSpace(clean))))
			fmt.Println(color.New(color.FgHiBlack).Sprint("  ────────────────────"))

		} else if strings.HasPrefix(line, "* ") || strings.HasPrefix(line, "- ") {
			clean := line[2:]
			if strings.Contains(clean, "**") {
				parts := strings.Split(clean, "**")
				fmt.Print("   • ")
				for i, p := range parts {
					if i%2 == 1 {
						fmt.Print(cBold(p))
					} else {
						fmt.Print(cList(p))
					}
				}
				fmt.Println()
			} else {
				fmt.Printf("   • %s\n", cList(clean))
			}

		} else {
			if strings.HasPrefix(line, "```") {
				continue
			}
			fmt.Printf("  %s\n", cNormal(line))
		}

		time.Sleep(2 * time.Millisecond)
	}
	fmt.Println()
}
