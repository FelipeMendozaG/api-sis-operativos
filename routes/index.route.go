package routes

import (
	"net/http"
)

func HomeRoute(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello, World! 2w"))
}
