package ui

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/fatih/color"
)

type MarkdownRenderer struct {
	heading1   func(a ...interface{}) string
	heading2   func(a ...interface{}) string
	heading3   func(a ...interface{}) string
	bold       func(a ...interface{}) string
	italic     func(a ...interface{}) string
	code       func(a ...interface{}) string
	codeBlock  func(a ...interface{}) string
	link       func(a ...interface{}) string
	quote      func(a ...interface{}) string
	listBullet func(a ...interface{}) string
	normal     func(a ...interface{}) string
}

func NewMarkdownRenderer() *MarkdownRenderer {
	return &MarkdownRenderer{
		heading1:   color.New(color.FgHiCyan, color.Bold).SprintFunc(),
		heading2:   color.New(color.FgCyan, color.Bold).SprintFunc(),
		heading3:   color.New(color.FgCyan).SprintFunc(),
		bold:       color.New(color.Bold).SprintFunc(),
		italic:     color.New(color.Italic).SprintFunc(),
		code:       color.New(color.FgYellow).SprintFunc(),
		codeBlock:  color.New(color.FgHiBlack, color.BgHiWhite).SprintFunc(),
		link:       color.New(color.FgBlue, color.Underline).SprintFunc(),
		quote:      color.New(color.FgHiBlack).SprintFunc(),
		listBullet: color.New(color.FgHiMagenta).SprintFunc(),
		normal:     color.New(color.FgWhite).SprintFunc(),
	}
}

func (m *MarkdownRenderer) Render(text string) string {
	lines := strings.Split(text, "\n")
	var result []string
	inCodeBlock := false
	codeLanguage := ""

	for i, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "```") {
			inCodeBlock = !inCodeBlock
			if inCodeBlock {
				parts := strings.Fields(strings.TrimSpace(line))
				if len(parts) > 1 {
					codeLanguage = parts[1]
				}
				result = append(result, "")
				result = append(result, m.code("  ┌─ Code: "+codeLanguage))
			} else {
				result = append(result, m.code("  └─"))
				result = append(result, "")
				codeLanguage = ""
			}
			continue
		}

		if inCodeBlock {
			result = append(result, m.renderCodeLine(line))
			continue
		}

		rendered := m.renderLine(line)
		result = append(result, rendered)

		if strings.HasPrefix(strings.TrimSpace(line), "#") && i < len(lines)-1 {
			if !strings.HasPrefix(strings.TrimSpace(lines[i+1]), "#") {
				result = append(result, "")
			}
		}
	}

	return strings.Join(result, "\n")
}

func (m *MarkdownRenderer) renderLine(line string) string {
	trimmed := strings.TrimSpace(line)

	if trimmed == "" {
		return ""
	}

	if strings.HasPrefix(trimmed, "### ") {
		return "  " + m.heading3("▍ "+strings.TrimPrefix(trimmed, "### "))
	}
	if strings.HasPrefix(trimmed, "## ") {
		return "  " + m.heading2("▊ "+strings.TrimPrefix(trimmed, "## "))
	}
	if strings.HasPrefix(trimmed, "# ") {
		return "  " + m.heading1("█ "+strings.TrimPrefix(trimmed, "# "))
	}

	if strings.HasPrefix(trimmed, ">") {
		quoted := strings.TrimPrefix(trimmed, ">")
		quoted = strings.TrimSpace(quoted)
		return "  " + m.quote("│ "+quoted)
	}

	if strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* ") {
		content := trimmed[2:]
		content = m.processInlineFormatting(content)
		return "  " + m.listBullet("•") + " " + content
	}

	re := regexp.MustCompile(`^(\d+)\.\s+(.*)`)
	if matches := re.FindStringSubmatch(trimmed); matches != nil {
		num := matches[1]
		content := m.processInlineFormatting(matches[2])
		return "  " + m.listBullet(num+".") + " " + content
	}

	processed := m.processInlineFormatting(line)
	if strings.TrimSpace(processed) != "" {
		return "  " + processed
	}
	return ""
}

func (m *MarkdownRenderer) processInlineFormatting(text string) string {
	re := regexp.MustCompile("`([^`]+)`")
	text = re.ReplaceAllStringFunc(text, func(match string) string {
		code := strings.Trim(match, "`")
		return m.code(code)
	})

	re = regexp.MustCompile(`\*\*([^\*]+)\*\*|__([^_]+)__`)
	text = re.ReplaceAllStringFunc(text, func(match string) string {
		content := strings.Trim(match, "*_")
		return m.bold(content)
	})

	re = regexp.MustCompile(`(?:^|[^*_])(\*|_)([^*_\s]+(?:\s+[^*_\s]+)*)(\*|_)(?:[^*_]|$)`)
	text = re.ReplaceAllStringFunc(text, func(match string) string {
		if len(match) > 2 {
			content := match[1 : len(match)-1]
			content = strings.Trim(content, "*_")
			if content != "" {
				return match[0:1] + m.italic(content) + match[len(match)-1:]
			}
		}
		return match
	})

	re = regexp.MustCompile(`\[([^\]]+)\]\(([^\)]+)\)`)
	text = re.ReplaceAllStringFunc(text, func(match string) string {
		matches := re.FindStringSubmatch(match)
		if len(matches) == 3 {
			return m.link(matches[1]) + m.quote(" ("+matches[2]+")")
		}
		return match
	})

	return text
}

func (m *MarkdownRenderer) renderCodeLine(line string) string {
	if strings.TrimSpace(line) == "" {
		return "  │"
	}
	return "  │ " + m.code(line)
}

func (m *MarkdownRenderer) StreamRender(text string, charDelay int) {
	rendered := m.Render(text)
	lines := strings.Split(rendered, "\n")

	for _, line := range lines {
		if strings.Contains(line, "█") || strings.Contains(line, "▊") || strings.Contains(line, "▍") {
			fmt.Println(line)
			continue
		}

		if strings.Contains(line, "┌─") || strings.Contains(line, "└─") || strings.HasPrefix(strings.TrimSpace(line), "│") {
			fmt.Println(line)
			continue
		}

		fmt.Println(line) 
	}
}
