package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/broman0x/forgeai-cli/internal/ai"
	"github.com/broman0x/forgeai-cli/internal/config"
	"github.com/broman0x/forgeai-cli/internal/lang"
	"github.com/broman0x/forgeai-cli/internal/ui"
	"github.com/fatih/color"
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
)

const Version = "1.0.1"

var (
	cfgFile         string
	noBanner        bool
	doInstall       bool
	doUninstall     bool
	showVersion     bool
	currentProvider ai.Provider
)

var rootCmd = &cobra.Command{
	Use:   "forgeai",
	Short: "Professional AI CLI",
	Args:  cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if showVersion {
			fmt.Printf("ForgeAI CLI v%s\n", Version)
			fmt.Println("AI-Powered Development Assistant")
			fmt.Println("by bromanprjkt")
			return nil
		}
		if doInstall {
			return runSelfInstall()
		}
		if doUninstall {
			return runSelfUninstall()
		}
		if len(args) > 0 {
			return runOneShot(strings.Join(args, " "))
		}
		return runMainMenu()
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file")
	rootCmd.PersistentFlags().BoolVar(&noBanner, "no-banner", false, "disable banner")
	rootCmd.Flags().BoolVarP(&showVersion, "version", "v", false, "show version information")
	rootCmd.Flags().BoolVar(&doInstall, "install", false, "install forge to PATH")
	rootCmd.Flags().BoolVar(&doUninstall, "uninstall", false, "uninstall forge from PATH")
	cobra.MousetrapHelpText = ""
}

func runMainMenu() error {
	config.ResetCache()
	cfg := config.Load()

	if cfg.FirstRun {
		if err := runFirstTimeSetup(); err != nil {
			return err
		}
		config.ResetCache()
		cfg = config.Load()
	} else {
	}

	lang.SetLanguage(cfg.Language)

	exePath, _ := os.Executable()
	installDir := filepath.Join(os.Getenv("LocalAppData"), "ForgeAI")
	installedPath := filepath.Join(installDir, "forge.exe")

	if runtime.GOOS == "windows" && exePath != installedPath {
		fmt.Println()
		color.Yellow("  ⚠ ForgeAI belum ter-install ke sistem")
		color.Cyan("  Ingin install sekarang agar bisa dipanggil dari mana saja?")
		fmt.Println()
		fmt.Print("  Install ke PATH? [Y/n]: ")

		scanner := bufio.NewScanner(os.Stdin)
		if scanner.Scan() {
			response := strings.ToLower(strings.TrimSpace(scanner.Text()))
			if response == "" || response == "y" || response == "yes" {
				if err := runSelfInstall(); err != nil {
					color.Red("  Install failed: %v", err)
				}
				return nil
			}
		}
	}

	ui.ShowStartupBanner()

	var err error
	currentProvider, err = ai.NewProvider()

	if err != nil {
		return runSetupWizard(err)
	}

	scanner := bufio.NewScanner(os.Stdin)
	cActive := color.New(color.BgCyan, color.FgBlack, color.Bold).SprintFunc()

	for {
		fmt.Printf("  %s %s\n\n", cActive(" "+lang.T("active_brain")+" "), currentProvider.Name())

		fmt.Printf("  [ 1 ]   %s %s\n", lang.T("chat_mode"), lang.T("chat_mode_desc"))
		fmt.Printf("  [ 2 ]   %s %s\n", lang.T("code_review"), lang.T("review_desc"))
		fmt.Printf("  [ 3 ]   %s %s\n", lang.T("code_editor"), lang.T("editor_desc"))
		fmt.Printf("  [ 4 ]   %s %s\n", lang.T("switch_model"), lang.T("switch_desc"))
		fmt.Printf("  [ 5 ]   %s %s\n", lang.T("system_info"), lang.T("info_desc"))
		fmt.Printf("  [ 6 ]   %s %s\n", lang.T("uninstall"), lang.T("uninstall_desc"))
		fmt.Printf("  [ 7 ]   %s\n", "Change API Key")
		fmt.Println()
		fmt.Printf("  [ 0 ]   %s %s\n", lang.T("exit"), lang.T("exit_desc"))

		fmt.Printf("\n  %s > ", lang.T("select_command"))
		if !scanner.Scan() {
			return nil
		}

		input := strings.TrimSpace(scanner.Text())

		switch input {
		case "1":
			startChatMode(scanner)
		case "2":
			StartReviewModeInteractive(scanner, currentProvider)
		case "3":
			StartEditModeInteractive(scanner, currentProvider)
		case "4":
			handleSwitchModel(scanner)
		case "5":
			fmt.Print("\033[H\033[2J")
			ui.ShowStartupBanner()
		case "6":
			runUninstaller()
			return nil
		case "7":
			handleChangeAPIKey(scanner)
		case "0":
			fmt.Printf("\n  %s\n\n", lang.T("shutting_down"))
			return nil
		default:
			fmt.Print("\033[H\033[2J")
			ui.ShowStartupBanner()
		}
	}
	return nil
}

