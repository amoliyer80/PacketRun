package main

import (
	"errors"
	"os"

	"encoding/json"
	"io/ioutil"
	"regexp"

	"github.com/amoliyer80/PacketRun/app"
	"github.com/amoliyer80/PacketRun/model"
)



func Load(filename string, config interface{}) error {
	// Read the config file.
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return errors.New("Error reading file")
	}

	// Regex monstrosity because of the lack of lookbehinds/aheads.

	// Standalone comments.
	r1, _ := regexp.Compile(`(?m)^(\s+)?//(.*)$`)

	// Numbers and boolean.
	r2, _ := regexp.Compile(`(?m)"(.+?)":(\s+)?([0-9\.\-]+|true|false|null)(\s+)?,(\s+)?//(.*)$`)

	// Strings.
	r3, _ := regexp.Compile(`(?m)"(.+?)":(\s+)?"(.+?)"(\s+)?,(\s+)?//(.*)$`)

	// Arrays and objects.
	r4, _ := regexp.Compile(`(?m)"(.+?)":(\s+)?([\{\[])(.+?)([\}\]])(\s+)?,(\s+)?//(.*)$`)

	res := r1.ReplaceAllString(string(data), "")
	res = r2.ReplaceAllString(res, `"$1": $3,`)
	res = r3.ReplaceAllString(res, `"$1": "$3",`)
	res = r4.ReplaceAllString(res, `"$1": $3$4$5,`)

	// Decode json.
	if err := json.Unmarshal([]byte(res), &config); err != nil {
		return err
	}

	return nil
}

func main() {
	var config = &model.Configuration{}
	Load("config"+string(os.PathSeparator)+"config.json", config)

	a := app.App{}
	user := model.User{
		Username: "amol",
		Password: "amolpw",
		Realm:    "test",
	}
	a.Initialize(os.Getenv("APP_DB_USERNAME"),
		os.Getenv("APP_DB_PASSWORD"),
		os.Getenv("APP_DB_NAME"), user)
	a.Run(config.Server)
}
