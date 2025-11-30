//веб - приложение, которое возвращает "Hello World!"

package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

func main() {
	// регистрируем обработчик для пути /
	http.HandleFunc("/", helloHandler)

	// помощь пользователям
	log.Println("сервер запущен")
	log.Println("откройте в браузере: http://localhost:8080")
	log.Println("для остановки сервера нажмите ctrl+c (на macOS command+c)")

	// создаём сервер с минимально необходимым таймаутом
	server := &http.Server{
		Addr:              ":8080",
		ReadHeaderTimeout: 2 * time.Second, // защита от медленных клиентов
	}

	// запуск. если сервер упадет, то ошибка покажется
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Printf("сервер остановлен с ошибкой: %v\n", err)
	}
}

// функция-обработчик запросов
func helloHandler(w http.ResponseWriter, r *http.Request) {
	// разрешаю только GET-запросы
	if r.Method != http.MethodGet {
		http.Error(w, "метод не разрешён", http.StatusMethodNotAllowed)
		return
	}

	// правильный тип содержимого
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	// отправка ответа (возвращение "Hello World!")
	fmt.Fprint(w, "Hello World!\n")
}