func runFirstTimeSetup() error {
	fmt.Print("\033[H\033[2J")

	cTitle := color.New(color.FgHiCyan, color.Bold).SprintFunc()
	cBorder := color.New(color.FgHiBlack).SprintFunc()
	cHighlight := color.New(color.FgHiYellow, color.Bold).SprintFunc()
	cWhite := color.New(color.FgWhite).SprintFunc()

	fmt.Println()
	fmt.Println()
	time.Sleep(200 * time.Millisecond)

	fmt.Println(cTitle("      ╔══════════════════════════════════════════╗"))
	time.Sleep(100 * time.Millisecond)
	fmt.Println(cTitle("      ║                                          ║"))
	fmt.Println(cTitle("      ║      Welcome to ForgeAI CLI v" + Version + "       ║"))
	time.Sleep(100 * time.Millisecond)
	fmt.Println(cTitle("      ║    AI-Powered Development Assistant      ║"))
	fmt.Println(cTitle("      ║                                          ║"))
	time.Sleep(100 * time.Millisecond)
	fmt.Println(cTitle("      ║          Developed by bromanprjkt        ║"))
	fmt.Println(cTitle("      ║                                          ║"))
	fmt.Println(cTitle("      ╚══════════════════════════════════════════╝"))
	time.Sleep(200 * time.Millisecond)

	fmt.Println()
	fmt.Println()
	fmt.Println(cHighlight("      ⚡ FIRST TIME SETUP"))
	fmt.Println(cBorder("      ────────────────────────────────────────────"))
	fmt.Println()
	time.Sleep(200 * time.Millisecond)

	fmt.Println(cWhite("      Choose your preferred language:"))
	fmt.Println()
	time.Sleep(100 * time.Millisecond)

	fmt.Println("         " + cBorder("[") + color.New(color.FgCyan, color.Bold).Sprint(" 1 ") + cBorder("]") + "  🇬🇧 English")
	time.Sleep(100 * time.Millisecond)
	fmt.Println("         " + cBorder("[") + color.New(color.FgCyan, color.Bold).Sprint(" 2 ") + cBorder("]") + "  🇮🇩 Bahasa Indonesia")
	fmt.Println()
	time.Sleep(100 * time.Millisecond)

	fmt.Print(cHighlight("      ➤ Select [1-2]: "))

	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		return fmt.Errorf("setup cancelled")
	}

	choice := strings.TrimSpace(scanner.Text())
	language := "en"

	if choice == "2" {
		language = "id"
	}

	if err := config.SaveLanguage(language); err != nil {
		return err
	}

	if err := config.SaveFirstRun(false); err != nil {
		return err
	}

	lang.SetLanguage(language)

	fmt.Println()
	fmt.Println()

	for i := 0; i < 3; i++ {
		fmt.Print("\r      ")
		time.Sleep(200 * time.Millisecond)
		if language == "id" {
			fmt.Print(color.GreenString("✓ Menyimpan pengaturan..."))
		} else {
			fmt.Print(color.GreenString("✓ Saving preferences..."))
		}
	}

	fmt.Println()
	fmt.Println()

	if language == "id" {
		color.Green("      ✓ Pengaturan berhasil! Memulai ForgeAI...")
	} else {
		color.Green("      ✓ Setup successful! Starting ForgeAI...")
	}
	fmt.Println()
	time.Sleep(1200 * time.Millisecond)

	return nil
}

