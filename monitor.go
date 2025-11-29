// мониторинг простого веб-приложения "Hello World!"
// автоматически запускает/перезапускает приложение при сбоях или недоступности по HTTP

package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"
)

const (
	appBinaryName = "web" // имя бинарника веб-приложения
	healthURL     = "http://127.0.0.1:8080"
	interval      = 10 * time.Second // как часто проверяем
	timeout       = 3 * time.Second  // таймаут HTTP-запроса
)

var logPath string // полный путь к monitor.log (определяется один раз)

// логгер, который сразу сбрасывает записи на диск
func logf(format string, v ...interface{}) {
	log.Printf(format, v...)
	if f, ok := log.Writer().(*os.File); ok {
		f.Sync() // принудительно сбрасываем буфер, файл видно сразу
	}
}

func main() {
	// определяем, где лежит сам бинарник monitor → лог будет рядом с ним
	exe, err := os.Executable()
	if err != nil {
		log.Fatal("не удалось определить путь к исполняемому файлу:", err)
	}
	logPath = filepath.Join(filepath.Dir(exe), "monitor.log")

	// открываем лог-файл рядом с monitor
	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Fatal("не удалось открыть monitor.log:", err)
	}
	defer f.Close()

	// настройка стандартного логгера
	log.SetOutput(f)
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)
	logf(" !!! МОНИТОРИНГ ЗАПУЩЕН !!!")
	logf("лог-файл: %s", logPath)
	logf("проверка каждые: %v", interval)

	// путь к веб-приложению (там же, где и monitor)
	appPath := filepath.Join(filepath.Dir(exe), appBinaryName)
	if runtime.GOOS == "windows" {
		appPath += ".exe"
	}

	// если бинарника нет, то собираем его из web.go (в той же папке)
	if _, err := os.Stat(appPath); err != nil {
		logf("бинарник %s не найден -> собираем из web.go", appBinaryName)
		if err := buildWeb(filepath.Dir(exe)); err != nil {
			log.Fatal("не удалось собрать web:", err)
		}
		logf("web успешно собран -> %s", appPath)
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		if isAppRunning() {
			if isHealthy() {
				logf("OK: приложение отвечает 200 на %s", healthURL)
			} else {
				logf("CRITICAL: нет ответа по HTTP -> перезапуск")
				killApp()
				time.Sleep(1 * time.Second)
				startApp(appPath)
			}
		} else {
			logf("WARNING: процесс %s не найден -> запуск", appBinaryName)
			startApp(appPath)
		}
		<-ticker.C
	}
}

// cборка web.go в указанной директории
func buildWeb(dir string) error {
	out := appBinaryName
	if runtime.GOOS == "windows" {
		out += ".exe"
	}
	cmd := exec.Command("go", "build", "-o", out, "web.go")
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// проверка: запущен ли процесс web
func isAppRunning() bool {
	// windows
	if runtime.GOOS == "windows" {
		// tasklist возвращает 0, если процесс найден
		return exec.Command("tasklist", "/NH", "/FI", fmt.Sprintf("IMAGENAME eq %s.exe", appBinaryName)).Run() == nil
	}
	// linux / macOS
	return exec.Command("pgrep", "-x", appBinaryName).Run() == nil
}

// проверка доступности по HTTP
func isHealthy() bool {
	client := &http.Client{Timeout: timeout}
	resp, err := client.Get(healthURL)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

// убиваем процесс (если завис)
func killApp() {
	if runtime.GOOS == "windows" {
		exec.Command("taskkill", "/F", "/IM", appBinaryName+".exe").Run()
	} else {
		exec.Command("pkill", "-x", appBinaryName).Run()
	}
}

// запуск приложения в фоне
func startApp(path string) {
	cmd := exec.Command(path)
	cmd.Stdout = log.Writer()
	cmd.Stderr = log.Writer()
	if err := cmd.Start(); err != nil {
		logf("ОШИБКА запуска %s: %v", path, err)
		return
	}
	logf("приложение запущено (PID %d)", cmd.Process.Pid)
}
