package main

import (
	"flag"
	"fmt"

	"casino/configs"
	"casino/internal/config"
	"casino/internal/database"
	"casino/internal/seed"
)

func main() {
	configFile := flag.String("c", configs.Location("casino.yml"), "config file")
	flag.Parse()

	if err := config.Init(*configFile); err != nil {
		panic("init config error: " + err.Error())
	}
	database.InitDB()
	defer func() {
		_ = database.CloseDB()
	}()

	if err := seed.SeedH5(database.GetDB()); err != nil {
		panic(err)
	}
	fmt.Println("H5 seed data inserted successfully")
}
