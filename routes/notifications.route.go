package routes

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/felipemendozag/api-sis-operativos/db"
	"github.com/felipemendozag/api-sis-operativos/middleware"
	"github.com/felipemendozag/api-sis-operativos/models"
	"github.com/golang-jwt/jwt/v5"
)

func GetAllNotificationsForUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	claims := r.Context().Value(middleware.UserContextKey).(jwt.MapClaims)
	// Obtener el userId desde la URL
	userID := claims["user_id"]

	var notifications []models.Notification

	// Consultar notificaciones por usuario
	if err := db.DB.Where("user_id = ?", userID).Order("created_at DESC").Find(&notifications).Error; err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Error al obtener notificaciones"})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(notifications)
}
func GetUnReadNotification(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	claims := r.Context().Value(middleware.UserContextKey).(jwt.MapClaims)
	// Obtener el userId desde la URL
	userID := claims["user_id"]

	var count int64

	// Contar solo notificaciones no leídas
	if err := db.DB.Model(&models.Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Count(&count).Error; err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Error al contar notificaciones no leídas",
		})
		return
	}

	// Respuesta
	response := map[string]interface{}{
		"user_id":      userID,
		"unread_count": count,
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
func MarkReadNotification(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Estructura para recibir el body
	type MarkReadRequest struct {
		UserID         uint `json:"user_id"`
		NotificationID uint `json:"notification_id"`
	}

	var req MarkReadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid request payload"}`, http.StatusBadRequest)
		return
	}

	// Buscar la notificación
	var notif models.Notification
	if err := db.DB.Where("id = ? AND user_id = ?", req.NotificationID, req.UserID).First(&notif).Error; err != nil {
		http.Error(w, `{"error": "Notification not found"}`, http.StatusNotFound)
		return
	}

	// Actualizar valores
	now := time.Now()
	notif.IsRead = true
	notif.ReadAt = &now

	if err := db.DB.Save(&notif).Error; err != nil {
		http.Error(w, `{"error": "Failed to update notification"}`, http.StatusInternalServerError)
		return
	}

	// Respuesta OK
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":      "Notification marked as read",
		"notification": notif,
	})
}
