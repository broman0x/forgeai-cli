package cmd

import (
	"bufio"
	"fmt"

	"github.com/broman0x/forgeai-cli/internal/ui"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Show system info and configuration dashboard",
	Run: func(cmd *cobra.Command, args []string) {
		showInfoLogic()
	},
}

func init() {
	rootCmd.AddCommand(infoCmd)
}

func StartInfoModeInteractive(scanner *bufio.Scanner) {
	fmt.Print("\033[H\033[2J")

	showInfoLogic()

	fmt.Println("\n  Press Enter to return to menu...")
	scanner.Scan()
	fmt.Print("\033[H\033[2J")
}

func showInfoLogic() {
	ui.ShowStartupBanner()

	cLabel := color.New(color.FgHiBlack).SprintFunc()
	cValue := color.New(color.FgWhite).SprintFunc()

	fmt.Println("  Configuration:")
	configFile := viper.ConfigFileUsed()
	if configFile == "" {
		configFile = "Default (No file found)"
	}

	fmt.Printf("   • %s %s\n", cLabel("Config Path:"), cValue(configFile))
	fmt.Printf("   • %s %s\n", cLabel("AI Provider:"), cValue(viper.GetString("provider")))
	fmt.Printf("   • %s %s\n", cLabel("AI Model:   "), cValue(viper.GetString("model")))
	fmt.Println()
}
