package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/broman0x/forgeai-cli/cmd"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("\n------------------------------------------------")
			fmt.Printf("CRITICAL PANIC: %v\n", r)
			fmt.Println("------------------------------------------------")
		}
		pauseExit()
	}()

	if err := cmd.Execute(); err != nil {
		if err.Error() != "" {
			fmt.Println("\n[!] Program Exited with Error:", err)
		}
		os.Exit(1)
	}
}

func pauseExit() {
	fmt.Println("\nPress 'Enter' to close window...")
	bufio.NewReader(os.Stdin).ReadBytes('\n')
}