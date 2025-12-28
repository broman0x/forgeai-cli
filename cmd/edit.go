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

	cwd, err := os.Getwd()
	var suggestedPath string

	if err == nil {
		hasCodeFiles := hasCodeFilesInDir(cwd)

		if hasCodeFiles {
			color.Cyan("  Current directory: %s", cwd)

			files := scanCodeFilesPreview(cwd)
			if len(files) > 0 {
				color.Green("  + Detected %d code file(s)", len(files))
				maxShow := 5
				for i, f := range files {
					if i >= maxShow {
						color.Cyan("     ... and %d more", len(files)-maxShow)
						break
					}
					color.Cyan("     - %s", f)
				}
			}
			fmt.Println()

			suggestedPath = cwd + string(filepath.Separator)
			color.Yellow("  Tip: Press Enter to use current directory, or type path")
		}
	}

	if suggestedPath != "" {
		fmt.Printf("%s [%s]: ", cPrompt("  File/Directory path"), cSubtle(suggestedPath))
	} else {
		fmt.Print(cPrompt("  File/Directory path: "))
	}

	if !scanner.Scan() {
		return
	}
	filePath := strings.TrimSpace(scanner.Text())

	if filePath == "" && suggestedPath != "" {
		filePath = suggestedPath
		color.Green("  → Using: %s", filePath)
	}

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
	isDir := false
	info, err := os.Stat(filePath)
	if err == nil && info.IsDir() {
		isDir = true
	} else if strings.HasSuffix(filePath, "/") || strings.HasSuffix(filePath, "\\") {
		isDir = true
	}

	instructionLower := strings.ToLower(instruction)
	projectType := detectProjectType(instructionLower)

	if isDir && projectType != "" {
		handleProjectCreation(prov, filePath, instruction, projectType, scanner)
		return
	}

	if isDir && isGeneralInstruction(instructionLower) {
		handleProjectAgentMode(prov, filePath, instruction, scanner)
		return
	}

	if isDir {
		filename := extractFilenameFromInstruction(instruction)

		if filename == "" {
			if strings.Contains(instructionLower, "buat") || strings.Contains(instructionLower, "create") || strings.Contains(instructionLower, "new") || strings.Contains(instructionLower, "add") {
				color.Yellow("  System detected you want to create a file.")
				fmt.Print("  Please enter the filename (e.g., main.py): ")
				if scanner.Scan() {
					inputName := strings.TrimSpace(scanner.Text())
					if inputName != "" {
						filename = inputName
					}
				}
			}
		}

		if filename != "" {
			filePath = filepath.Join(filePath, filename)
			color.Cyan("  Target file: %s", filename)
			fmt.Println()
		} else {
			color.Red("  Error: '%s' is a directory.", filePath)
			return
		}
	}

	content, err := os.ReadFile(filePath)
	fileExists := err == nil

	isCreateNewFile := !fileExists && (strings.Contains(instructionLower, "buatkan") ||
		strings.Contains(instructionLower, "create") ||
		strings.Contains(instructionLower, "buat file") ||
		strings.Contains(instructionLower, "new file") ||
		strings.Contains(instructionLower, "add"))

	if !fileExists && !isCreateNewFile {
		color.Yellow("  File not found: %s", filePath)

		fmt.Print("  Do you want to create this file? [y/N]: ")
		if scanner != nil && scanner.Scan() {
			response := strings.ToLower(strings.TrimSpace(scanner.Text()))
			if response == "y" || response == "yes" {
				isCreateNewFile = true
				content = []byte{}
			} else {
				return
			}
		} else {
			return
		}
	}

	dir := filepath.Dir(filePath)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			color.Red("  Error creating directory: %v", err)
			return
		}
	}

	ext := filepath.Ext(filePath)
	lang := detectLanguage(ext)

	spinner := ui.NewSpinner("Processing request")
	spinner.Start()

	prompt := fmt.Sprintf(`You are a world-class senior software engineer with deep expertise in %s and software architecture.

TASK: Execute this instruction: "%s"

IMPORTANT - UNDERSTAND THE INTENT:
- If instruction says "buatkan/create/add/implement NEW feature" → CREATE completely new code/content
- If instruction says "modify/change/fix/update EXISTING" → MODIFY the existing code
- If file is empty or minimal → User wants you to CREATE from scratch
- If instruction is about design/UI → Create visually stunning, modern, professional design

When CREATING NEW (e.g., landing pages, components, features):
1. START FROM SCRATCH - Don't just modify what's there
2. IMPLEMENT COMPLETE SOLUTION with all requested features
3. USE MODERN DESIGN:
   - Beautiful color schemes (gradients, modern palettes)
   - Responsive layouts (mobile-first)
   - Smooth animations and transitions
   - Premium aesthetics (glassmorphism, shadows, etc.)
   - Professional typography
4. INCLUDE ALL NECESSARY CODE (HTML + CSS + JS if needed)
5. Make it PRODUCTION-READY and VISUALLY IMPRESSIVE

When MODIFYING EXISTING:
1. PRESERVE original structure and intent
2. APPLY requested changes cleanly
3. IMPROVE code quality and best practices
4. FIX bugs and add error handling

CODE QUALITY STANDARDS:
1. BEST PRACTICES: Industry standards & design patterns
2. CLEAN CODE: Readable, maintainable, DRY principles
3. OPTIMIZATION: Performance-first approach
4. SECURITY: Input validation, prevent vulnerabilities
5. ERROR HANDLING: Proper error handling & edge cases
6. MODERN SYNTAX: Use latest language features
7. COMMENTS: Only for complex business logic

ARCHITECTURE PRINCIPLES:
- Single Responsibility Principle
- DRY (Don't Repeat Yourself)
- SOLID principles when applicable
- Clean separation of concerns
- Modular and reusable code

FOR WEB DEVELOPMENT (%s):
- Semantic HTML5 elements
- Modern CSS (Flexbox, Grid, CSS Variables, animations)
- Responsive design (mobile-first)
- Accessibility (ARIA labels, semantic markup)
- Performance optimization (lazy loading, efficient selectors)
- Beautiful, modern UI/UX design
- Professional color schemes and typography

OUTPUT REQUIREMENTS:
- Return ONLY the complete code
- NO explanations, NO markdown blocks, NO comments about changes
- Code must be immediately usable
- If creating new: Make it COMPLETE and IMPRESSIVE
- If modifying: Preserve working parts, improve requested areas

Current File: %s

Current Code (if any):
%s

EXECUTE THE INSTRUCTION ABOVE. If user wants something NEW, create it from scratch. If they want to modify, improve the existing code.`, lang, instruction, lang, filePath, string(content))

	newCode, err := prov.Send(prompt)
	spinner.Stop()

	if err != nil {
		color.Red("  Error: %v", err)
		return
	}

	newCode = cleanMarkdown(newCode)

	if isCreateNewFile || len(content) == 0 {
		cTitle := color.New(color.FgHiCyan, color.Bold).SprintFunc()
		cSubtle := color.New(color.FgHiBlack).SprintFunc()
		cCode := color.New(color.FgWhite).SprintFunc()

		fmt.Printf("\n  %s\n", cTitle("NEW FILE PREVIEW"))
		fmt.Println(cSubtle("  ───────────────────────────────────────────"))

		lines := strings.Split(newCode, "\n")
		maxLines := 30
		if len(lines) > maxLines {
			for i := 0; i < maxLines; i++ {
				fmt.Printf("  %s\n", cCode(lines[i]))
			}
			fmt.Printf("  %s\n", cSubtle(fmt.Sprintf("... and %d more lines", len(lines)-maxLines)))
		} else {
			for _, line := range lines {
				fmt.Printf("  %s\n", cCode(line))
			}
		}

		fmt.Println(cSubtle("  ───────────────────────────────────────────"))

		if confirm(scanner, fmt.Sprintf("\n  Create file '%s'? [y/N]", filepath.Base(filePath))) {
			if err := os.WriteFile(filePath, []byte(newCode), 0644); err != nil {
				color.Red("  Error writing file: %v", err)
				return
			}
			color.Green("  + File created successfully: %s", filePath)
		} else {
			color.Yellow("  File creation cancelled")
		}
		return
	}

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

