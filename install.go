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

const appDir = "hello-world" // папка, куда всё установится

func main() {
	fmt.Println("установка веб-приложения, которое возвращает Hello World! + скрипта мониторинга")
	fmt.Println("ОС:", runtime.GOOS)

	// 1. сбор бинарников (web и monitor)
	mustBuild("web", "web.go")
	mustBuild("monitor", "monitor.go")

	// 2. папка установки рядом с текущим проектом
	currentDir, _ := os.Getwd()
	targetDir := filepath.Join(currentDir, appDir)
	os.MkdirAll(targetDir, 0755)

	copyBinary("web", targetDir)
	copyBinary("monitor", targetDir)

	fmt.Printf("приложение установлено в: %s\n", targetDir)

	// 3. настройка автозапуска мониторинга
	switch runtime.GOOS {
	case "linux":
		linuxSetup(targetDir)
	case "darwin":
		macSetup(targetDir)
	case "windows":
		winSetup(targetDir)
	default:
		fmt.Println("автозапуск для вашей ОС не реализован (но можно запускать вручную, подробности в README)")
	}

	fmt.Println("\nура ура, готово!")
	fmt.Println("проверить работу: http://127.0.0.1:8080")
	fmt.Printf("логи мониторинга: tail -f %s\n", filepath.Join(targetDir, "monitor.log"))
	fmt.Println("остановить мониторинг: pkill -f monitor   (или taskkill на Windows)")
}

// сборка одного бинарника
func mustBuild(name, source string) {
	out := name
	if runtime.GOOS == "windows" {
		out += ".exe"
	}
	fmt.Printf("сборка %s → %s ... ", source, out)
	cmd := exec.Command("go", "build", "-o", out, source)
	if err := cmd.Run(); err != nil {
		log.Fatalf("\nошибка сборки %s: %v", name, err)
	}
	fmt.Println("OK")
}

// копирование бинарника с правильным расширением
func copyBinary(name, targetDir string) {
	src := name
	if runtime.GOOS == "windows" {
		src += ".exe"
	}
	data, err := os.ReadFile(src)
	if err != nil {
		log.Fatalf("не могу прочитать %s: %v", src, err)
	}
	dst := filepath.Join(targetDir, name)
	if runtime.GOOS == "windows" {
		dst += ".exe"
	}
	if err := os.WriteFile(dst, data, 0755); err != nil {
		log.Fatalf("не могу скопировать %s: %v", name, err)
	}
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
