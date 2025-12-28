package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/fatih/color"
)

func runSelfInstall() error {
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("could not get executable path: %v", err)
	}

	var installDir, destPath string

	if runtime.GOOS == "windows" {
		installDir = filepath.Join(os.Getenv("LocalAppData"), "ForgeAI")
		destPath = filepath.Join(installDir, "forge.exe")
	} else {
		home, _ := os.UserHomeDir()
		installDir = filepath.Join(home, ".local", "bin")
		destPath = filepath.Join(installDir, "forge")
	}

	fmt.Println()
	color.Cyan("  ╔══════════════════════════════════════════╗")
	color.Cyan("  ║     ForgeAI CLI - Self Installer         ║")
	color.Cyan("  ╚══════════════════════════════════════════╝")
	fmt.Println()

	if filepath.Clean(exePath) == filepath.Clean(destPath) {
		color.Yellow("  Already installed and running from install location!")
		fmt.Println()

		if runtime.GOOS != "windows" {
			color.Cyan("  Checking PATH configuration...")
			time.Sleep(500 * time.Millisecond)
			if err := ensurePathInShellRC(installDir); err == nil {
				color.Green("  ✓ PATH configured in shell")
			} else {
				color.Green("  ✓ Already in PATH")
			}
		} else {
			color.Cyan("  Checking PATH configuration...")
			time.Sleep(500 * time.Millisecond)
			cmd := exec.Command("powershell", "-Command",
				fmt.Sprintf("$path = [Environment]::GetEnvironmentVariable('Path', 'User'); if ($path -notlike '*%s*') { [Environment]::SetEnvironmentVariable('Path', $path + ';%s', 'User'); exit 0 } else { exit 1 }", installDir, installDir))
			if err := cmd.Run(); err == nil {
				color.Green("  ✓ Added to PATH")
			} else {
				color.Green("  ✓ Already in PATH")
			}
		}

		fmt.Println()
		color.Green("  ══════════════════════════════════════════")
		color.Green("  ForgeAI is already installed!")
		color.Green("  ══════════════════════════════════════════")
		fmt.Println()
		color.Cyan("  You can run 'forge' from any directory")
		fmt.Println()
		return nil
	}

	if err := os.MkdirAll(installDir, 0755); err != nil {
		return fmt.Errorf("could not create install directory: %v", err)
	}

	fmt.Print("  [1/2] Copying executable...")
	time.Sleep(200 * time.Millisecond)

	input, err := os.ReadFile(exePath)
	if err != nil {
		color.Red(" Failed")
		return fmt.Errorf("could not read executable: %v", err)
	}

	if err := os.WriteFile(destPath, input, 0755); err != nil {
		color.Red(" Failed")
		return fmt.Errorf("could not copy executable: %v", err)
	}
	color.Green(" Done ✓")

	fmt.Print("  [2/2] Configuring PATH...")
	time.Sleep(200 * time.Millisecond)

	if runtime.GOOS == "windows" {
		cmd := exec.Command("powershell", "-Command",
			fmt.Sprintf("$path = [Environment]::GetEnvironmentVariable('Path', 'User'); if ($path -notlike '*%s*') { [Environment]::SetEnvironmentVariable('Path', $path + ';%s', 'User') }", installDir, installDir))
		if err := cmd.Run(); err != nil {
			color.Yellow(" Warning")
			fmt.Println()
			color.Yellow("  Please add manually: %s", installDir)
		} else {
			color.Green(" Done ✓")
		}
	} else {
		if err := ensurePathInShellRC(installDir); err != nil {
			color.Yellow(" Warning")
			fmt.Println()
			color.Yellow("  Please add to your shell rc file:")
			color.Cyan("  export PATH=\"$HOME/.local/bin:$PATH\"")
		} else {
			color.Green(" Done ✓")
		}
	}

	fmt.Println()
	fmt.Println()
	color.Green("  ══════════════════════════════════════════")
	color.Green("  Installation complete!")
	color.Green("  ══════════════════════════════════════════")
	fmt.Println()

	if runtime.GOOS == "windows" {
		color.Cyan("  Run 'forge' from any directory")
		color.Yellow("  Note: Restart terminal for PATH changes")
	} else {
		color.Cyan("  Installed to: %s", destPath)
		color.Yellow("  Run: source ~/.bashrc  (or ~/.zshrc)")
		color.Cyan("  Then: forge")
	}
	fmt.Println()

	return nil
}

func ensurePathInShellRC(installDir string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	pathExport := fmt.Sprintf("export PATH=\"$HOME/.local/bin:$PATH\"")

	shell := os.Getenv("SHELL")
	var rcFiles []string

	if strings.Contains(shell, "zsh") {
		rcFiles = []string{filepath.Join(home, ".zshrc")}
	} else if strings.Contains(shell, "bash") {
		rcFiles = []string{
			filepath.Join(home, ".bashrc"),
			filepath.Join(home, ".bash_profile"),
		}
	} else {
		rcFiles = []string{
			filepath.Join(home, ".bashrc"),
			filepath.Join(home, ".zshrc"),
			filepath.Join(home, ".profile"),
		}
	}

	updated := false
	for _, rcFile := range rcFiles {
		if _, err := os.Stat(rcFile); os.IsNotExist(err) {
			continue
		}

		content, err := os.ReadFile(rcFile)
		if err != nil {
			continue
		}

		if strings.Contains(string(content), "$HOME/.local/bin") {
			updated = true
			continue
		}

		f, err := os.OpenFile(rcFile, os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			continue
		}

		fmt.Fprintf(f, "\n# Added by ForgeAI CLI installer\n%s\n", pathExport)
		f.Close()
		updated = true
	}

	if !updated {
		return fmt.Errorf("could not update shell rc files")
	}

	return nil
}

