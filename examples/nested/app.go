package main

import (
	"fmt"
	"github.com/trhura/simplecli"
)

// Database ...
type Database struct {
	Path string
}

// Create database
func (db Database) Create() {
	fmt.Println("Creating database.")
}

// Drop database
func (db Database) Drop() {
	fmt.Println("Dropping database.")
}

// App ...
type App struct {
	Database *Database
	Port     int
}

// Start the app
func (app App) Start() {
	fmt.Printf("Listening app at %d.\n", app.Port)
}

// Reload the app
func (app App) Reload() {
	fmt.Println("Reloading app.")
}

// Kill the app
func (app App) Kill() {
	fmt.Println("Stoping app.")
}

func main() {
	simplecli.Handle(&App{
		Database: &Database{},
		Port:     8080,
	})
}
