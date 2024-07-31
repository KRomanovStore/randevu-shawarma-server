package connection

import (
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"

	"randevu-shawarma-server/supply"
	"randevu-shawarma-server/users"
)

func СreateServer() {
	router := httprouter.New()
	users.RegisterRoutes(router)
	supply.RegisterRoutes(router)

	log.Fatal(http.ListenAndServe(":8090", router))
}
