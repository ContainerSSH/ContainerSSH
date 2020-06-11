package main

import (
	"fmt"
)

type Cat struct {
	age     int
	name    string
	friends []string
}

func main() {
	wilson := Cat{7, "Wilson", []string{"Tom", "Tabata", "Willie"}}
	nikita := Cat{}

	nikita.friends = append(nikita.friends, "Syd")

	fmt.Println(wilson)
	fmt.Println(nikita)
}
