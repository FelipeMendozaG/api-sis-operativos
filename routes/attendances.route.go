package routes

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/felipemendozag/api-sis-operativos/db"
	"github.com/felipemendozag/api-sis-operativos/models"
	"github.com/gorilla/mux"
	"github.com/xuri/excelize/v2"
	gomail "gopkg.in/gomail.v2"
)

type AttendanceUser struct {
	FirstName      string    `json:"first_name"`
	LastName       string    `json:"last_name"`
	ID             int       `json:"id"`
	Email          string    `json:"email"`
	Status         string    `json:"status"`
	DateAttendance time.Time `json:"date_attendance"`
	RoleName       string    `json:"role_name"`
}

func GetAttendanceHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var attendances []models.Attendance
	params := mux.Vars(r)
	fmt.Println(params)
	db.DB.Where("user_id = ?", params["user_id"]).Order("created_at DESC").Find(&attendances)
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

	/* attendanceDate := attendance.Date.In(location) */
	now := time.Now().In(location)
	attendance.Date = now
	// Definir hora de inicio (9:00 AM) y reglas
	startTime := time.Date(now.Year(), now.Month(), now.Day(), 9, 0, 0, 0, location)
	earlyLimit := startTime.Add(-1 * time.Hour)
	tolerance := startTime.Add(5 * time.Minute)

	// VALIDAMOS PARA QUE SOLO EL USUARIO PUEDA MARCAR UNA ASISTENCIA POR DIA
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, location)
	endOfDay := startOfDay.Add(24 * time.Hour)
	var existing models.Attendance
	if err := db.DB.Where("user_id = ? AND date >= ? AND date < ? and status not in('falto')",
		attendance.UserID, startOfDay, endOfDay).
		First(&existing).Error; err == nil {

		w.WriteHeader(http.StatusConflict) // 409
		json.NewEncoder(w).Encode(map[string]string{
			"message_error": "Ya registraste tu asistencia hoy",
		})
		return
	}

	//Reglas de validación
	if now.Before(earlyLimit) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)

		json.NewEncoder(w).Encode(map[string]string{
			"message_error": "No puedes marcar antes de las 8:00 am",
		})
		return
	}

	if now.After(tolerance) {
		attendance.Status = "tardanza"
	} else {
		attendance.Status = "puntual"
	}
	err := db.DB.Where("user_id = ? AND date >= ? AND date < ? AND status in('falto')",
		attendance.UserID, startOfDay, endOfDay).
		First(&existing).Error
	if err == nil {
		// Si existe, actualizar
		existing.Status = attendance.Status
		existing.Date = attendance.Date
		if err := db.DB.Save(&existing).Error; err != nil {
			http.Error(w, `{"error":"Error updating attendance"}`, http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Asistencia actualizada correctamente",
		})
		return
	}

	// Guardar en DB
	// Si no existe, crear nuevo
	if err := db.DB.Create(&attendance).Error; err != nil {
		http.Error(w, `{"error":"Error creating attendance"}`, http.StatusInternalServerError)
		return
	}

	sendMailHandler()
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(attendance)
}

func GetAttendancesHandler(w http.ResponseWriter, r *http.Request) {

}

func SendAlert(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var user models.User
	type UpdateAttendanceRequest struct {
		ID     uint   `json:"id"`
		Status string `json:"status"`
	}
	var req UpdateAttendanceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid request payload"}`, http.StatusBadRequest)
		return
	}
	db.DB.First(&user, req.ID)
	if user.ID == 0 {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode("User not found")
		return
	}

	// registramos la notificacion
	var notif models.Notification
	if req.Status == "tardanza" {
		notif = models.Notification{
			UserID:  user.ID,
			Title:   "NOTIFICACIÓN DE TARDANZA",
			Message: "Se está enviando este mensaje por una notificación de tardanza. Marque su asistencia a tiempo, por favor.",
			Type:    "tardanza",
		}
	} else {
		notif = models.Notification{
			UserID:  user.ID,
			Title:   "NOTIFICACIÓN DE FALTA",
			Message: "Se está enviando este mensaje por una notificación de falta en el trabajo. Por favor, comuníquese con el supervisor.",
			Type:    "falta",
		}
	}
	// Guardar en la BD
	if err := db.DB.Create(&notif).Error; err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Error al guardar la notificación"})
		return
	}
	//

	if req.Status == "tardanza" {
		sendMailAlert(user.Email, "NOTIFICACION DE TARDANZA", "SE ESTA ENVIANDO ESTE MENSAJE POR UNA NOTIFICACION DE TARDANZA, MARQUE SU ASISTENCIA A TIEMPO POR FAVOR")
	} else {
		sendMailAlert(user.Email, "NOTIFICACION DE FALTAS", "SE ESTA ENVIANDO ESTE MENSAJE POR UNA NOTIFICACION DE FALTA EN EL TRABAJO. POR FAVOR, COMUNICARSE CON EL SUPERVISOR")
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}

