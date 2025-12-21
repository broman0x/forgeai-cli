package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/broman0x/forgeai-cli/internal/ai"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var fileContext string

var askCmd = &cobra.Command{
	Use:   "ask [prompt]",
	Short: "Ask a question to the AI",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		prompt := strings.Join(args, " ")

		provider, err := ai.NewProvider()
		if err != nil {
			color.Red("Error: %v", err)
			return
		}

		finalPrompt := prompt
		if fileContext != "" {
			content, _ := os.ReadFile(fileContext)

			sb := strings.Builder{}
			sb.WriteString("Context file (")
			sb.WriteString(fileContext)
			sb.WriteString("):\n\n")
			sb.WriteString(string(content))
			sb.WriteString("\n\nQuestion: ")
			sb.WriteString(prompt)

			finalPrompt = sb.String()

			fmt.Printf("Using context from: %s\n", fileContext)
		}

		fmt.Printf("Asking %s...\n", provider.Name())
		response, err := provider.Send(finalPrompt)
		if err != nil {
			color.Red("Failed: %v", err)
			return
		}

		color.Cyan("\n%s\n", response)
	},
}

func init() {
	rootCmd.AddCommand(askCmd)
	askCmd.Flags().StringVarP(&fileContext, "file", "f", "", "Attach file context")
}
