# Hello World! с мониторингом и автозапуском

простое кроссплатформенное решение на go:
1. веб-сервер возвращает «hello world!» на http://127.0.0.1:8080
2. мониторинг каждые 10 секунд проверяет доступность и перезапускает при сбоях
3. одна команда установки + автозапуск при старте системы на linux, macos и windows

## структура проекта
web.go         -> исходник веб-сервера
monitor.go     -> мониторинг и перезапуск
install.go     -> установщик + настройка автозапуска
hello-world/   -> создаётся автоматически: готовые бинарники и лог
readme.md      -> этот файл

## команды управления

| действие                     | linux                                                                 | macos                                                                                     | windows (powershell / cmd)                                                                 |
|------------------------------|-----------------------------------------------------------------------|-------------------------------------------------------------------------------------------|--------------------------------------------------------------------------------------------|
| установка                    | `go run install.go`                                                   | `go run install.go`                                                                       | `go run install.go`                                                                        |
| проверка работы              | `curl http://127.0.0.1:8080`                                          | `curl http://127.0.0.1:8080`                                                              | `curl http://127.0.0.1:8080` или открыть в браузере                                        |
| логи в реальном времени      | `tail -f hello-world/monitor.log`                                     | `tail -f hello-world/monitor.log`                                                         | `Get-Content -Path hello-world\monitor.log -Wait` (ps)<br>`type hello-world\monitor.log` (cmd) |
| запустить мониторинг вручную | `./hello-world/monitor`                                               | `./hello-world/monitor`                                                                   | `hello-world\monitor.exe`                                                                  |
| остановить быстро            | `pkill -f monitor`                                                    | `pkill -f monitor`                                                                        | `taskkill /f /im monitor.exe`                                                              |
| остановить через систему     | `systemctl --user stop hello-monitor.service`                        | `launchctl unload ~/Library/LaunchAgents/local.helloworld.monitor.plist`                 | удалить `HelloWorld-Monitor.bat` из автозагрузки                                           |
| статус сервиса               | `systemctl --user status hello-monitor.service`                       | `launchctl list \| grep helloworld`                                                       | —                                                                                          |
| полностью убрать автозапуск  | `systemctl --user disable --now hello-monitor.service`<br>`rm ~/.config/systemd/user/hello-monitor.service` | `launchctl unload ~/Library/LaunchAgents/local.helloworld.monitor.plist`<br>`rm ~/Library/LaunchAgents/local.helloworld.monitor.plist` | удалить bat-файл из `%APPDATA%\Microsoft\Windows\Start Menu\Programs\Startup`             |
| удалить всё                  | `rm -rf hello-world`                                                  | `rm -rf hello-world`                                                                      | `rmdir /s /q hello-world`                                                                  |


## настраиваемые параметры (monitor.go)

const (
	// имя исполняемого файла веб-приложения (без .exe на windows - добавляется автоматически)
	appBinaryName = "web"

	// url, по которому мониторим провер проверять здоровье приложения
	healthURL = "http://127.0.0.1:8080"

	// интервал между проверками
	interval = 10 * time.Second

	// таймаут http-запроса при проверке доступности
	timeout = 3 * time.Second
)
