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
)

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

	// POST /update (+1)
	if strings.HasPrefix(request, "POST /update") {

		lines := strings.Split(request, "\r\n")
		requestLine := lines[0]

		parts := strings.Split(requestLine, " ")
		path := parts[1]

		pathParts := strings.SplitN(path, "?", 2)
		route := pathParts[0]

		if route == "/update" && len(pathParts) > 1 {

			params, err := url.ParseQuery(pathParts[1])
			if err != nil {
				return
			}

			idStr := params.Get("id")
			id, err := strconv.Atoi(idStr)
			if err != nil {
				return
			}

			_, err = db.Exec(
				`UPDATE series
				 SET current_episode = current_episode + 1
				 WHERE id = ? AND current_episode < total_episodes`,
				id,
			)
			if err != nil {
				log.Println(err)
				return
			}

			conn.Write([]byte("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\n\r\nok"))
			return
		}
	}

	// POST /decrement (-1)
	if strings.HasPrefix(request, "POST /decrement") {

		lines := strings.Split(request, "\r\n")
		requestLine := lines[0]

		parts := strings.Split(requestLine, " ")
		path := parts[1]

		pathParts := strings.SplitN(path, "?", 2)
		route := pathParts[0]

		if route == "/decrement" && len(pathParts) > 1 {

			params, err := url.ParseQuery(pathParts[1])
			if err != nil {
				return
			}

			idStr := params.Get("id")
			id, err := strconv.Atoi(idStr)
			if err != nil {
				return
			}

			_, err = db.Exec(
				`UPDATE series
				 SET current_episode = current_episode - 1
				 WHERE id = ? AND current_episode > 1`,
				id,
			)
			if err != nil {
				log.Println(err)
				return
			}

			conn.Write([]byte("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\n\r\nok"))
			return
		}
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

		lines := strings.Split(request, "\r\n")
		var contentLength int

		for _, line := range lines {
			if strings.HasPrefix(line, "Content-Length:") {
				lengthStr := strings.TrimSpace(strings.TrimPrefix(line, "Content-Length:"))
				contentLength, _ = strconv.Atoi(lengthStr)
			}
		}

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
			return
		}

		name := values.Get("series_name")
		currentStr := values.Get("current_episode")
		totalStr := values.Get("total_episodes")

		current, err := strconv.Atoi(currentStr)
		if err != nil {
			sendBadRequest(conn, "Current episode must be a number")
			return
		}

		total, err := strconv.Atoi(totalStr)
		if err != nil {
			sendBadRequest(conn, "Total episodes must be a number")
			return
		}

		// Validaciones
		if strings.TrimSpace(name) == "" {
			sendBadRequest(conn, "Name cannot be empty")
			return
		}

		if current < 1 {
			sendBadRequest(conn, "Current episode must be at least 1")
			return
		}

		if total < 1 {
			sendBadRequest(conn, "Total episodes must be at least 1")
			return
		}

		if current > total {
			sendBadRequest(conn, "Current episode cannot be greater than total episodes")
			return
		}

		_, err = db.Exec(
			"INSERT INTO series (name, current_episode, total_episodes) VALUES (?, ?, ?)",
			name, current, total,
		)
		if err != nil {
			log.Println(err)
			return
		}

		conn.Write([]byte("HTTP/1.1 303 See Other\r\nLocation: /\r\n\r\n"))
		return
	}
}

func renderHome(conn net.Conn, db *sql.DB) {

	rows, err := db.Query("SELECT id, name, current_episode, total_episodes FROM series")
	if err != nil {
		return
	}
	defer rows.Close()

	var rowsHTML string

	for rows.Next() {
		var id, current, total int
		var name string

		rows.Scan(&id, &name, &current, &total)

		progress := 0
		if total > 0 {
			progress = (current * 100) / total
		}

		completed := current >= total
		statusText := ""
		barClass := "progress-fill"

		if completed {
			statusText = " <span class='completed'>✅ Completed</span>"
			barClass = "progress-fill completed-bar"
		}

		rowsHTML += fmt.Sprintf(
			`<tr>
				<td>%d</td>
				<td>%s</td>
				<td>%d</td>
				<td>%d</td>
				<td>
					<div class="progress-bar">
						<div class="%s" style="width:%d%%;"></div>
					</div>
					%d%%%s
				</td>
				<td>
					<button onclick='prevEpisode(%d)'>-1</button>
					<button onclick='nextEpisode(%d)'>+1</button>
				</td>
			</tr>`,
			id, name, current, total,
			barClass, progress,
			progress, statusText,
			id, id,
		)
	}

	content, _ := os.ReadFile("index.html")
	html := strings.Replace(string(content), "{{ROWS}}", rowsHTML, 1)

	response := "HTTP/1.1 200 OK\r\nContent-Type: text/html; charset=utf-8\r\n\r\n" + html
	conn.Write([]byte(response))
}
