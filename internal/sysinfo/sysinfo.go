package sysinfo

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
)

type SystemDetails struct {
	OS              string
	Arch            string
	CPUModel        string
	CPUCores        int
	TotalRAM        string
	GPUName         string
	InternetConn    bool
	OllamaActive    bool
	OllamaModelPath string
}

func GetSystemDetails() (*SystemDetails, error) {
	details := &SystemDetails{
		OS:   runtime.GOOS,
		Arch: runtime.GOARCH,
	}

	details.CPUCores = runtime.NumCPU()
	cpuInfo, err := cpu.Info()
	if err == nil && len(cpuInfo) > 0 {
		details.CPUModel = cpuInfo[0].ModelName
	} else {
		details.CPUModel = "Generic CPU"
	}

	vmStat, err := mem.VirtualMemory()
	if err == nil {
		details.TotalRAM = fmt.Sprintf("%.1f GB", float64(vmStat.Total)/1024/1024/1024)
	} else {
		details.TotalRAM = "Unknown"
	}

	details.GPUName = getGPUName()

	details.InternetConn = checkInternetRaw()
	details.OllamaActive = checkOllamaRaw()

	homeDir, err := os.UserHomeDir()
	if err == nil {
		details.OllamaModelPath = filepath.Join(homeDir, ".ollama", "models")
	} else {
		details.OllamaModelPath = "Unknown"
	}

	return details, nil
}

func getGPUName() string {
	if runtime.GOOS == "windows" {
		cmd := exec.Command("wmic", "path", "win32_VideoController", "get", "name")
		out, err := cmd.Output()
		if err == nil {
			lines := strings.Split(string(out), "\n")
			for _, line := range lines {
				clean := strings.TrimSpace(line)
				if clean != "" && clean != "Name" {
					return clean
				}
			}
		}
	}
	return "Integrated / Not Detected"
}

func checkInternetRaw() bool {
	timeout := 2 * time.Second
	conn, err := net.DialTimeout("tcp", "8.8.8.8:53", timeout)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

func checkOllamaRaw() bool {
	timeout := 100 * time.Millisecond
	conn, err := net.DialTimeout("tcp", "localhost:11434", timeout)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}
