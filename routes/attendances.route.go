package routes

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/felipemendozag/api-sis-operativos/db"
	"github.com/felipemendozag/api-sis-operativos/models"
	"github.com/gorilla/mux"
)

func GetAttendanceHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var attendances []models.Attendance
	params := mux.Vars(r)
	fmt.Println(params)
	db.DB.Where("user_id = ?", params["user_id"]).Find(&attendances)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(attendances)
}

func CreateAttendanceHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var attendance models.Attendance
	if err := json.NewDecoder(r.Body).Decode(&attendance); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	location, _ := time.LoadLocation("America/Lima")

	attendanceDate := attendance.Date.In(location)
	// Generamos la fecha en backend
	now := time.Now().In(location)
	attendance.Date = now
	// Definir hora de inicio (9:00 AM) y reglas
	startTime := time.Date(now.Year(), now.Month(), now.Day(), 9, 0, 0, 0, location)
	earlyLimit := startTime.Add(-1 * time.Hour) // 8:00 am
	tolerance := startTime.Add(5 * time.Minute) // 9:05 am

	// Reglas de validación
	if attendanceDate.Before(earlyLimit) {
		http.Error(w, "No puedes marcar antes de las 8:00 am", http.StatusForbidden)
		return
	}

	if attendanceDate.After(tolerance) {
		attendance.Status = "tardanza"
	} else {
		attendance.Status = "puntual"
	}

	// Guardar en DB
	if err := db.DB.Create(&attendance).Error; err != nil {
		http.Error(w, "Error creating attendance", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(attendance)
}

func GetAttendancesHandler(w http.ResponseWriter, r *http.Request) {

}
