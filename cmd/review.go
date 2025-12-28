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
		prompt = fmt.Sprintf(`Kamu adalah SENIOR SOFTWARE ARCHITECT dengan keahlian mendalam di %s, code review, dan system design.

TASK: Lakukan code review profesional yang komprehensif untuk kode ini.

ANALISA MENDALAM:
1. ARCHITECTURE & DESIGN PATTERNS: Evaluasi struktur kode, design patterns yang digunakan/dibutuhkan
2. CODE QUALITY: Clean code principles, readability, maintainability
3. SECURITY VULNERABILITIES: Input validation, injection attacks, authentication/authorization issues
4. PERFORMANCE: Bottlenecks, memory leaks, algorithmic complexity (Big O)
5. ERROR HANDLING: Edge cases, exception handling, graceful degradation
6. BEST PRACTICES: Language-specific idioms, modern features usage
7. TESTING: Testability, coverage potential, test cases yang diperlukan
8. SCALABILITY: Apakah kode siap untuk scale? Potential issues di production

FORMAT RESPONSE (RINGKAS & PADAT):

## Ringkasan Eksekutif
Gambaran umum kualitas kode dalam 2-3 kalimat

## Masalah Kritis ⚠️
- [Severity: HIGH/MEDIUM/LOW] Issue - dampak dan solusi singkat

## Keamanan 🔒
- Vulnerability yang ditemukan
- Rekomendasi perbaikan security

## Performa & Optimasi ⚡
- Bottleneck dan inefficiencies
- Saran optimasi konkrit

## Architecture & Design 🏗️
- Design pattern yang bisa diterapkan
- Code structure improvements

## Best Practices & Clean Code ✨
- Pelanggaran convention
- Saran improvement readability

## Testing & Maintainability 🧪
- Test coverage gaps
- Refactoring opportunities

## Skor Kualitas 📊
X/10 - Penjelasan detail berdasarkan Production Readiness, Security, Performance, Maintainability

File: %s
Kode:
%s`, lang, filePath, string(content))
	} else {
		prompt = fmt.Sprintf(`You are a SENIOR SOFTWARE ARCHITECT with deep expertise in %s, code review, and system design.

TASK: Perform a comprehensive professional code review.

IN-DEPTH ANALYSIS:
1. ARCHITECTURE & DESIGN PATTERNS: Evaluate code structure, patterns used/needed
2. CODE QUALITY: Clean code principles, readability, maintainability
3. SECURITY VULNERABILITIES: Input validation, injection attacks, auth/authz issues
4. PERFORMANCE: Bottlenecks, memory leaks, algorithmic complexity (Big O)
5. ERROR HANDLING: Edge cases, exception handling, graceful degradation
6. BEST PRACTICES: Language-specific idioms, modern features usage
7. TESTING: Testability, coverage potential, required test cases
8. SCALABILITY: Is code production-ready? Potential issues at scale

RESPONSE FORMAT (CONCISE & ACTIONABLE):

## Executive Summary
Overall code quality in 2-3 sentences

## Critical Issues ⚠️
- [Severity: HIGH/MEDIUM/LOW] Issue - impact and brief solution

## Security 🔒
- Vulnerabilities found
- Security hardening recommendations

## Performance & Optimization ⚡
- Bottlenecks and inefficiencies
- Concrete optimization suggestions

## Architecture & Design 🏗️
- Applicable design patterns
- Code structure improvements

## Best Practices & Clean Code ✨
- Convention violations
- Readability improvement suggestions

## Testing & Maintainability 🧪
- Test coverage gaps
- Refactoring opportunities

## Quality Score 📊
X/10 - Detailed explanation based on Production Readiness, Security, Performance, Maintainability

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
