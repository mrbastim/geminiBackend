package main

import (
	"geminiBackend/config"
	"geminiBackend/internal/handlers"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	// Настройка логгера: дата, время с микросекундами и короткое имя файла в выводе
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)
	log.SetPrefix("[gemini] ")
	log.Println("Запуск сервера...")
	r := mux.NewRouter()

	// Настройка маршрутов
	r.HandleFunc("/items", handlers.GetOptions).Methods("GET")
	r.HandleFunc("/items", handlers.PostResponse).Methods("POST")

	//Читаем конфигурацию
	config, err := config.LoadConfig("config/config.json")
	if err != nil {
		log.Fatalf("Ошибка при загрузке конфигурации: %v", err)
	}

	// Запуск HTTP-сервера
	log.Println("Сервер запущен на порту", config.Port)
	if err := http.ListenAndServe(":"+config.Port, r); err != nil {
		log.Fatalf("Ошибка при запуске сервера: %v", err)
	}
}
