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
	cTitle := color.New(color.FgHiCyan, color.Bold).SprintFunc()
	cPrompt := color.New(color.FgWhite).SprintFunc()
	cSubtle := color.New(color.FgHiBlack).SprintFunc()

	fmt.Println()
	fmt.Println(cTitle("  CODE REVIEW"))
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

	fmt.Print(cPrompt("  Language [1=English, 2=Indonesian]: "))
	if !scanner.Scan() {
		return
	}
	langChoice := strings.TrimSpace(scanner.Text())

	reviewLang := "english"
	if langChoice == "2" {
		reviewLang = "indonesian"
	}

	runReviewLogicWithLang(prov, filePath, reviewLang)

	fmt.Print(cSubtle("\n  Press Enter to continue..."))
	scanner.Scan()
	fmt.Print("\033[H\033[2J")
	ui.ShowStartupBanner()
}

func runReviewLogic(prov ai.Provider, filePath string) {
	runReviewLogicWithLang(prov, filePath, "english")
}

func runReviewLogicWithLang(prov ai.Provider, filePath string, language string) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		color.Red("  Error: %v", err)
		return
	}

	ext := filepath.Ext(filePath)
	lang := detectLanguageForReview(ext)

	cInfo := color.New(color.FgCyan).SprintFunc()
	cSubtle := color.New(color.FgHiBlack).SprintFunc()
	cSection := color.New(color.FgCyan, color.Bold).SprintFunc()
	cBullet := color.New(color.FgMagenta).SprintFunc()
	cText := color.New(color.FgWhite).SprintFunc()

	fmt.Println()
	fmt.Printf("  %s %s\n", cInfo("File:"), filePath)
	fmt.Printf("  %s %s\n", cInfo("Language:"), lang)
	fmt.Printf("  %s %s\n", cInfo("Engine:"), prov.Name())
	fmt.Println()

	spinner := ui.NewSpinner("Analyzing code")
	spinner.Start()

	var prompt string
	if language == "indonesian" {
		prompt = fmt.Sprintf(`Kamu adalah senior code reviewer. Analisa kode %s ini dan berikan review terstruktur.

PENTING: Berikan response yang SINGKAT dan PADAT. Setiap poin maksimal 1 kalimat. Format PERSIS seperti ini:

## Ringkasan
Gambaran singkat 1-2 kalimat

## Masalah Kritis
- Masalah 1 (singkat)
- Masalah 2 (singkat)

## Keamanan
- Poin keamanan 1 (singkat)

## Performa
- Poin performa 1 (singkat)

## Saran Improvement
- Saran 1 (singkat)
- Saran 2 (singkat)

## Skor Kualitas
X/10 - Alasan singkat

File: %s
Kode:
%s`, lang, filePath, string(content))
	} else {
		prompt = fmt.Sprintf(`You are a senior code reviewer. Analyze this %s code and provide a structured review.

IMPORTANT: Be VERY CONCISE. Each point should be max 1 sentence. Format EXACTLY like this:

## Summary
Brief 1-2 sentence overview

## Critical Issues
- Issue 1 (brief)
- Issue 2 (brief)

## Security
- Security point 1 (brief)

## Performance
- Performance point 1 (brief)

## Improvements
- Improvement 1 (brief)
- Improvement 2 (brief)

## Quality Score
X/10 - Brief reason

File: %s
Code:
%s`, lang, filePath, string(content))
	}

	resp, err := prov.Send(prompt)
	spinner.Stop()

	if err != nil {
		color.Red("  Error: %v", err)
		return
	}

	fmt.Println(cSubtle("  ───────────────────────────────────────────"))
	fmt.Println()

	lines := strings.Split(resp, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)

		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "```") {
			continue
		}

		if strings.HasPrefix(line, "## ") {
			section := strings.TrimPrefix(line, "## ")
			fmt.Printf("\n  %s\n", cSection(strings.ToUpper(section)))
		} else if strings.HasPrefix(line, "# ") {
			section := strings.TrimPrefix(line, "# ")
			fmt.Printf("\n  %s\n", cSection(strings.ToUpper(section)))
		} else if strings.HasPrefix(line, "- ") || strings.HasPrefix(line, "* ") {
			item := strings.TrimPrefix(line, "- ")
			item = strings.TrimPrefix(item, "* ")
			item = cleanText(item)

			wrapped := wrapText(item, 65)
			for i, wrappedLine := range wrapped {
				if i == 0 {
					fmt.Printf("  %s %s\n", cBullet(">"), cText(wrappedLine))
				} else {
					fmt.Printf("    %s\n", cText(wrappedLine))
				}
			}
		} else {
			cleaned := cleanText(line)
			wrapped := wrapText(cleaned, 70)
			for _, wrappedLine := range wrapped {
				fmt.Printf("  %s\n", cText(wrappedLine))
			}
		}
	}

	fmt.Println()
	fmt.Println(cSubtle("  ───────────────────────────────────────────"))
}

func cleanText(text string) string {
	text = strings.ReplaceAll(text, "**", "")
	text = strings.ReplaceAll(text, "__", "")
	text = strings.ReplaceAll(text, "`", "")
	text = strings.ReplaceAll(text, "────────────────────", "")
	return strings.TrimSpace(text)
}

func wrapText(text string, maxWidth int) []string {
	if len(text) <= maxWidth {
		return []string{text}
	}

	var lines []string
	words := strings.Fields(text)
	currentLine := ""

	for _, word := range words {
		testLine := currentLine
		if testLine != "" {
			testLine += " "
		}
		testLine += word

		if len(testLine) <= maxWidth {
			currentLine = testLine
		} else {
			if currentLine != "" {
				lines = append(lines, currentLine)
			}
			currentLine = word
		}
	}

	if currentLine != "" {
		lines = append(lines, currentLine)
	}

	return lines
}

func detectLanguageForReview(ext string) string {
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
	return "Unknown"
}
