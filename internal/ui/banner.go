package ui

import (
	"fmt"
	"strings"

	"github.com/broman0x/forgeai-cli/internal/sysinfo"
	"github.com/fatih/color"
)

func ShowStartupBanner() {
	cTitle := color.New(color.FgHiCyan, color.Bold)
	cSubtitle := color.New(color.FgCyan)
	cLabel := color.New(color.FgHiWhite).SprintFunc()
	cValue := color.New(color.FgWhite).SprintFunc()
	cAccent := color.New(color.FgHiMagenta).SprintFunc()
	cSuccess := color.New(color.FgGreen).SprintFunc()
	cFail := color.New(color.FgRed).SprintFunc()
	cBorder := color.New(color.FgHiBlack).SprintFunc()

	fmt.Println()

	cTitle.Println("  ███████╗ ██████╗ ██████╗  ██████╗ ███████╗")
	cTitle.Println("  ██╔════╝██╔═══██╗██╔══██╗██╔════╝ ██╔════╝")
	cTitle.Println("  █████╗  ██║   ██║██████╔╝██║  ███╗█████╗  ")
	cTitle.Println("  ██╔══╝  ██║   ██║██╔══██╗██║   ██║██╔══╝  ")
	cTitle.Println("  ██║     ╚██████╔╝██║  ██║╚██████╔╝███████╗")
	cTitle.Println("  ╚═╝      ╚═════╝ ╚═╝  ╚═╝ ╚═════╝ ╚══════╝")

	cSubtitle.Println("          AI-Powered Development Assistant")
	fmt.Printf("          %s\n", cBorder("v1.0.1 • by bromanprjkt"))
	fmt.Println()

	info, err := sysinfo.GetSystemDetails()
	if err != nil {
		return
	}

	fmt.Println(cBorder("  ┌───────────────────────────────────────────────────────┐"))

	printCompactRow := func(icon, label, value string, status ...bool) {
		useSuccess := len(status) > 0 && status[0]
		useFail := len(status) > 1 && status[1]

		colorFunc := cValue
		if useSuccess {
			colorFunc = cSuccess
		} else if useFail {
			colorFunc = cFail
		}

		displayValue := value
		if len(value) > 45 {
			displayValue = value[:42] + "..."
		}

		fmt.Printf("%s  %s %-12s %s %s\n",
			cBorder("  │"),
			cAccent(icon),
			cLabel(label),
			cBorder("│"),
			colorFunc(displayValue))
	}

	printCompactRow(">", "Platform", fmt.Sprintf("%s/%s", info.OS, info.Arch))
	printCompactRow(">", "CPU", info.CPUModel)
	printCompactRow(">", "Cores", fmt.Sprintf("%d threads", info.CPUCores))
	printCompactRow(">", "Memory", info.TotalRAM)

	if info.GPUName != "" && info.GPUName != "N/A" {
		printCompactRow(">", "GPU", info.GPUName)
	}

	fmt.Println(cBorder("  ├───────────────────────────────────────────────────────┤"))

	if info.InternetConn {
		printCompactRow("+", "Network", "Connected", true, false)
	} else {
		printCompactRow("X", "Network", "Offline", false, true)
	}

	if info.OllamaActive {
		printCompactRow("+", "Ollama", "Running", true, false)
	} else {
		printCompactRow("X", "Ollama", "Not Available", false, true)
	}

	if info.OllamaModelPath != "" && info.OllamaModelPath != "N/A" {
		printCompactRow(">", "Models", info.OllamaModelPath)
	}

	fmt.Println(cBorder("  └───────────────────────────────────────────────────────┘"))
	fmt.Println()
}

func PrintHeader(title string) {
	cBox := color.New(color.FgHiCyan, color.Bold).SprintFunc()
	fmt.Println()
	fmt.Println(cBox("  " + strings.ToUpper(title)))
	fmt.Println()
}
