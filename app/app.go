package app

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"time"

	"github.com/amoliyer80/PacketRun/model"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

type App struct {
	Router *mux.Router
	DB     *sql.DB
	User   model.User
}

func (a *App) Initialize(dbuser, dbpassword, dbname string, user model.User) {
	connectionString := fmt.Sprintf("user=%s password=%s dbname=%s", dbuser, dbpassword, dbname)
	var err error
	a.DB, err = sql.Open("postgres", connectionString)
	a.User = user
	if err != nil {
		log.Fatal(err)
	}
	a.Router = mux.NewRouter()
	a.initializeRoutes()
}

func (a *App) authHandler(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gotuser, gotpass, ok := r.BasicAuth()
		if !ok {
			w.Header().Set("Content-Type", "application/json")
			http.Error(w, "Unauthorized.", http.StatusUnauthorized)
			return
		}
		//fmt.Printf("User: %s, Password: %s\n", gotuser, gotpass)
		// Get the Password!
		user := model.User{Username: gotuser}

		err := user.GetUser(a.DB)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			http.Error(w, "Unauthorized.", http.StatusUnauthorized)
			return
		}
		//fmt.Printf("DBUser: %s, DBPassword: %s\n", user.Username, user.Password)

		if gotpass != user.Password {
			w.Header().Set("Content-Type", "application/json")
			http.Error(w, "Unauthorized.", http.StatusUnauthorized)
			return
		}
		handler(w, r)
	}
}

func (a *App) GetUsers(w http.ResponseWriter, r *http.Request) {
	// fmt.Println("Hello, you have requested: %s", r.URL.Path)

	users, err := model.GetUsers(a.DB)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithJSON(w, http.StatusOK, users)
}

func (a *App) GetUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uname := vars["uname"]

	u := model.User{Username: uname}
	if err := u.GetUser(a.DB); err != nil {
		switch err {
		case sql.ErrNoRows:
			respondWithError(w, http.StatusNotFound, "User not found")
		default:
			respondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	respondWithJSON(w, http.StatusOK, u)
}

func (a *App) CreateUser(w http.ResponseWriter, r *http.Request) {

	var u model.User
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&u); err != nil {
		fmt.Println(err.Error())
		respondWithError(w, http.StatusInternalServerError, "Invaild Request payload")
	}

	defer r.Body.Close()

	if err := u.CreateUser(a.DB); err != nil {
		fmt.Println(err.Error())
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusCreated, u)
}

func (a *App) DeleteUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uname := vars["uname"]

	u := model.User{Username: uname}
	if err := u.DeleteUser(a.DB); err != nil {
		fmt.Println(err.Error())
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"result": "success"})
}

func (a *App) UpdateUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uname := vars["uname"]

	var u model.User
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&u); err != nil {
		fmt.Println(err.Error())
		respondWithError(w, http.StatusInternalServerError, "Invalid request payload")
		return
	}

	defer r.Body.Close()
	u.Username = uname
	if err := u.UpdateUser(a.DB); err != nil {
		fmt.Println(err.Error())
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithJSON(w, http.StatusOK, u)
}

func (a *App) initializeRoutes() {

	a.Router.HandleFunc("/users", a.authHandler(a.GetUsers)).Methods("GET")
	a.Router.HandleFunc("/user/{uname}", a.authHandler(a.GetUser)).Methods("GET")
	a.Router.HandleFunc("/user", a.authHandler(a.CreateUser)).Methods("POST")
	a.Router.HandleFunc("/user/{uname}", a.authHandler(a.DeleteUser)).Methods("DELETE")
	a.Router.HandleFunc("/user/{uname}", a.authHandler(a.UpdateUser)).Methods("PUT")

	a.Router.Handle("/PacketRun/", http.StripPrefix("/PacketRun/", http.FileServer(http.Dir("../client/html/"))))

}


func (a *App) Run(s model.Server) {
	if s.UseHTTP && s.UseHTTPS {
		go func() {
			a.startHTTPS(s)
		}()
		a.startHTTP(s)
	} else if s.UseHTTP {
		a.startHTTP(s)
	} else if s.UseHTTPS {
		a.startHTTPS(s)
	} else {

	}
}
// startHTTPs starts the HTTPS listener
func (a *App) startHTTPS(s model.Server) {
	fmt.Println(time.Now().Format("2006-01-02 03:04:05 PM"), "Running HTTPS On Port "+(s.HTTPSPort))
	// Start the HTTPS listener
	log.Fatal(http.ListenAndServeTLS(":" + s.HTTPSPort, s.CertFile, s.KeyFile, a.Router))
}
// startHTTP starts the HTTP listener
func (a *App) startHTTP(s model.Server) {
	fmt.Println(time.Now().Format("2006-01-02 03:04:05 PM"), "Running HTTP on Port"+(s.HTTPPort))

	// Start the HTTP listener
	log.Fatal(http.ListenAndServe(":" + s.HTTPPort, a.Router))
}
