package main

import (
	"fmt"
	"rssgator/internal/config"
)

func main() {
	configOptions := config.Read()
	fmt.Println("DB url: " + configOptions.DbUrl)
}