func extractFilenameFromInstruction(instruction string) string {
	extensions := []string{
		".py", ".js", ".ts", ".jsx", ".tsx", ".go", ".java", ".cpp", ".c", ".cs",
		".rb", ".php", ".rs", ".kt", ".swift", ".html", ".css", ".scss", ".json",
		".xml", ".yaml", ".yml", ".md", ".txt", ".sh", ".sql", ".vue",
	}

	words := strings.Fields(instruction)

	for _, word := range words {
		word = strings.TrimRight(word, ".,;:!?")

		for _, ext := range extensions {
			if strings.HasSuffix(strings.ToLower(word), ext) {
				return word
			}
		}
	}

	return ""
}

func detectProjectType(instructionLower string) string {
	if strings.Contains(instructionLower, "landing page") ||
		strings.Contains(instructionLower, "website") ||
		strings.Contains(instructionLower, "web") && (strings.Contains(instructionLower, "html") || strings.Contains(instructionLower, "css")) {
		return "web-html"
	}

	if strings.Contains(instructionLower, "react") || strings.Contains(instructionLower, "jsx") {
		return "react"
	}

	if strings.Contains(instructionLower, "vue") {
		return "vue"
	}

	return ""
}

func handleProjectCreation(prov ai.Provider, dirPath, instruction, projectType string, scanner *bufio.Scanner) {
	color.Cyan("\n  SMART PROJECT DETECTION")
	color.Cyan("  Project Type: %s", strings.ToUpper(projectType))
	fmt.Println()

	if err := os.MkdirAll(dirPath, 0755); err != nil {
		color.Red("  Error creating directory: %v", err)
		return
	}

	existingFiles := scanExistingFiles(dirPath)

	color.Cyan("  Scanning directory...")
	if len(existingFiles) > 0 {
		color.Yellow("  Found %d existing files", len(existingFiles))
		for _, f := range existingFiles {
			fmt.Printf("    - %s\n", f)
		}
	} else {
		color.Green("  Directory is empty - creating fresh project")
	}
	fmt.Println()

	var filesToCreate []string
	switch projectType {
	case "web-html":
		filesToCreate = getWebProjectFiles(existingFiles, dirPath)
	case "react":
		filesToCreate = getReactProjectFiles(existingFiles, dirPath)
	case "vue":
		filesToCreate = getVueProjectFiles(existingFiles, dirPath)
	default:
		color.Yellow("  Project type not fully supported yet")
		return
	}

	var toCreate []string
	for _, f := range filesToCreate {
		if _, err := os.Stat(f); os.IsNotExist(err) {
			toCreate = append(toCreate, f)
		}
	}
	filesToCreate = toCreate

	if len(filesToCreate) == 0 {
		color.Green("  All project files already exist.")
		return
	}

	color.Cyan("  Proposed to create:")
	for _, f := range filesToCreate {
		fmt.Printf("    - %s\n", filepath.Base(f))
	}
	fmt.Println()

	if !confirm(scanner, "  Proceed with file creation? [Y/n]") {
		color.Yellow("  Cancelled.")
		return
	}

	for i, filePath := range filesToCreate {
		ext := filepath.Ext(filePath)
		filename := filepath.Base(filePath)

		color.Cyan("\n  [%d/%d] Creating %s...", i+1, len(filesToCreate), filename)

		fileInstruction := generateFileInstruction(instruction, filename, projectType)

		spinner := ui.NewSpinner("Generating code")
		spinner.Start()

		lang := detectLanguage(ext)

		var prompt string
		if strings.HasSuffix(filename, ".html") {
			prompt = fmt.Sprintf(`You are a world-class web developer.

Create ONLY the HTML file (%s) for this project.

PROJECT: %s
FILE: %s
TYPE: %s

CRITICAL REQUIREMENTS:
- Return ONLY HTML code for THIS file
- NO CSS code (that goes in style.css)
- NO JavaScript code (that goes in script.js)
- Include proper <link> to style.css
- Include proper <script> to script.js
- Use semantic HTML5 elements
- Modern, clean structure
- Responsive meta tags

IMPORTANT: Return ONLY the HTML code. NO explanations, NO markdown, NO other files.
DO NOT START WITH "Here is the code".

%s

OUTPUT (ONLY CODE):`, filename, instruction, filename, projectType, fileInstruction)

		} else if strings.HasSuffix(filename, ".css") {
			prompt = fmt.Sprintf(`You are a world-class CSS designer.

Create ONLY the CSS file (%s) for this project.

PROJECT: %s
FILE: %s
TYPE: %s

CRITICAL REQUIREMENTS:
- Return ONLY CSS code for THIS file
- NO HTML code
- NO JavaScript code
- Use CSS variables for colors
- Modern layouts (flexbox/grid)
- Responsive design with media queries
- Beautiful color schemes and animations
- Clean, organized code

IMPORTANT: Return ONLY the CSS code. NO explanations, NO markdown, NO other files.
DO NOT START WITH "Here is the code".

%s

OUTPUT (ONLY CODE):`, filename, instruction, filename, projectType, fileInstruction)

		} else if strings.HasSuffix(filename, ".js") {
			prompt = fmt.Sprintf(`You are a world-class JavaScript developer.

Create ONLY the JavaScript file (%s) for this project.

PROJECT: %s
FILE: %s
TYPE: %s

CRITICAL REQUIREMENTS:
- Return ONLY JavaScript code for THIS file
- NO HTML code
- NO CSS code
- Use modern ES6+ syntax
- Clean event handling
- Proper DOM manipulation
- No external dependencies (vanilla JS)

IMPORTANT: Return ONLY the JavaScript code. NO explanations, NO markdown, NO other files.
DO NOT START WITH "Here is the code".

%s

OUTPUT (ONLY CODE):`, filename, instruction, filename, projectType, fileInstruction)

		} else {
			prompt = fmt.Sprintf(`You are a world-class software engineer.

Create ONLY THIS file: %s

PROJECT: %s
FILE TYPE: %s
REQUIREMENTS: %s

CRITICAL: Return ONLY the code for THIS SINGLE file. NO explanations, NO markdown blocks, NO other files.
DO NOT START WITH "Here is the code".

OUTPUT (ONLY CODE):`, filename, instruction, lang, fileInstruction)
		}

		code, err := prov.Send(prompt)
		spinner.Stop()

		if err != nil {
			color.Red("  ! Error generating %s: %v", filename, err)
			continue
		}

		code = cleanMarkdown(code)

		if err := os.WriteFile(filePath, []byte(code), 0644); err != nil {
			color.Red("  ! Error writing %s: %v", filename, err)
			continue
		}

		color.Green("  + Created %s", filename)
	}

	fmt.Println()
	color.Green("  * Project created successfully!")
	color.Cyan("  Location: %s", dirPath)
}