func runSetupWizard(originalErr error) error {
	color.Yellow("\n  SETUP REQUIRED")
	fmt.Println("  ForgeAI needs an AI provider to function.")
	fmt.Printf("  Error: %v\n", originalErr)
	fmt.Println()

	fmt.Println("  Select AI Provider:")
	fmt.Println("  1. Ollama (Free, Local, Offline)")
	fmt.Println("  2. Google Gemini (Free tier available)")
	fmt.Println("  3. OpenAI ChatGPT (Paid)")
	fmt.Println("  4. Anthropic Claude (Paid)")
	fmt.Println("  0. Exit")

	fmt.Print("\n  Select > ")
	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		return nil
	}

	choice := strings.TrimSpace(scanner.Text())

	switch choice {
	case "1":
		color.Yellow("\n  Make sure Ollama is running on port 11434.")
		fmt.Println("  Install: https://ollama.ai")
		fmt.Println("  Press Enter to check connection...")
		scanner.Scan()
		fmt.Print("\033[H\033[2J")
		return runMainMenu()

	case "2":
		fmt.Println("\n  Get your key: https://aistudio.google.com/app/apikey")
		fmt.Print("  Paste Gemini API Key: ")
		if scanner.Scan() {
			apiKey := strings.TrimSpace(scanner.Text())
			if apiKey == "" {
				color.Red("  Empty key provided.")
				return fmt.Errorf("setup aborted")
			}

			err := config.SaveAPIKey("GEMINI_API_KEY", apiKey)
			if err != nil {
				color.Red("  Failed to save key: %v", err)
				return err
			}

			color.Green("  Configuration saved! Restarting...")
			time.Sleep(1 * time.Second)
			godotenv.Load()
			fmt.Print("\033[H\033[2J")
			return runMainMenu()
		}

	case "3":
		fmt.Println("\n  Get your key: https://platform.openai.com/api-keys")
		fmt.Print("  Paste OpenAI API Key: ")
		if scanner.Scan() {
			apiKey := strings.TrimSpace(scanner.Text())
			if apiKey == "" {
				color.Red("  Empty key provided.")
				return fmt.Errorf("setup aborted")
			}

			err := config.SaveAPIKey("OPENAI_API_KEY", apiKey)
			if err != nil {
				color.Red("  Failed to save key: %v", err)
				return err
			}

			color.Green("  Configuration saved! Restarting...")
			time.Sleep(1 * time.Second)
			godotenv.Load()
			fmt.Print("\033[H\033[2J")
			return runMainMenu()
		}

	case "4":
		fmt.Println("\n  Get your key: https://console.anthropic.com/settings/keys")
		fmt.Print("  Paste Claude API Key: ")
		if scanner.Scan() {
			apiKey := strings.TrimSpace(scanner.Text())
			if apiKey == "" {
				color.Red("  Empty key provided.")
				return fmt.Errorf("setup aborted")
			}

			err := config.SaveAPIKey("ANTHROPIC_API_KEY", apiKey)
			if err != nil {
				color.Red("  Failed to save key: %v", err)
				return err
			}

			color.Green("  Configuration saved! Restarting...")
			time.Sleep(1 * time.Second)
			godotenv.Load()
			fmt.Print("\033[H\033[2J")
			return runMainMenu()
		}
	}

	return nil
}

