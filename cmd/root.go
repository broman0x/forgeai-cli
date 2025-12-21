package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/broman0x/forgeai-cli/internal/ai"
	"github.com/broman0x/forgeai-cli/internal/config"
	"github.com/broman0x/forgeai-cli/internal/ui"
	"github.com/fatih/color"
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile         string
	noBanner        bool
	currentProvider ai.Provider
)

var rootCmd = &cobra.Command{
	Use:   "forgeai",
	Short: "Professional AI CLI",
	Args:  cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
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
	cobra.MousetrapHelpText = ""
}

func runMainMenu() error {
	ui.ShowStartupBanner()
	
	var err error
	currentProvider, err = ai.NewProvider()
	
	if err != nil {
		return runSetupWizard(err)
	}

	scanner := bufio.NewScanner(os.Stdin)
	cBracket := color.New(color.FgHiBlack).SprintFunc()
	cNum := color.New(color.FgCyan, color.Bold).SprintFunc()
	cName := color.New(color.FgHiWhite).SprintFunc()
	cDesc := color.New(color.FgWhite, color.Faint).SprintFunc()
	cActive := color.New(color.BgCyan, color.FgBlack, color.Bold).SprintFunc()

	for {
		fmt.Printf("  %s %s\n\n", cActive(" ACTIVE BRAIN "), currentProvider.Name())

		printMenu := func(key, name, desc string) {
			fmt.Printf("  %s %s %s   %-14s %s\n", 
				cBracket("["), cNum(key), cBracket("]"), 
				cName(name), cDesc(desc))
		}

		printMenu("1", "Chat Mode",    "Interactive conversation")
		printMenu("2", "Code Review",  "Scan file for bugs")
		printMenu("3", "Code Editor",  "Edit file with AI diff")
		printMenu("4", "Switch Model", "Change AI provider")
		printMenu("5", "System Info",  "Hardware dashboard")
		fmt.Println()
		printMenu("0", "Exit",         "Close application")
		
		fmt.Print("\n  Select Command > ")

		if !scanner.Scan() { break }
		choice := strings.TrimSpace(scanner.Text())

		switch choice {
		case "1": startChatMode(scanner)
		case "2": StartReviewModeInteractive(scanner, currentProvider)
		case "3": StartEditModeInteractive(scanner, currentProvider)
		case "4": handleSwitchModel(scanner)
		case "5": StartInfoModeInteractive(scanner)
		case "0", "exit":
			color.Cyan("\n  Shutting down... Goodbye!\n")
			return nil
		default:
			fmt.Print("\033[H\033[2J")
			ui.ShowStartupBanner()
		}
	}
	return nil
}

func runSetupWizard(originalErr error) error {
	color.Yellow("\n  [!] SETUP REQUIRED")
	fmt.Println("  ForgeAI needs an API Key to function properly.")
	fmt.Printf("  (Error: %v)\n", originalErr)
	fmt.Println()
	
	fmt.Println("  1. Setup Google Gemini Key (Recommended/Free)")
	fmt.Println("  2. I want to use Ollama (Local Only)")
	fmt.Println("  0. Exit")
	
	fmt.Print("\n  Select > ")
	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() { return nil }
	
	choice := strings.TrimSpace(scanner.Text())
	
	if choice == "1" {
		fmt.Println("\n  Get your key here: https://aistudio.google.com/app/apikey")
		fmt.Print("  Paste Gemini API Key: ")
		if scanner.Scan() {
			apiKey := strings.TrimSpace(scanner.Text())
			if apiKey == "" {
				color.Red("  Empty key provided.")
				return fmt.Errorf("setup aborted")
			}

			err := config.CreateEnvFile(apiKey)
			if err != nil {
				color.Red("  Failed to create .env file: %v", err)
				return err
			}
			
			color.Green("  [OK] Configuration saved! Restarting...")
			time.Sleep(1 * time.Second)
			
			godotenv.Load() 
			fmt.Print("\033[H\033[2J")
			return runMainMenu()
		}
	} else if choice == "2" {
		color.Yellow("\n  Make sure Ollama is running on port 11434.")
		fmt.Println("  Press Enter to check connection...")
		scanner.Scan()
		fmt.Print("\033[H\033[2J")
		return runMainMenu()
	}
	
	return nil
}

