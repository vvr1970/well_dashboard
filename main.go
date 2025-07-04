package main

import (
	"html/template"
	"log"
	"net/http"
	"strconv"
	"sync"
)

type Well struct {
	ID     int
	Name   string
	Field  string
	Status string // "active", "repair", "conservation"
}

type Stats struct {
	Active       int
	Repair       int
	Conservation int
}

type PageData struct {
	Title         string
	Wells         []Well
	ShowAddButton bool
	ActiveTab     string
	Stats         Stats
}

var (
	data     PageData
	dataLock sync.Mutex
)

func main() {
	// Инициализация тестовых данных
	data = PageData{
		Title: "Управление скважинами",
		Wells: []Well{
			{1, "Скважина-1", "Месторождение-1", "active"},
			{2, "Скважина-2", "Месторождение-2", "repair"},
			{3, "Скважина-3", "Месторождение-1", "conservation"},
		},
		ShowAddButton: true,
	}
	// Статические файлы
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// Маршруты
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/add", addWellHandler)
	http.HandleFunc("/delete/", deleteWellHandler)

	log.Println("Сервер запущен на http://localhost:8081")
	log.Fatal(http.ListenAndServe(":8081", nil))
}

// Основноф Хандлер
func homeHandler(w http.ResponseWriter, r *http.Request) {
	dataLock.Lock()
	defer dataLock.Unlock()

	tmpl := template.Must(template.ParseGlob("templates/*.html"))
	err := tmpl.ExecuteTemplate(w, "base", data)
	if err != nil {
		log.Printf("Ошибка рендеринга: %v", err)
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
	}
}

// -/-add
func addWellHandler(w http.ResponseWriter, r *http.Request) {
	dataLock.Lock()
	defer dataLock.Unlock()

	if r.Method == "POST" {
		// Обработка отправки формы
		id := len(data.Wells) + 1
		name := r.FormValue("name")
		field := r.FormValue("field")
		status := r.FormValue("status")

		data.Wells = append(data.Wells, Well{
			ID:     id,
			Name:   name,
			Field:  field,
			Status: status,
		})
		// В обработчике /:
		data.Stats = calculateStats(data.Wells)
		data.ActiveTab = "dashboard"

		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// Показ формы добавления
	data.ShowAddButton = false
	tmpl := template.Must(template.ParseGlob("templates/*.html"))
	err := tmpl.ExecuteTemplate(w, "base", data)
	if err != nil {
		log.Printf("Ошибка рендеринга: %v", err)
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
	}
}

func deleteWellHandler(w http.ResponseWriter, r *http.Request) {
	dataLock.Lock()
	defer dataLock.Unlock()

	idStr := r.URL.Path[len("/delete/"):]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Неверный ID", http.StatusBadRequest)
		return
	}

	// Удаление скважины
	for i, well := range data.Wells {
		if well.ID == id {
			data.Wells = append(data.Wells[:i], data.Wells[i+1:]...)
			break
		}
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func renderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
	t, err := template.ParseFiles(
		"templates/base.html",
		"templates/header.html",
		"templates/sidebar.html",
		"templates/"+tmpl+".html",
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = t.ExecuteTemplate(w, "base", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// Новая функция:
func calculateStats(wells []Well) Stats {
	var stats Stats
	for _, w := range wells {
		switch w.Status {
		case "active":
			stats.Active++
		case "repair":
			stats.Repair++
		case "conservation":
			stats.Conservation++
		}
	}
	return stats
}