func isGeneralInstruction(instruction string) bool {
	keywords := []string{
		"fix", "perbaiki", "repair", "refactor", "optimize", "improve", "enhance",
		"check", "analisa", "analyze", "scan", "audit", "review", "debug",
		"clean", "tidy", "format", "update", "upgrade", "modernize",

		"design", "desain", "theme", "tema", "style", "tampilan", "ui", "ux", "layout",
		"color", "warna", "font", "typografi", "css", "bootstrap", "tailwind",

		"change", "ubah", "ganti", "modify", "modifikasi",
		"perbagus", "beautify", "cantik", "bagus", "keren",
		"remove", "hapus", "delete", "hilangkan", "bersihkan",
	}

	instruction = strings.ToLower(instruction)
	for _, kw := range keywords {
		if strings.Contains(instruction, kw) {
			return true
		}
	}

	if strings.Contains(instruction, "this") && len(strings.Fields(instruction)) < 5 {
		return true
	}

	return false
}

func handleProjectAgentMode(prov ai.Provider, dirPath, instruction string, scanner *bufio.Scanner) {
	cTitle := color.New(color.FgHiMagenta, color.Bold).SprintFunc()
	cSubtle := color.New(color.FgHiBlack).SprintFunc()

	fmt.Println()
	fmt.Println(cTitle("  PROJECT AGENT MODE"))
	fmt.Println(cSubtle("  ───────────────────────────────────────────"))
	color.Cyan("  Instruction: %s", instruction)
	color.Cyan("  Target: %s", dirPath)
	fmt.Println()

	spinner := ui.NewSpinner("Scanning project files")
	spinner.Start()

	files := scanExistingFiles(dirPath)
	spinner.Stop()

	if len(files) == 0 {
		color.Yellow("  No code files found in this directory to process.")
		return
	}

	fmt.Printf("  Found %d files to analyze:\n", len(files))
	for _, f := range files {
		fmt.Printf("   - %s\n", f)
	}
	fmt.Println()

	if !confirm(scanner, "  Proceed with analysis and improvements? [Y/n]") {
		color.Yellow("  Agent stopped.")
		return
	}

	processedCount := 0
	changedCount := 0

	for i, relPath := range files {
		fullPath := filepath.Join(dirPath, relPath)
		filename := filepath.Base(fullPath)

		fmt.Println(cSubtle("  ───────────────────────────────────────────"))
		color.Cyan("  [%d/%d] Analyzing %s...", i+1, len(files), filename)

		content, err := os.ReadFile(fullPath)
		if err != nil {
			color.Red("  Error reading %s: %v", filename, err)
			continue
		}

		if len(content) == 0 {
			color.Yellow("  Skipping empty file.")
			continue
		}

		spinner := ui.NewSpinner("  Agent is thinking")
		spinner.Start()

		ext := filepath.Ext(filename)
		lang := detectLanguage(ext)

		prompt := fmt.Sprintf(`You are an expert AI software agent.
USER INSTRUCTION: "%s"

CONTEXT: You are processing file "%s".
FILE TYPE: %s
LANGUAGE: %s

CRITICAL RULES:
1. FOCUS ONLY ON THIS FILE (%s). Do NOT generate code for other files.
2. Analyze the code based on the user instruction.
3. If the instruction applies to this file, MODIFY the code.
4. If the instruction does NOT apply (e.g. instruction is 'fix css' but this is 'script.js'), return the code EXACTLY AS IS.
5. DO NOT START WITH "Here is the code" or "I have fixed it".
6. RETURN ONLY THE CODE. NO MARKDOWN BLOCK START/END.

CODE CONTENT:
%s

OUTPUT (ONLY CODE):`, instruction, filename, ext, lang, filename, string(content))

		newCode, err := prov.Send(prompt)
		spinner.Stop()

		if err != nil {
			color.Red("  Agent error: %v", err)
			continue
		}

		newCode = cleanMarkdown(newCode)

		if strings.Contains(newCode, "/* style.css */") || strings.Contains(newCode, "<!-- index.html -->") {
			color.Yellow("  ! Warning: AI output might contain multiple files. Attempting to extract relevant code.")
		}

		diff := difflib.UnifiedDiff{
			A:        difflib.SplitLines(string(content)),
			B:        difflib.SplitLines(newCode),
			FromFile: "Original",
			ToFile:   "Agent Modified",
			Context:  3,
		}
		diffText, _ := difflib.GetUnifiedDiffString(diff)

		if strings.TrimSpace(diffText) == "" {
			color.Green("  + No changes needed.")
		} else {
			color.Yellow("  * Proposed changes for %s:", filename)
			printDiff(diffText)

			if confirm(scanner, fmt.Sprintf("  Apply changes to %s? [y/N]", filename)) {
				if err := os.WriteFile(fullPath, []byte(newCode), 0644); err != nil {
					color.Red("  Error saving: %v", err)
				} else {
					color.Green("  + Changes saved.")
					changedCount++
				}
			} else {
				color.Yellow("  Skipped.")
			}
		}
		processedCount++
	}

	fmt.Println()
	fmt.Println(cTitle("  AGENT SUMMARY"))
	fmt.Println(cSubtle("  ───────────────────────────────────────────"))
	color.Green("  Processed: %d files", processedCount)
	color.Green("  Modified:  %d files", changedCount)
	fmt.Println()
}

