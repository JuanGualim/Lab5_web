package main

import (
	"database/sql"
	"fmt"
	"log"
	"net"

	_ "modernc.org/sqlite"
)

func main() {
	// Abrir base de datos
	db, err := sql.Open("sqlite", "file:series.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Verificar conexión
	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}

	// Iniciar servidor TCP
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	fmt.Println("Servidor escuchando en http://localhost:8080")

	// Aceptar conexiones
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println(err)
			continue
		}

		// Delegar manejo de la conexión
		go handleClient(conn, db)
	}
}
