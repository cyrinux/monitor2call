package main

import (
	"log"
	"net/http"
	"strconv"

	"github.com/cyrinux/monitor2call/api"      // api lib
	"github.com/cyrinux/monitor2call/database" // database
	"github.com/cyrinux/monitor2call/helpers"  // common helpers lib
)

// @title Monitor2Call api
// @version 1.0
// @description This is a monitoring to call server api.
// @termsOfService http://swagger.io/terms/
// @contact.name Cyril Levis <c.levis@oodrive.com>
// @contact.url https://github.com/cyrinux/monitor2call/issues/new
// @contact.email Cyril Levis <c.levis@oodrive.com>
// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
// @BasePath /
func main() {
	// Init DB
	db := database.InitDb()
	database.CreateOrUpdateView(db)

	httpsEnabled := helpers.GetEnv("HTTPS_ENABLED", "false")

	var err error

	if bool, _ := strconv.ParseBool(httpsEnabled); bool == true {
		err = http.ListenAndServeTLS(":3000", "/go/keys/localhost.pem", "/go/keys/localhost.key", api.Handlers())
	} else {
		err = http.ListenAndServe(":3000", api.Handlers())
	}

	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}

}