func scanExistingFiles(dirPath string) []string {
	var files []string
	codeExts := map[string]bool{
		".html": true, ".css": true, ".js": true, ".jsx": true, ".tsx": true,
		".vue": true, ".ts": true, ".json": true, ".py": true, ".go": true,
	}

	filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}

		ext := filepath.Ext(path)
		if codeExts[ext] {
			relPath, _ := filepath.Rel(dirPath, path)
			files = append(files, relPath)
		}

		return nil
	})

	return files
}

func getWebProjectFiles(existing []string, dirPath string) []string {
	required := []string{"index.html", "style.css", "script.js"}
	var toCreate []string

	existingMap := make(map[string]bool)
	for _, f := range existing {
		existingMap[strings.ToLower(filepath.Base(f))] = true
	}

	for _, req := range required {
		if !existingMap[req] {
			toCreate = append(toCreate, filepath.Join(dirPath, req))
		}
	}

	return toCreate
}

func getReactProjectFiles(existing []string, dirPath string) []string {
	required := []string{"App.jsx", "index.html", "style.css"}
	var toCreate []string

	existingMap := make(map[string]bool)
	for _, f := range existing {
		existingMap[strings.ToLower(filepath.Base(f))] = true
	}

	for _, req := range required {
		if !existingMap[strings.ToLower(req)] {
			toCreate = append(toCreate, filepath.Join(dirPath, req))
		}
	}

	return toCreate
}

