package main

import (
	"encoding/base64"
	"log"
	"net/http"
	"os"
	"testing"
	"net/http/httptest"
	"github.com/amoliyer80/PacketRun/app"
	"github.com/amoliyer80/PacketRun/model"
	"fmt"
	"reflect"
	"encoding/json"
)

var a app.App

const usersCreationQuery = `CREATE TABLE IF NOT EXISTS users
(
uname TEXT NOT NULL,
password TEXT NOT NULL,
realm TEXT NOT NULL,
uright TEXT NOT NULL,
CONSTRAINT users_pkey PRIMARY KEY (uname)
)`

func ensureTableExists() {
	if _, err := a.DB.Exec(usersCreationQuery); err != nil {
		log.Fatal(err)
	}
}

func createUser(username, password, realm string, uright string) {
	_, err := a.DB.Exec("INSERT INTO users (uname, password, realm, uright) VALUES($1, $2, $3, $4)",
		username, password, realm, uright)
	if err != nil {
		log.Fatal(err)
	}
}

func clearTable() {

	a.DB.Exec("DELETE FROM users")
	createUser("amol", "amolpw", "test", "1")

}

func executeRequest(req *http.Request) (rr *httptest.ResponseRecorder) {
	rr = httptest.NewRecorder()
	a.Router.ServeHTTP(rr, req)
	return
}

func checkResponseCode(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("expected response code %d, actual response code : %d", expected, actual)
	}
}

func TestMain(m *testing.M) {
	// call flag.Parse() here if TestMain uses flags
	a = app.App{}
	user := model.User{Username: "amol",
		Password: "amolpw",
		Realm:    "test",
	}
	a.Initialize(os.Getenv("APP_DB_USERNAME"),
		os.Getenv("APP_DB_PASSWORD"),
		os.Getenv("APP_DB_NAME"), user)
	ensureTableExists()
	code := m.Run()
	clearTable()
	os.Exit(code)
}

func appendBasicAuth(r *http.Request) {
	upass := "amol:amolpw"
	encoded := base64.StdEncoding.EncodeToString([]byte(upass))
	r.Header.Add("Authorization", "Basic "+encoded)
	// fmt.Println(r.Header)

}

func TestEmptyUsers(t *testing.T) {
	clearTable()

	req, _ := http.NewRequest("GET", "/users", nil)
	appendBasicAuth(req)
	response := executeRequest(req)

	u1 := model.User{Username: "amol", Password: "amolpw", Realm:"test", Right:1}
	setUsers := []model.User{u1}

	var users []model.User
	err := json.Unmarshal(response.Body.Bytes(), &users)
	if err != nil {
		fmt.Println(response.Body)
		t.Errorf("Error When Unmarshalling user body : %s", err.Error())
	}
	checkResponseCode(t, http.StatusOK, response.Code)

	if len(users) != 1 {
		t.Errorf("Expected 1 Entry found %d entry(s)", len(users))
	} else {
		for i, user := range setUsers {
			same := reflect.DeepEqual(user, users[i])
			if !same {
				t.Error("Got Todo not equal to the expected one")
			}
		}
	}

}