func startChatMode(scanner *bufio.Scanner) {
	fmt.Print("\033[H\033[2J")
	ui.ShowStartupBanner()
	ui.PrintHeader("CHAT INTERFACE")
	
	fmt.Println(color.New(color.FgHiBlack).Sprint("  ----------------------------------------------------"))
	fmt.Println()

	cUser := color.New(color.FgHiGreen, color.Bold).SprintFunc()
	cAI := color.New(color.FgHiCyan, color.Bold).SprintFunc()
	cArrow := color.New(color.FgHiBlack).SprintFunc()
	cDiv := color.New(color.FgHiBlack).SprintFunc()
	cFaint := color.New(color.FgWhite, color.Faint).SprintFunc()

	for {
		fmt.Printf("  %s\n  %s ", cUser("USER"), cArrow(">"))
		
		if !scanner.Scan() { break }
		input := strings.TrimSpace(scanner.Text())
		
		if input == "" { continue }
		if input == "back" || input == "exit" {
			fmt.Print("\033[H\033[2J")
			ui.ShowStartupBanner()
			break
		}
		if input == "clear" || input == "cls" {
			fmt.Print("\033[H\033[2J")
			ui.ShowStartupBanner()
			ui.PrintHeader("CHAT INTERFACE")
			continue
		}

		fmt.Print("\n  " + cFaint("Thinking..."))
		resp, err := currentProvider.Send(input)
		fmt.Print("\r\033[K") 

		if err != nil {
			color.Red("  [ERROR] %v\n", err)
		} else {
			fmt.Printf("  %s\n", cAI("FORGE AI"))
			streamPrintIndented(resp, "   ") 
			
			fmt.Println()
			fmt.Println("  " + cDiv(strings.Repeat("-", 50)))
			fmt.Println()
		}
	}
}

func streamPrintIndented(text, indent string) {
	lines := strings.Split(text, "\n")
	for _, line := range lines {
		fmt.Print(indent)
		if strings.HasPrefix(strings.TrimSpace(line), "```") {
			color.New(color.FgYellow).Println(line)
		} else {
			chars := []rune(line)
			for _, char := range chars {
				fmt.Print(string(char))
				time.Sleep(1 * time.Millisecond)
			}
			fmt.Println()
		}
	}
}

func handleSwitchModel(scanner *bufio.Scanner) {
	ui.PrintHeader("SWITCH BRAIN")
	fmt.Println("  1. Gemini 2.5 Flash")
	fmt.Println("  2. Gemini Pro")
	fmt.Println("  3. Ollama (Custom Model)")
	fmt.Print("\n  Selection: ")
	
	if !scanner.Scan() { return }
	
	var p ai.Provider
	var err error
	
	switch strings.TrimSpace(scanner.Text()) {
	case "1": p, err = ai.CreateProvider("gemini", "gemini-2.5-flash")
	case "2": p, err = ai.CreateProvider("gemini", "gemini-pro")
	case "3": 
		fmt.Print("  Enter Model Name (default: llama3): ")
		scanner.Scan()
		model := strings.TrimSpace(scanner.Text())
		if model == "" { model = "llama3" }
		p, err = ai.CreateProvider("ollama", model)
	default: return
	}

	if err == nil {
		currentProvider = p
		fmt.Print("\033[H\033[2J")
		ui.ShowStartupBanner()
		color.Green("\n  [OK] Brain switched to: %s\n", p.Name())
	} else {
		color.Red("  [ERROR] Failed: %v\n", err)
		time.Sleep(2 * time.Second)
	}
}

func runOneShot(prompt string) error {
	p, err := ai.NewProvider()
	if err != nil { return err }
	res, err := p.Send(prompt)
	if err != nil { return err }
	fmt.Println(res)
	return nil
}

func initConfig() {
	godotenv.Load()
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, _ := os.UserHomeDir()
		configPath := filepath.Join(home, ".forgeai")
		os.MkdirAll(configPath, 0755)
		viper.AddConfigPath(configPath)
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
	}
	viper.AutomaticEnv()
	viper.SetDefault("first_run", true)
	viper.SetDefault("provider", "gemini")
	viper.SetDefault("model", "gemini-2.5-flash")
	viper.ReadInConfig()

	cfg := config.Load()
	if cfg.FirstRun {
		config.SaveFirstRun(false)
	}
}