func startChatMode(scanner *bufio.Scanner) {
	fmt.Print("\033[H\033[2J")
	ui.ShowStartupBanner()

	cHeader := color.New(color.FgHiCyan, color.Bold).SprintFunc()
	cSubtle := color.New(color.FgHiBlack).SprintFunc()
	fmt.Println()
	fmt.Printf("  %s\n", cHeader("━━━ CHAT MODE ━━━"))
	fmt.Printf("  %s\n", cSubtle("Type 'exit' to return • 'clear' to reset screen"))
	fmt.Println()

	cAI := color.New(color.FgHiCyan, color.Bold).SprintFunc()
	cPrompt := color.New(color.FgCyan).SprintFunc()

	mdRenderer := ui.NewMarkdownRenderer()

	var systemPrompt string
	if lang.GetLanguage() == "id" {
		systemPrompt = `Anda adalah Forge AI, asisten coding profesional yang dikembangkan oleh bromanprjkt. 
Ketika ditanya siapa Anda, jawab: "Saya adalah Forge AI, asisten coding cerdas Anda yang dirancang untuk membantu code review, editing, dan tugas pengembangan."
Ketika ditanya siapa yang membuat Anda, jawab: "Saya dikembangkan oleh bromanprjkt, developer yang passionate tentang AI-powered development tools."
Selalu profesional, membantu, dan ringkas dalam respons Anda. Gunakan Bahasa Indonesia untuk semua respons.`
	} else {
		systemPrompt = `You are Forge AI, a professional coding assistant developed by bromanprjkt. 
When asked who you are, respond: "I am Forge AI, your intelligent coding assistant designed to help with code review, editing, and development tasks."
When asked who created you, respond: "I was developed by bromanprjkt, a skilled developer passionate about AI-powered development tools."
Always be professional, helpful, and concise in your responses.`
	}

	currentProvider.Send(systemPrompt)

	for {
		fmt.Printf("\n  %s ", cPrompt("You >"))

		if !scanner.Scan() {
			break
		}
		input := strings.TrimSpace(scanner.Text())

		if input == "" {
			continue
		}
		if input == "back" || input == "exit" {
			fmt.Print("\033[H\033[2J")
			ui.ShowStartupBanner()
			break
		}
		if input == "clear" || input == "cls" {
			fmt.Print("\033[H\033[2J")
			ui.ShowStartupBanner()
			fmt.Println()
			fmt.Printf("  %s\n", cHeader("━━━ CHAT MODE ━━━"))
			fmt.Printf("  %s\n", cSubtle("Type 'exit' to return • 'clear' to reset screen"))
			fmt.Println()
			continue
		}

		spinner := ui.NewSpinner("Thinking")
		spinner.Start()
		resp, err := currentProvider.Send(input)
		spinner.Stop()

		if err != nil {
			color.Red("  Error: %v\n", err)
		} else {
			fmt.Printf("\n  %s\n\n", cAI("Forge AI >"))

			rendered := mdRenderer.Render(resp)
			fmt.Println(rendered)

			fmt.Println()
		}
	}
}

