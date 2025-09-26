package routes

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/felipemendozag/api-sis-operativos/db"
	"github.com/felipemendozag/api-sis-operativos/middleware"
	"github.com/felipemendozag/api-sis-operativos/models"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"

	"github.com/golang-jwt/jwt/v5"
)

var jwtKey = []byte("api-key-sis-operativo")

type DateFilter struct {
	DateStart string `json:"date_start"`
	DateEnd   string `json:"date_end"`
}

func UserRoute(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello, User!"))
}

func UpdateUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(r)
	id := vars["id"]
	var user models.User
	if err := db.DB.First(&user, id).Error; err != nil {
		http.Error(w, `{"error": "User not found"}`, http.StatusNotFound)
		return
	}
	var input models.User
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, `{"error": "Invalid request body"}`, http.StatusBadRequest)
		return
	}
	if input.FirstName != "" {
		user.FirstName = input.FirstName
	}
	if input.LastName != "" {
		user.LastName = input.LastName
	}
	if input.Email != "" {
		user.Email = input.Email
	}
	if input.Password != "" {
		hashedPassword, errCrypt := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
		if errCrypt != nil {
			http.Error(w, `{"error": "Error hashing password"}`, http.StatusInternalServerError)
			return
		}
		user.PasswordHash = string(hashedPassword)
	}
	user.Status = true
	if input.RoleID != 0 {
		user.RoleID = input.RoleID
	}
	if err := db.DB.Save(&user).Error; err != nil {
		http.Error(w, `{"error": "Error updating user"}`, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(&user)
}

func ChangeStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var user models.User
	params := mux.Vars(r)
	id := params["id"]
	if err := db.DB.First(&user, id).Error; err != nil {
		http.Error(w, `{"error": "Usuario no encontrado"}`, http.StatusNotFound)
		return
	}
	user.Status = false
	if err := db.DB.Save(&user).Error; err != nil {
		http.Error(w, `{"error": "Error al actualizar el usuario"}`, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}

func GetUsersHandler(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value(middleware.UserContextKey).(jwt.MapClaims)
	roleID := claims["role_id"]
	var users []models.User
	db.DB.Where("role_id <> ?", roleID).Where("status = ?", true).Find(&users)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(users)
}

func GetUserHandler(w http.ResponseWriter, r *http.Request) {
	var user models.User
	params := mux.Vars(r)
	db.DB.First(&user, params["id"])
	if user.ID == 0 {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode("User not found")
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}

func CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var user models.User

	json.NewDecoder(r.Body).Decode(&user)

	hashedPassword, errCrypt := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if errCrypt != nil {
		http.Error(w, `{"error": "Error hashing password"}`, http.StatusInternalServerError)
		return
	}

	user.PasswordHash = string(hashedPassword)
	user.Status = true
	user.RoleID = 1
	createUser := db.DB.Create(&user)
	err := createUser.Error
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Error creating user"))
		json.NewEncoder(w).Encode(err)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(&user)
}

func LoginUserHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var req models.User
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid request payload"}`, http.StatusBadRequest)
		return
	}
	var user models.User
	if err := db.DB.Preload("Role").Preload("Role.Permissions").
		Where("email = ?", req.Email).
		First(&user).Error; err != nil {
		http.Error(w, `{"error": "User not found"}`, http.StatusUnauthorized)
		return
	}
	/* if err := db.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		http.Error(w, `{"error": "User not found"}`, http.StatusUnauthorized)
		return
	} */
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		http.Error(w, `{"error": "Invalid credentials"}`, http.StatusUnauthorized)
		return
	}
	// 4. Generar JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"role":    user.Role.Name,
		"role_id": user.Role.ID,
		"exp":     time.Now().Add(time.Hour * 72).Unix(),
	})
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		http.Error(w, `{"error": "Error generating token"}`, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"user": map[string]interface{}{
			"id":    user.ID,
			"email": user.Email,
			"role": map[string]interface{}{
				"id":   user.Role.ID,
				"name": user.Role.Name,
				"permissions": func() []string {
					var perms []string
					for _, p := range user.Role.Permissions {
						perms = append(perms, p.Action)
					}
					return perms
				}(),
			},
		},
		"token": tokenString,
	})
}

func GetAttendanceUsersReports(w http.ResponseWriter, r *http.Request) {

	type AttendanceResponse struct {
		FirstName      string    `json:"first_name"`
		LastName       string    `json:"last_name"`
		ID             int       `json:"id"`
		Email          string    `json:"email"`
		Status         string    `json:"status"`
		DateAttendance time.Time `json:"date_attendance"`
		RoleName       string    `json:"role_name"`
	}

	var results []AttendanceResponse

	query := `
        SELECT u.first_name, u.last_name, u.id, u.email,
               CASE WHEN a.id IS NULL THEN 'falto' ELSE a.status END AS status,
               CASE WHEN a.date IS NULL THEN NOW() ELSE a.date END AS date_attendance,
               r.name as role_name
        FROM attendances a
        RIGHT JOIN users u ON u.id = a.user_id
        INNER JOIN roles r ON r.id = u.role_id
        WHERE u.status = true AND u.role_id = 1
    `

	if err := db.DB.Raw(query).Scan(&results).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)

}

func GetViewAllAttendances(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var filter DateFilter

	if err := json.NewDecoder(r.Body).Decode(&filter); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	dateStart := filter.DateStart + " 00:00:00"
	dateEnd := filter.DateEnd + " 23:59:59"

	type AttendanceResponse struct {
		FirstName      string    `json:"first_name"`
		LastName       string    `json:"last_name"`
		ID             int       `json:"id"`
		Email          string    `json:"email"`
		Status         string    `json:"status"`
		DateAttendance time.Time `json:"date_attendance"`
		RoleName       string    `json:"role_name"`
	}

	var results []AttendanceResponse

	query := `
        select
		u.first_name , u.last_name , u.id , u.email,
		CASE 
				WHEN a.id is null THEN 'falto'
				ELSE a.status 
			END AS status,
		case 
			when a.date is null then now()
			else a.date
		end as date_attendance, r."name" as role_name
		from attendances a
		right join users u on u.id = a.user_id
		inner join roles r on r.id = u.role_id 
		where u.status = true and u.role_id = 1
		and "date" between ? and ?
    `
	if err := db.DB.Raw(query, dateStart, dateEnd).Scan(&results).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if len(results) == 0 {
		json.NewEncoder(w).Encode(map[string]any{})
		return
	}
	json.NewEncoder(w).Encode(results)
}