func getVueProjectFiles(existing []string, dirPath string) []string {
	required := []string{"App.vue", "index.html", "main.js"}
	var toCreate []string

	existingMap := make(map[string]bool)
	for _, f := range existing {
		existingMap[strings.ToLower(filepath.Base(f))] = true
	}

	for _, req := range required {
		if !existingMap[strings.ToLower(req)] {
			toCreate = append(toCreate, filepath.Join(dirPath, req))
		}
	}

	return toCreate
}

func generateFileInstruction(baseInstruction, filename, projectType string) string {
	filenameLower := strings.ToLower(filename)

	if strings.HasSuffix(filenameLower, ".html") {
		return fmt.Sprintf("%s - HTML structure with semantic markup", baseInstruction)
	}
	if strings.HasSuffix(filenameLower, ".css") {
		return fmt.Sprintf("%s - Modern CSS with beautiful design, responsive layout", baseInstruction)
	}
	if strings.HasSuffix(filenameLower, ".js") {
		return fmt.Sprintf("%s - JavaScript for interactivity and dynamic features", baseInstruction)
	}
	if strings.HasSuffix(filenameLower, ".jsx") {
		return fmt.Sprintf("%s - React component with hooks and modern patterns", baseInstruction)
	}
	if strings.HasSuffix(filenameLower, ".vue") {
		return fmt.Sprintf("%s - Vue component with composition API", baseInstruction)
	}

	return baseInstruction
}

