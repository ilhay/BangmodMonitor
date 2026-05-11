package checker

import (
	"fmt"
	"net"
	"time"
)

type TCPResult struct {
	Host       string
	Port       int
	ResponseMS int64
	IsUp       bool
	Error      string
}

func CheckTCP(host string, port int) TCPResult {
	addr := fmt.Sprintf("%s:%d", host, port)
	start := time.Now()

	conn, err := net.DialTimeout("tcp", addr, 10*time.Second)
	elapsed := time.Since(start).Milliseconds()

	if err != nil {
		return TCPResult{Host: host, Port: port, ResponseMS: elapsed, IsUp: false, Error: err.Error()}
	}
	conn.Close()

	return TCPResult{Host: host, Port: port, ResponseMS: elapsed, IsUp: true}
}
