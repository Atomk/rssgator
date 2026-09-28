package main

import (
	"fmt"
	"rssgator/internal/config"
)

func main() {
	fmt.Println("--- before ---")
	configOptions := config.Read()
	fmt.Printf("%+v\n", configOptions)

	fmt.Println("--- after ---")
	configOptions.SetUser("atomk")
	configOptions = config.Read()
	fmt.Printf("%+v\n", configOptions)
}