func sendMailHandler() {
	m := gomail.NewMessage()
	m.SetHeader("From", "no-reply@felipemendozag.com")
	m.SetHeader("To", "felipe188.mendoza@gmail.com")
	m.SetHeader("Subject", "USUARIO MARCO ASISTENCIA:")
	m.SetBody("text/plain", "El usuario ha registrado su asistencia correctamente.")

	d := gomail.NewDialer(
		"mail.felipemendozag.com",
		25,
		"",
		"",
	)

	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	if err := d.DialAndSend(m); err != nil {
		log.Printf("Error al enviar correo: %s", err)
		// http.Error(w, "No se pudo enviar el correo", http.StatusInternalServerError)
		return
	}
	log.Printf("SE ENVIO EL CORREO")
}

func sendMailAlert(email string, subject string, text string) {
	m := gomail.NewMessage()
	m.SetHeader("From", "no-reply@felipemendozag.com") // Cambia por tu dominio
	m.SetHeader("To", email)
	m.SetHeader("Subject", subject)
	m.SetBody("text/plain", text)

	d := gomail.NewDialer(
		"mail.felipemendozag.com",
		25,
		"",
		"",
	)

	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	if err := d.DialAndSend(m); err != nil {
		log.Printf("Error al enviar correo: %s", err)
		return
	}
	log.Printf("SE ENVIO EL CORREO")
}

func DownloadAttendancesExcelHandler(w http.ResponseWriter, r *http.Request) {
	var filter DateFilter
	if err := json.NewDecoder(r.Body).Decode(&filter); err != nil {
		http.Error(w, `{"error":"invalid payload"}`, http.StatusBadRequest)
		return
	}
	if filter.DateStart == "" || filter.DateEnd == "" {
		http.Error(w, `{"error":"date_start and date_end are required"}`, http.StatusBadRequest)
		return
	}
	dateStart := filter.DateStart + " 00:00:00"
	dateEnd := filter.DateEnd + " 23:59:59"
	query := `
		SELECT
			u.first_name, u.last_name, u.id, u.email,
			CASE 
				WHEN a.id IS NULL THEN 'falto'
				ELSE a.status 
			END AS status,
			CASE 
				WHEN a.date IS NULL THEN NOW()
				ELSE a.date
			END AS date_attendance,
			r."name" AS role_name
		FROM attendances a
		RIGHT JOIN users u ON u.id = a.user_id
		INNER JOIN roles r ON r.id = u.role_id 
		WHERE u.status = true
		  AND u.role_id = 1
		  AND a.date BETWEEN ? AND ?
	`

	var results []AttendanceUser
	if err := db.DB.Raw(query, dateStart, dateEnd).Scan(&results).Error; err != nil {
		http.Error(w, `{"error":"db query error"}`, http.StatusInternalServerError)
		return
	}
	f := excelize.NewFile()
	sheet := "Reporte"
	f.SetSheetName("Sheet1", sheet)

	headers := []string{"ID", "Nombres", "Apellidos", "Email", "Rol", "Status", "Fecha asistencia"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, h)
	}

	// Contenido
	for rowIdx, item := range results {
		r := rowIdx + 2
		f.SetCellValue(sheet, fmt.Sprintf("A%d", r), item.ID)
		f.SetCellValue(sheet, fmt.Sprintf("B%d", r), item.FirstName)
		f.SetCellValue(sheet, fmt.Sprintf("C%d", r), item.LastName)
		f.SetCellValue(sheet, fmt.Sprintf("D%d", r), item.Email)
		f.SetCellValue(sheet, fmt.Sprintf("E%d", r), item.RoleName)
		f.SetCellValue(sheet, fmt.Sprintf("F%d", r), item.Status)

		if !item.DateAttendance.IsZero() {
			f.SetCellValue(sheet, fmt.Sprintf("G%d", r), item.DateAttendance.Format("2006-01-02 15:04:05"))
		} else {
			f.SetCellValue(sheet, fmt.Sprintf("G%d", r), "")
		}
	}

	f.SetColWidth(sheet, "A", "G", 18)

	filename := fmt.Sprintf("reporte_asistencias_%s_al_%s.xlsx", filter.DateStart, filter.DateEnd)
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	w.WriteHeader(http.StatusOK)

	if err := f.Write(w); err != nil {
		http.Error(w, `{"error":"error generating file"}`, http.StatusInternalServerError)
		return
	}
}

func InitDay(w http.ResponseWriter, r *http.Request) {
	query_sql := `
	insert into attendances (created_at, updated_at, user_id, "date", status)
	select now(), now(),  u2.id, now(), 'falto' from users u2
	where u2.status = true and u2.role_id = 1 and u2.id not in (select distinct user_id from attendances a where DATE(a.created_at) = DATE(now()));
	`
	if err := db.DB.Exec(query_sql).Error; err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Asistencias generadas correctamente"})
}
