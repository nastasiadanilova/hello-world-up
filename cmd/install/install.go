// установка и настройка автозапуска системы мониторинга Hello World

// установка и настройка автозапуска системы мониторинга Hello World

package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

const appDir = "hello-world"

func main() {
	fmt.Println("установка Hello World + мониторинг")
	fmt.Println("ОС:", runtime.GOOS)

	// Сборка обоих бинарников из cmd/monitor/
	mustBuild("web", "../monitor/web.go")
	mustBuild("monitor", "../monitor/monitor.go")

	currentDir, _ := os.Getwd()
	targetDir := filepath.Join(currentDir, appDir)
	os.MkdirAll(targetDir, 0755)

	copyBinary("web", targetDir)
	copyBinary("monitor", targetDir)

	fmt.Printf("готово → %s\n", targetDir)

	switch runtime.GOOS {
	case "linux":
		linuxSetup(targetDir)
	case "darwin":
		macSetup(targetDir)
	case "windows":
		winSetup(targetDir)
	default:
		fmt.Println("автозапуск не поддерживается на этой ОС")
	}

	fmt.Println("\nВсё готово!")
	fmt.Println("→ http://127.0.0.1:8080")
	fmt.Printf("→ логи: tail -f %s\n", filepath.Join(targetDir, "monitor.log"))
}

func mustBuild(name, source string) {
	out := name
	if runtime.GOOS == "windows" {
		out += ".exe"
	}
	fmt.Printf("сборка %s → %s ... ", source, out)
	cmd := exec.Command("go", "build", "-o", out, source)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Fatalf("ошибка сборки %s: %v", name, err)
	}
	fmt.Println("OK")
}

func copyBinary(name, targetDir string) {
	src := name
	if runtime.GOOS == "windows" {
		src += ".exe"
	}
	data, err := os.ReadFile(src)
	if err != nil {
		log.Fatal(err)
	}
	dst := filepath.Join(targetDir, name)
	if runtime.GOOS == "windows" {
		dst += ".exe"
	}
	os.WriteFile(dst, data, 0755)
}

// !!! автозапуск для Linux (systemd --user) !!!
func linuxSetup(dir string) {
	monitorPath := filepath.Join(dir, "monitor")
	service := fmt.Sprintf(`[Unit]
Description=Hello World Monitor
After=network.target

[Service]
Type=simple
ExecStart=%s
WorkingDirectory=%s
Restart=always
RestartSec=5

[Install]
WantedBy=default.target
`, monitorPath, dir)

	userDir := filepath.Join(os.Getenv("HOME"), ".config", "systemd", "user")
	os.MkdirAll(userDir, 0755)
	servicePath := filepath.Join(userDir, "hello-monitor.service")

	if err := os.WriteFile(servicePath, []byte(service), 0644); err != nil {
		fmt.Println("не удалось создать systemd-сервис:", err)
		return
	}

	exec.Command("systemctl", "--user", "daemon-reload").Run()
	exec.Command("systemctl", "--user", "enable", "--now", "hello-monitor.service").Run()
	fmt.Println("Linux: автозапуск включён (systemd --user)")
}

// !!! автозапуск для macOS (LaunchAgent) !!!
func macSetup(dir string) {
	monitorPath := filepath.Join(dir, "monitor")
	plist := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>Label</key><string>local.helloworld.monitor</string>
	<key>ProgramArguments</key><array><string>%s</string></array>
	<key>WorkingDirectory</key><string>%s</string>
	<key>RunAtLoad</key><true/>
	<key>KeepAlive</key><true/>
	<key>StandardOutPath</key><string>%s</string>
	<key>StandardErrorPath</key><string>%s</string>
</dict>
</plist>`, monitorPath, dir, filepath.Join(dir, "monitor.log"), filepath.Join(dir, "monitor.log"))

	launchDir := filepath.Join(os.Getenv("HOME"), "Library", "LaunchAgents")
	os.MkdirAll(launchDir, 0755)
	plistPath := filepath.Join(launchDir, "local.helloworld.monitor.plist")

	if err := os.WriteFile(plistPath, []byte(plist), 0644); err != nil {
		fmt.Println("не удалось создать LaunchAgent:", err)
		return
	}

	// перезагружаем (если уже был)
	exec.Command("launchctl", "unload", plistPath).Run()
	exec.Command("launchctl", "load", plistPath).Run()
	fmt.Println("macOS: автозапуск включён (LaunchAgent)")
}

// !!! автозапуск для Windows (ярлык в автозагрузку) !!!
func winSetup(dir string) {
	batContent := "@echo off\r\n" +
		"cd /d \"" + dir + "\"\r\n" +
		"start \"\" monitor.exe\r\n"

	startupDir := filepath.Join(os.Getenv("APPDATA"), "Microsoft", "Windows", "Start Menu", "Programs", "Startup")
	batPath := filepath.Join(startupDir, "HelloWorld-Monitor.bat")

	if err := os.WriteFile(batPath, []byte(batContent), 0755); err != nil {
		fmt.Println("не удалось создать bat-файл в автозагрузке:", err)
		return
	}
	fmt.Println("Windows: автозапуск включён (через Startup)")
}
