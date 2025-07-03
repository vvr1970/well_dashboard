package handlers

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"time"
	"wells-app/models"

	"github.com/gorilla/mux"
)

type WellHandlers struct {
	DB        *sql.DB
	Templates map[string]*template.Template
}

func NewWellHandlers(db *sql.DB) *WellHandlers {
	// Инициализация шаблонов
	templates := make(map[string]*template.Template)
	templates["list"] = template.Must(template.ParseFiles(
		"templates/base.html",
		"templates/list.html",
	))
	templates["view"] = template.Must(template.ParseFiles(
		"templates/base.html",
		"templates/view.html",
	))
	templates["edit"] = template.Must(template.ParseFiles(
		"templates/base.html",
		"templates/edit.html",
	))
	// В функции NewWellHandlers добавить:
	templates["create"] = template.Must(template.ParseFiles(
		"templates/base.html",
		"templates/create.html",
	))

	return &WellHandlers{
		DB:        db,
		Templates: templates,
	}
}

func (h *WellHandlers) ListWells(w http.ResponseWriter, r *http.Request) {
	wells, err := models.AllWells(h.DB)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := h.Templates["list"].ExecuteTemplate(w, "base", wells); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *WellHandlers) ViewWell(w http.ResponseWriter, r *http.Request) {
	// Получаем ID из URL с помощью mux.Vars
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid well ID", http.StatusBadRequest)
		return
	}

	well, err := models.GetWell(h.DB, id)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Well not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	data := map[string]interface{}{
		"Well":                  well,
		"DrillingDateFormatted": well.DrillingDate.Format("2006-01-02"),
	}

	if err := h.Templates["view"].ExecuteTemplate(w, "base", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *WellHandlers) EditWell(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid well ID", http.StatusBadRequest)
		return
	}
	well, err := models.GetWell(h.DB, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	data := struct {
		Well     *models.Well
		Statuses []string
	}{
		Well:     well,
		Statuses: []string{"active", "inactive", "maintenance", "abandoned"},
	}

	if err := h.Templates["edit"].ExecuteTemplate(w, "base", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *WellHandlers) UpdateWell(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid well ID", http.StatusBadRequest)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	well := models.Well{
		ID:           id,
		Name:         r.FormValue("name"),
		Depth:        parseFloat(r.FormValue("depth")),
		Location:     r.FormValue("location"),
		Status:       r.FormValue("status"),
		Productivity: parseFloat(r.FormValue("productivity")),
		Field:        r.FormValue("field"),
		Operator:     r.FormValue("operator"),
	}

	if err := models.UpdateWell(h.DB, &well); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/wells/%d", id), http.StatusSeeOther)
}

func parseFloat(s string) float64 {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return f
}

// ********************************************
// CreateWellForm отображает форму создания скважины
func (h *WellHandlers) CreateWellForm(w http.ResponseWriter, r *http.Request) {
	data := struct {
		Statuses  []string
		CSRFToken string
	}{
		Statuses:  []string{"active", "inactive", "maintenance", "abandoned"},
		CSRFToken: "generate-real-csrf-token-in-production", // TODO: заменить на реальную генерацию
	}

	if err := h.Templates["create"].ExecuteTemplate(w, "base", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// CreateWell обрабатывает отправку формы создания скважины
func (h *WellHandlers) CreateWell(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	drillingDate, _ := time.Parse("2006-01-02", r.FormValue("drilling_date"))
	well := models.Well{
		Name:         r.FormValue("name"),
		Depth:        parseFloat(r.FormValue("depth")),
		Location:     r.FormValue("location"),
		Status:       r.FormValue("status"),
		Productivity: parseFloat(r.FormValue("productivity")),
		DrillingDate: drillingDate,
		Field:        r.FormValue("field"),
		Operator:     r.FormValue("operator"),
	}

	log.Printf("Creating well: %+v", well) // Выведет все поля структуры

	if err := models.CreateWell(h.DB, &well); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/wells/%d", well.ID), http.StatusSeeOther)
}
