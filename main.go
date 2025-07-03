package main

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

// Модель скважины
type GasWell struct {
	ID         int
	Name       string
	Location   string
	Production float64
	Status     string
}

var db *sql.DB
var tmpl *template.Template

func init() {
	// Подключение к PostgreSQL
	var err error
	connStr := "user=post password=post dbname=wells_db sslmode=disable"
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	// Загрузка шаблонов
	tmpl = template.Must(template.ParseGlob("templates/*.html"))
}

func main() {
	r := mux.NewRouter()

	// Статические файлы (графики)
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Роуты
	r.HandleFunc("/", listWells).Methods("GET")
	r.HandleFunc("/create", createWellForm).Methods("GET")
	r.HandleFunc("/create", createWell).Methods("POST")
	r.HandleFunc("/edit/{id}", editWellForm).Methods("GET")
	r.HandleFunc("/edit/{id}", editWell).Methods("POST")
	r.HandleFunc("/delete/{id}", deleteWell).Methods("GET")
	r.HandleFunc("/view/{id}", viewWell).Methods("GET") // Важно!

	log.Println("Server started on :8081")
	http.ListenAndServe(":8081", r)
}

// Список всех скважин
func listWells(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT id, name, location, production, status FROM gas_wells")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var wells []GasWell
	for rows.Next() {
		var well GasWell
		err := rows.Scan(&well.ID, &well.Name, &well.Location, &well.Production, &well.Status)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		wells = append(wells, well)
	}

	tmpl.ExecuteTemplate(w, "index.html", wells)
}

// Форма добавления
func createWellForm(w http.ResponseWriter, r *http.Request) {
	tmpl.ExecuteTemplate(w, "create.html", nil)
}

// Добавление скважины
func createWell(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	location := r.FormValue("location")
	production, _ := strconv.ParseFloat(r.FormValue("production"), 64)
	status := r.FormValue("status")

	_, err := db.Exec(
		"INSERT INTO gas_wells (name, location, production, status) VALUES ($1, $2, $3, $4)",
		name, location, production, status,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// Форма редактирования
func editWellForm(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])

	var well GasWell
	err := db.QueryRow("SELECT id, name, location, production, status FROM gas_wells WHERE id = $1", id).Scan(
		&well.ID, &well.Name, &well.Location, &well.Production, &well.Status,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl.ExecuteTemplate(w, "edit.html", well)
}

// Обновление данных
func editWell(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])

	name := r.FormValue("name")
	location := r.FormValue("location")
	production, _ := strconv.ParseFloat(r.FormValue("production"), 64)
	status := r.FormValue("status")

	_, err := db.Exec(
		"UPDATE gas_wells SET name=$1, location=$2, production=$3, status=$4 WHERE id=$5",
		name, location, production, status, id,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// Удаление скважины
func deleteWell(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])

	_, err := db.Exec("DELETE FROM gas_wells WHERE id = $1", id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// Просмотр деталей скважины
func viewWell(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		log.Println("Ошибка парсинга ID:", err)
		http.Error(w, "Некорректный ID", http.StatusBadRequest)
		return
	}

	var well GasWell
	err = db.QueryRow("SELECT id, name, location, production, status FROM gas_wells WHERE id = $1", id).Scan(
		&well.ID, &well.Name, &well.Location, &well.Production, &well.Status,
	)
	if err != nil {
		log.Println("Ошибка запроса к БД:", err) // Логируем ошибку
		http.Error(w, "Скважина не найдена", http.StatusNotFound)
		return
	}

	// Логируем полученные данные
	log.Printf("Данные скважины: %+v\n", well)

	// Проверяем, что шаблон существует
	tmpl, err := template.ParseFiles("templates/view.html")
	if err != nil {
		log.Println("Ошибка загрузки шаблона:", err)
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}

	// Рендерим шаблон
	err = tmpl.Execute(w, well)
	if err != nil {
		log.Println("Ошибка рендеринга шаблона:", err)
		return
	}
}