func getAndValidateAPIKey(scanner *bufio.Scanner, keyName, keyURL, providerType, modelName string) bool {
	maxRetries := 3
	for attempt := 1; attempt <= maxRetries; attempt++ {
		if attempt == 1 {
			fmt.Printf("\n  Get your key: %s\n", keyURL)
		} else {
			color.Yellow("\n  Attempt %d of %d", attempt, maxRetries)
		}

		fmt.Print("  Paste API Key: ")
		scanner.Scan()
		apiKey := strings.TrimSpace(scanner.Text())

		if apiKey == "" {
			color.Red("  Empty key provided.")
			if attempt < maxRetries {
				fmt.Print("  Try again? [Y/n]: ")
				scanner.Scan()
				response := strings.ToLower(strings.TrimSpace(scanner.Text()))
				if response == "n" || response == "no" {
					return false
				}
				continue
			}
			return false
		}

		if err := config.SaveAPIKey(keyName, apiKey); err != nil {
			color.Red("  Failed to save key: %v", err)
			return false
		}
		godotenv.Load()

		color.Cyan("\n  Validating API key...")
		spinner := ui.NewSpinner("Testing connection")
		spinner.Start()

		testProvider, err := ai.CreateProvider(providerType, modelName)
		if err != nil {
			spinner.Stop()
			color.Red("  ✗ Invalid API key: %v", err)

			if attempt < maxRetries {
				fmt.Print("\n  Try again? [Y/n]: ")
				scanner.Scan()
				response := strings.ToLower(strings.TrimSpace(scanner.Text()))
				if response == "n" || response == "no" {
					return false
				}
				continue
			}
			return false
		}

		_, err = testProvider.Send("Hi")
		spinner.Stop()

		if err != nil {
			errMsg := strings.ToLower(err.Error())
			if strings.Contains(errMsg, "api key") || strings.Contains(errMsg, "unauthorized") ||
				strings.Contains(errMsg, "authentication") || strings.Contains(errMsg, "invalid") ||
				strings.Contains(errMsg, "401") || strings.Contains(errMsg, "403") {
				color.Red("  ✗ API Key validation failed: %v", err)

				if attempt < maxRetries {
					fmt.Print("\n  Try again? [Y/n]: ")
					scanner.Scan()
					response := strings.ToLower(strings.TrimSpace(scanner.Text()))
					if response == "n" || response == "no" {
						return false
					}
					continue
				}
				return false
			}
		}

		color.Green("  ✓ API Key validated successfully!")
		time.Sleep(500 * time.Millisecond)
		return true
	}

	color.Red("\n  Maximum retry attempts reached. Please check your API key.")
	return false
}

func handleChangeAPIKey(scanner *bufio.Scanner) {
	ui.PrintHeader("CHANGE API KEY")
	fmt.Println("  Select provider to change API key:")
	fmt.Println("  1. Google Gemini")
	fmt.Println("  2. OpenAI ChatGPT")
	fmt.Println("  3. Anthropic Claude")
	fmt.Println("  0. Back")
	fmt.Print("\n  Selection: ")
	scanner.Scan()
	choice := strings.TrimSpace(scanner.Text())

	switch choice {
	case "1":
		color.Cyan("\n  Changing Gemini API Key...")
		if getAndValidateAPIKey(scanner, "GEMINI_API_KEY", "https://aistudio.google.com/app/apikey", "gemini", "gemini-2.5-flash") {
			color.Green("\n  ✓ Gemini API Key updated successfully!")
			time.Sleep(1 * time.Second)
		}
	case "2":
		color.Cyan("\n  Changing OpenAI API Key...")
		if getAndValidateAPIKey(scanner, "OPENAI_API_KEY", "https://platform.openai.com/api-keys", "openai", "gpt-3.5-turbo") {
			color.Green("\n  ✓ OpenAI API Key updated successfully!")
			time.Sleep(1 * time.Second)
		}
	case "3":
		color.Cyan("\n  Changing Claude API Key...")
		if getAndValidateAPIKey(scanner, "ANTHROPIC_API_KEY", "https://console.anthropic.com/settings/keys", "claude", "claude-3-haiku-20240307") {
			color.Green("\n  ✓ Claude API Key updated successfully!")
			time.Sleep(1 * time.Second)
		}
	case "0":
		// Back to menu
		return
	default:
		color.Yellow("  Invalid selection")
		time.Sleep(1 * time.Second)
	}

	fmt.Print("\033[H\033[2J")
	ui.ShowStartupBanner()
}

