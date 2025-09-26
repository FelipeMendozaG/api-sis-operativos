package main

import (
	"log"
	"net/http"

	"github.com/felipemendozag/api-sis-operativos/db"
	"github.com/felipemendozag/api-sis-operativos/middleware"
	"github.com/felipemendozag/api-sis-operativos/models"
	"github.com/felipemendozag/api-sis-operativos/routes"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

func main() {

	db.DBConnection()
	db.DB.AutoMigrate(&models.User{}, &models.Attendance{}, &models.Role{}, &models.Permission{}, &models.RoleSchedule{})
	router := mux.NewRouter()
	router.HandleFunc("/", routes.HomeRoute)
	// RUTAS DE MI API
	router.HandleFunc("/api/users", middleware.AuthMiddleware(routes.GetUsersHandler)).Methods("GET")
	router.HandleFunc("/api/user/{id}", routes.GetUserHandler).Methods("GET")
	router.HandleFunc("/api/user/{id}", middleware.AuthMiddleware(routes.UpdateUser)).Methods("PUT")
	router.HandleFunc("/api/user", middleware.AuthMiddleware(routes.CreateUserHandler)).Methods("POST")
	router.HandleFunc("/api/user/login", routes.LoginUserHandler).Methods("POST")
	// RUTAS PARA REGISTRAR MI ASISTENCIA
	router.HandleFunc("/api/attendances", middleware.AuthMiddleware(routes.GetAttendancesHandler)).Methods("GET")
	router.HandleFunc("/api/attendance/{user_id}", middleware.AuthMiddleware(routes.GetAttendanceHandler)).Methods("GET")
	router.HandleFunc("/api/attendance", middleware.AuthMiddleware(routes.CreateAttendanceHandler)).Methods("POST")
	//
	router.HandleFunc("/api/user/delete/{id}", middleware.AuthMiddleware(routes.ChangeStatus)).Methods("POST")
	router.HandleFunc("/api/user/get/attendances", middleware.AuthMiddleware(routes.GetAttendanceUsersReports)).Methods("GET")
	router.HandleFunc("/api/user/send_alert", middleware.AuthMiddleware(routes.SendAlert)).Methods("POST")
	//
	router.HandleFunc("/api/attendances/all", middleware.AuthMiddleware(routes.GetViewAllAttendances)).Methods("POST")
	router.HandleFunc("/api/attendances/export", middleware.AuthMiddleware(routes.DownloadAttendancesExcelHandler)).Methods("POST")
	// Configurar CORS
	//
	router.HandleFunc("/api/notifications", middleware.AuthMiddleware(routes.GetAllNotificationsForUser)).Methods("GET")
	router.HandleFunc("/api/notifications", middleware.AuthMiddleware(routes.MarkReadNotification)).Methods("POST")
	router.HandleFunc("/api/notifications/unread", middleware.AuthMiddleware(routes.GetUnReadNotification)).Methods("GET")
	//
	router.HandleFunc("/api/reports", middleware.AuthMiddleware(routes.GetReports)).Methods("POST")
	//
	router.HandleFunc("/api/attendances/init_day", routes.InitDay).Methods("GET")
	//
	headersOk := handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"})
	originsOk := handlers.AllowedOrigins([]string{"*"}) // "*" permite todos los orígenes
	//originsOk := handlers.AllowedOrigins([]string{"https://dev-sis-operativos.rj.r.appspot.com"})
	methodsOk := handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"})

	log.Println("Servidor corriendo en http://localhost:9090")
	log.Fatal(http.ListenAndServe(":9090", handlers.CORS(originsOk, headersOk, methodsOk)(router)))
}
