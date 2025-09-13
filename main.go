package main

import (
	"log"
	"net/http"

	"github.com/felipemendozag/api-sis-operativos/db"
	"github.com/felipemendozag/api-sis-operativos/models"
	"github.com/felipemendozag/api-sis-operativos/routes"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

func main() {

	db.DBConnection()
	db.DB.AutoMigrate(&models.User{}, &models.Attendance{})
	router := mux.NewRouter()
	router.HandleFunc("/", routes.HomeRoute)
	// RUTAS DE MI API
	router.HandleFunc("/api/users", routes.GetUsersHandler).Methods("GET")
	router.HandleFunc("/api/user/{id}", routes.GetUserHandler).Methods("GET")
	router.HandleFunc("/api/user", routes.CreateUserHandler).Methods("POST")
	router.HandleFunc("/api/user/login", routes.LoginUserHandler).Methods("POST")
	// RUTAS PARA REGISTRAR MI ASISTENCIA
	router.HandleFunc("/api/attendances", routes.GetAttendancesHandler).Methods("GET")
	router.HandleFunc("/api/attendance/{user_id}", routes.GetAttendanceHandler).Methods("GET")
	router.HandleFunc("/api/attendance", routes.CreateAttendanceHandler).Methods("POST")
	//
	// Configurar CORS
	headersOk := handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"})
	originsOk := handlers.AllowedOrigins([]string{"*"}) // "*" permite todos los orígenes
	//originsOk := handlers.AllowedOrigins([]string{"https://dev-sis-operativos.rj.r.appspot.com"})
	methodsOk := handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"})

	log.Println("Servidor corriendo en http://localhost:9090")
	log.Fatal(http.ListenAndServe(":9090", handlers.CORS(originsOk, headersOk, methodsOk)(router)))
}
