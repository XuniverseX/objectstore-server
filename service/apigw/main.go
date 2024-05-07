package main

import (
	"objectstore-server/service/apigw/route"
)

func main() {
	r := route.Router()
	r.Run("0.0.0.0:8080")
}