func hasCodeFilesInDir(dirPath string) bool {
	codeExts := map[string]bool{
		".go": true, ".js": true, ".ts": true, ".py": true, ".jsx": true, ".tsx": true,
		".html": true, ".css": true, ".scss": true, ".java": true, ".cpp": true, ".c": true,
		".cs": true, ".rb": true, ".php": true, ".rs": true, ".kt": true, ".swift": true,
		".vue": true, ".json": true, ".yaml": true, ".yml": true, ".sql": true,
	}

	hasFiles := false
	filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || hasFiles {
			return filepath.SkipDir
		}

		if info.IsDir() {
			name := info.Name()
			if name == "node_modules" || name == ".git" || name == "vendor" ||
				name == "target" || name == "build" || name == "dist" || name == ".idea" {
				return filepath.SkipDir
			}
			return nil
		}

		ext := filepath.Ext(path)
		if codeExts[ext] {
			hasFiles = true
			return filepath.SkipDir
		}

		return nil
	})

	return hasFiles
}

func scanCodeFilesPreview(dirPath string) []string {
	var files []string
	codeExts := map[string]bool{
		".go": true, ".js": true, ".ts": true, ".py": true, ".jsx": true, ".tsx": true,
		".html": true, ".css": true, ".scss": true, ".java": true, ".cpp": true, ".c": true,
		".cs": true, ".rb": true, ".php": true, ".rs": true, ".kt": true, ".swift": true,
		".vue": true, ".json": true, ".yaml": true, ".yml": true, ".sql": true,
	}

	maxFiles := 10

	filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || len(files) >= maxFiles {
			return filepath.SkipDir
		}

		if info.IsDir() {
			name := info.Name()
			if name == "node_modules" || name == ".git" || name == "vendor" ||
				name == "target" || name == "build" || name == "dist" || name == ".idea" {
				return filepath.SkipDir
			}
			return nil
		}

		ext := filepath.Ext(path)
		if codeExts[ext] {
			relPath, _ := filepath.Rel(dirPath, path)
			files = append(files, relPath)
		}

		return nil
	})

	return files
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