func runSelfUninstall() error {
	var installDir, exePath, configDir string

	if runtime.GOOS == "windows" {
		installDir = filepath.Join(os.Getenv("LocalAppData"), "ForgeAI")
		exePath = filepath.Join(installDir, "forge.exe")
		configDir = filepath.Join(os.Getenv("APPDATA"), "ForgeAI")
	} else {
		home, _ := os.UserHomeDir()
		installDir = filepath.Join(home, ".local", "bin")
		exePath = filepath.Join(installDir, "forge")
		configDir = filepath.Join(home, ".config", "forgeai")
	}

	fmt.Println()
	color.New(color.FgYellow, color.Bold).Println("  ╔══════════════════════════════════════════╗")
	color.New(color.FgYellow, color.Bold).Println("  ║    ForgeAI CLI - Uninstaller v" + Version + "       ║")
	color.New(color.FgYellow, color.Bold).Println("  ╚══════════════════════════════════════════╝")
	fmt.Println()

	color.Yellow("  This will remove:")
	fmt.Println("  • Executable from PATH")
	fmt.Println("  • All configuration files")
	fmt.Println("  • Saved preferences")
	fmt.Println()
	color.Red("  ⚠ This action cannot be undone!")
	fmt.Println()
	fmt.Print("  Continue? [y/N]: ")

	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() || strings.ToLower(strings.TrimSpace(scanner.Text())) != "y" {
		color.Cyan("\n  Uninstall cancelled.\n")
		return nil
	}

	fmt.Println()
	color.Cyan("  Starting uninstallation...")
	time.Sleep(500 * time.Millisecond)
	fmt.Println()

	if runtime.GOOS == "windows" {
		fmt.Print("  [1/3] Removing from PATH...")
		cmd := exec.Command("powershell", "-Command",
			fmt.Sprintf("$path = [Environment]::GetEnvironmentVariable('Path', 'User'); $newPath = ($path -split ';' | Where-Object { $_ -ne '%s' }) -join ';'; [Environment]::SetEnvironmentVariable('Path', $newPath, 'User')", installDir))
		if err := cmd.Run(); err != nil {
			color.Red(" Failed")
		} else {
			color.Green(" Done ✓")
		}

		fmt.Print("  [2/3] Removing configuration...")
		time.Sleep(300 * time.Millisecond)
		if err := os.RemoveAll(configDir); err != nil && !os.IsNotExist(err) {
			color.Red(" Failed")
		} else {
			color.Green(" Done ✓")
		}

		fmt.Print("  [3/3] Scheduling executable removal...")
		time.Sleep(300 * time.Millisecond)

		scriptPath := filepath.Join(os.TempDir(), "forge_uninstall.ps1")
		script := fmt.Sprintf(`Start-Sleep -Seconds 2
Remove-Item -Path "%s" -Force -ErrorAction SilentlyContinue
Remove-Item -Path "%s" -Recurse -Force -ErrorAction SilentlyContinue
Remove-Item -Path "%s" -Force
`, exePath, installDir, scriptPath)

		if err := os.WriteFile(scriptPath, []byte(script), 0644); err != nil {
			color.Red(" Failed")
		} else {
			cmd := exec.Command("powershell", "-WindowStyle", "Hidden", "-File", scriptPath)
			cmd.Start()
			color.Green(" Done ✓")
		}
	} else {
		fmt.Print("  [1/2] Removing executable...")
		if err := os.Remove(exePath); err != nil && !os.IsNotExist(err) {
			color.Red(" Failed")
		} else {
			color.Green(" Done ✓")
		}

		fmt.Print("  [2/2] Removing configuration...")
		time.Sleep(300 * time.Millisecond)
		if err := os.RemoveAll(configDir); err != nil && !os.IsNotExist(err) {
			color.Red(" Failed")
		} else {
			color.Green(" Done ✓")
		}

		fmt.Println()
		color.Yellow("  Note: Please manually remove this line from your shell rc file:")
		color.Cyan("  export PATH=\"$HOME/.local/bin:$PATH\"")
	}

	fmt.Println()
	fmt.Println()
	color.Green("  ══════════════════════════════════════════")
	color.Green("  Uninstallation complete!")
	color.Green("  ══════════════════════════════════════════")
	fmt.Println()
	color.Cyan("  Thank you for using ForgeAI CLI!")

	if runtime.GOOS == "windows" {
		color.Yellow("  Restart your terminal for PATH changes to take effect")
		fmt.Println()
		fmt.Println("  ForgeAI will exit in 3 seconds...")
		time.Sleep(3 * time.Second)
	} else {
		fmt.Println("  ForgeAI will exit now...")
		time.Sleep(2 * time.Second)
	}

	return nil
}
