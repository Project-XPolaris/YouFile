package main

import (
	"log"
	"youfile/api"
	"youfile/config"
	"youfile/service"
)

func main() {
	err := config.LoadAppConfig()
	if err != nil {
		log.Fatal(err)
	}
	err = service.LoadFstab()
	if err != nil {
		log.Fatal(err)
	}
	api.RunApiService()
}
