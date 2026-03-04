package main

import (
	"database/sql"
	"fmt"
	"log"
	"net"
	"os"
	"strings"

	_ "modernc.org/sqlite"
)

func main() {
	// Abrir base de datos UNA sola vez
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

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println(err)
			continue
		}

		go handleClient(conn, db)
	}
}

func handleClient(conn net.Conn, db *sql.DB) {
	defer conn.Close()

	buffer := make([]byte, 1024)
	conn.Read(buffer)
	request := string(buffer)

	// Servir CSS
	if contains(request, "GET /style.css") {
		serveFile(conn, "style.css", "text/css")
		return
	}

	// Servir JS
	if contains(request, "GET /script.js") {
		serveFile(conn, "script.js", "application/javascript")
		return
	}

	// Página principal
	if contains(request, "GET / ") {

		rows, err := db.Query("SELECT id, name, current_episode, total_episodes FROM series")
		if err != nil {
			log.Println(err)
			return
		}
		defer rows.Close()

		var rowsHTML string

		for rows.Next() {
			var id int
			var name string
			var current int
			var total int

			err := rows.Scan(&id, &name, &current, &total)
			if err != nil {
				log.Println(err)
				continue
			}

			rowsHTML += fmt.Sprintf(
				"<tr><td>%d</td><td>%s</td><td>%d</td><td>%d</td></tr>",
				id, name, current, total,
			)
		}

		content, err := os.ReadFile("index.html")
		if err != nil {
			log.Println(err)
			return
		}

		html := string(content)
		html = strings.Replace(html, "{{ROWS}}", rowsHTML, 1)

		response := "HTTP/1.1 200 OK\r\n" +
			"Content-Type: text/html; charset=utf-8\r\n" +
			"\r\n" +
			html

		conn.Write([]byte(response))
	}

	if strings.Contains(request, "GET /create ") {
		serveFile(conn, "create.html", "text/html")
		return
	}
}
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

func serveFile(conn net.Conn, filename string, contentType string) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return
	}

	response := "HTTP/1.1 200 OK\r\n" +
		"Content-Type: " + contentType + "\r\n" +
		"\r\n"

	conn.Write([]byte(response))
	conn.Write(content)
}
