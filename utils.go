package main

import (
	"net"
	"os"
)

// Sirve archivos estáticos (CSS, JS, HTML)
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

// Envía error 400 cuando falla una validación
func sendBadRequest(conn net.Conn, message string) {
	response := "HTTP/1.1 400 Bad Request\r\n" +
		"Content-Type: text/plain\r\n" +
		"\r\n" +
		message

	conn.Write([]byte(response))
}
