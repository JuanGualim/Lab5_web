package main

import (
	"database/sql"
	"fmt"
	"log"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"

	_ "modernc.org/sqlite"
)

func main() {
	db, err := sql.Open("sqlite", "file:series.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}

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

	buffer := make([]byte, 2048)
	n, err := conn.Read(buffer)
	if err != nil {
		return
	}

	request := string(buffer[:n])

	// Servir CSS
	if strings.Contains(request, "GET /style.css") {
		serveFile(conn, "style.css", "text/css")
		return
	}

	// Servir JS
	if strings.Contains(request, "GET /script.js") {
		serveFile(conn, "script.js", "application/javascript")
		return
	}

	// GET /
	if strings.Contains(request, "GET / ") {
		renderHome(conn, db)
		return
	}

	// GET /create
	if strings.Contains(request, "GET /create ") {
		serveFile(conn, "create.html", "text/html")
		return
	}

	// POST /create
	if strings.Contains(request, "POST /create") {

		// Obtener Content-Length
		lines := strings.Split(request, "\r\n")
		var contentLength int

		for _, line := range lines {
			if strings.HasPrefix(line, "Content-Length:") {
				lengthStr := strings.TrimSpace(strings.TrimPrefix(line, "Content-Length:"))
				contentLength, _ = strconv.Atoi(lengthStr)
			}
		}

		// Separar headers del body
		parts := strings.Split(request, "\r\n\r\n")
		if len(parts) < 2 {
			return
		}

		body := parts[1]

		if len(body) > contentLength {
			body = body[:contentLength]
		}

		values, err := url.ParseQuery(body)
		if err != nil {
			log.Println(err)
			return
		}

		name := values.Get("series_name")
		currentStr := values.Get("current_episode")
		totalStr := values.Get("total_episodes")

		current, err := strconv.Atoi(currentStr)
		if err != nil {
			log.Println(err)
			return
		}

		total, err := strconv.Atoi(totalStr)
		if err != nil {
			log.Println(err)
			return
		}

		// INSERT en SQLite
		_, err = db.Exec(
			"INSERT INTO series (name, current_episode, total_episodes) VALUES (?, ?, ?)",
			name, current, total,
		)
		if err != nil {
			log.Println(err)
			return
		}

		// Redirect 303
		response := "HTTP/1.1 303 See Other\r\n" +
			"Location: /\r\n" +
			"\r\n"

		conn.Write([]byte(response))
		return
	}
}

func renderHome(conn net.Conn, db *sql.DB) {

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
