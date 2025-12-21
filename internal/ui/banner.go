package ui

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/broman0x/forgeai-cli/internal/sysinfo"
	"github.com/common-nighthawk/go-figure"
	"github.com/fatih/color"
)

func ShowStartupBanner() {
	cSub := color.New(color.FgHiBlack)
	cLabel := color.New(color.FgHiWhite).SprintFunc()
	cValue := color.New(color.FgWhite, color.Faint).SprintFunc()
	cBorder := color.New(color.FgHiBlue).SprintFunc()
	cSuccess := color.New(color.FgGreen, color.Bold).SprintFunc()
	cFail := color.New(color.FgRed).SprintFunc()
	cAccent := color.New(color.FgCyan).SprintFunc()

	fmt.Println()
	myFigure := figure.NewColorFigure("THE FORGE", "slant", "cyan", true)
	myFigure.Print()
	cSub.Println("   v1.0.0 • by bromanprjkt")
	fmt.Println()

	info, err := sysinfo.GetSystemDetails()
	if err != nil {
		return
	}

	width := 60
	hr := strings.Repeat("-", width)

	fmt.Println(cBorder("  +" + hr + "+"))

	printRow := func(label, value string, valColor func(a ...interface{}) string) {
		if valColor == nil {
			valColor = cValue
		}
		prefix := "  " + cAccent("+") + " " + cLabel(label)
		rawPrefix := "  + " + label
		cleanVal := strings.TrimSpace(value)
		lenPrefix := utf8.RuneCountInString(rawPrefix)
		lenVal := utf8.RuneCountInString(cleanVal)
		gap := width - lenPrefix - lenVal - 1

		if gap < 2 {
			maxLen := width - lenPrefix - 4
			if maxLen > 10 {
				head := cleanVal[:10]
				tail := cleanVal[lenVal-(maxLen-13):]
				cleanVal = head + "..." + tail
				lenVal = utf8.RuneCountInString(cleanVal)
				gap = width - lenPrefix - lenVal - 1
			} else {
				gap = 1
			}
		}

		fmt.Printf(cBorder("  |")+"%s%s%s "+cBorder("|")+"\n",
			prefix,
			strings.Repeat(" ", gap),
			valColor(cleanVal),
		)
	}

	printRow("OS Platform", fmt.Sprintf("%s/%s", info.OS, info.Arch), nil)
	printRow("CPU Chipset", info.CPUModel, nil)
	printRow("Core Threads", fmt.Sprintf("%d Cores", info.CPUCores), nil)
	printRow("RAM Capacity", info.TotalRAM, nil)
	printRow("GPU Adapter", info.GPUName, nil)

	fmt.Println(cBorder("  +" + hr + "+"))
	netTxt, netClr := "Offline [X]", cFail
	if info.InternetConn {
		netTxt, netClr = "Connected [OK]", cSuccess
	}
	printRow("Network", netTxt, netClr)

	aiTxt, aiClr := "Disconnected [X]", cFail
	if info.OllamaActive {
		aiTxt, aiClr = "Online (Local) [OK]", cSuccess
	}
	printRow("Ollama Service", aiTxt, aiClr)

	printRow("Model Storage", info.OllamaModelPath, nil)

	fmt.Println(cBorder("  +" + hr + "+"))
	fmt.Println()
}

func PrintHeader(title string) {
	cBox := color.New(color.FgHiCyan, color.Bold).SprintFunc()
	fmt.Println()
	fmt.Println(cBox("  :: " + strings.ToUpper(title)))
	fmt.Println()
}