func handleSwitchModel(scanner *bufio.Scanner) {
	ui.PrintHeader("SWITCH AI PROVIDER")
	fmt.Println("  1. Gemini 2.5 Flash")
	fmt.Println("  2. Gemini Pro")
	fmt.Println("  3. OpenAI GPT-3.5 Turbo")
	fmt.Println("  4. OpenAI GPT-4")
	fmt.Println("  5. Claude 3 Haiku")
	fmt.Println("  6. Claude 3 Sonnet")
	fmt.Println("  7. Ollama (Local)")
	fmt.Print("\n  Selection: ")
	scanner.Scan()
	choice := strings.TrimSpace(scanner.Text())

	var p ai.Provider
	var err error
	var selectedModel string
	var providerType string

	switch choice {
	case "1":
		providerType = "gemini"
		selectedModel = "gemini-2.5-flash"
		if os.Getenv("GEMINI_API_KEY") == "" {
			if !getAndValidateAPIKey(scanner, "GEMINI_API_KEY", "https://aistudio.google.com/app/apikey", providerType, selectedModel) {
				return
			}
		}
		p, err = ai.CreateProvider("gemini", selectedModel)
	case "2":
		providerType = "gemini"
		selectedModel = "gemini-pro"
		if os.Getenv("GEMINI_API_KEY") == "" {
			if !getAndValidateAPIKey(scanner, "GEMINI_API_KEY", "https://aistudio.google.com/app/apikey", providerType, selectedModel) {
				return
			}
		}
		p, err = ai.CreateProvider("gemini", selectedModel)
	case "3":
		providerType = "openai"
		selectedModel = "gpt-3.5-turbo"
		if os.Getenv("OPENAI_API_KEY") == "" {
			if !getAndValidateAPIKey(scanner, "OPENAI_API_KEY", "https://platform.openai.com/api-keys", providerType, selectedModel) {
				return
			}
		}
		p, err = ai.CreateProvider("openai", selectedModel)
	case "4":
		providerType = "openai"
		selectedModel = "gpt-4"
		if os.Getenv("OPENAI_API_KEY") == "" {
			if !getAndValidateAPIKey(scanner, "OPENAI_API_KEY", "https://platform.openai.com/api-keys", providerType, selectedModel) {
				return
			}
		}
		p, err = ai.CreateProvider("openai", selectedModel)
	case "5":
		providerType = "claude"
		selectedModel = "claude-3-haiku-20240307"
		if os.Getenv("ANTHROPIC_API_KEY") == "" {
			if !getAndValidateAPIKey(scanner, "ANTHROPIC_API_KEY", "https://console.anthropic.com/settings/keys", providerType, selectedModel) {
				return
			}
		}
		p, err = ai.CreateProvider("claude", selectedModel)
	case "6":
		providerType = "claude"
		selectedModel = "claude-3-sonnet-20240229"
		if os.Getenv("ANTHROPIC_API_KEY") == "" {
			if !getAndValidateAPIKey(scanner, "ANTHROPIC_API_KEY", "https://console.anthropic.com/settings/keys", providerType, selectedModel) {
				return
			}
		}
		p, err = ai.CreateProvider("claude", selectedModel)
	case "7":
		providerType = "ollama"
		ui.PrintHeader("ENTER OLLAMA MODEL")
		fmt.Print("  Model name: ")
		scanner.Scan()
		selectedModel = strings.TrimSpace(scanner.Text())
		if selectedModel == "" {
			selectedModel = "llama3"
		}
		p, err = ai.CreateProvider("ollama", selectedModel)
	default:
		return
	}

	if err == nil {
		currentProvider = p

		if err := config.SaveLastModel(providerType, selectedModel); err != nil {
			color.Red("  Warning: Could not save model preference: %v\n", err)
		}

		fmt.Print("\033[H\033[2J")
		ui.ShowStartupBanner()
		color.Green("\n  Provider switched to: %s\n", p.Name())
	} else {
		color.Red("  Error: %v\n", err)
		time.Sleep(2 * time.Second)
	}
}

func runOneShot(prompt string) error {
	p, err := ai.NewProvider()
	if err != nil {
		return err
	}
	res, err := p.Send(prompt)
	if err != nil {
		return err
	}
	fmt.Println(res)
	return nil
}

func initConfig() {
	godotenv.Load()
}

func runUninstaller() {
	if err := runSelfUninstall(); err != nil {
		color.Red("\n  Error during uninstall: %v\n", err)
		time.Sleep(3 * time.Second)
	}
}
