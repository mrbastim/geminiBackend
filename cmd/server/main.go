package main

import (
	"geminiBackend/config"
	"geminiBackend/internal/handlers"
	"geminiBackend/internal/middleware"
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
	//Читаем конфигурацию
	config, err := config.LoadConfig("config/config.json")
	if err != nil {
		log.Fatalf("Ошибка при загрузке конфигурации: %v", err)
	}

	// Публичный маршрут для авторизации
	r.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Handling /login request")
		handlers.Login(w, r, config)
	}).Methods("POST")

	// Защищенные маршруты
	// /options - только для админов
	adminRouter := r.PathPrefix("/options").Subrouter()
	adminRouter.Use(middleware.JWTAuth(config.JWTSecret))
	adminRouter.Use(middleware.RequireAdmin)
	adminRouter.HandleFunc("", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Handling /options request")
		handlers.GetOptions(w, r, config)
	}).Methods("GET")

	// /items - для всех авторизованных пользователей
	authRouter := r.PathPrefix("/items").Subrouter()
	authRouter.Use(middleware.JWTAuth(config.JWTSecret))
	authRouter.HandleFunc("", handlers.PostResponse).Methods("POST")

	// Запуск HTTP-сервера
	log.Println("Сервер запущен на порту", config.Port)
	if err := http.ListenAndServe(":"+config.Port, r); err != nil {
		log.Fatalf("Ошибка при запуске сервера: %v", err)
	}
}
