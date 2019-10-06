package cmd

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"

	"github.com/lpicanco/microcache-server/cache"
	"github.com/spf13/cobra"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start server daemon",
	Run:   startServer,
}

var port uint16

func init() {
	rootCmd.AddCommand(serveCmd)
	serveCmd.Flags().Uint16VarP(&port, "port", "p", 6542, "Port to bind to")
}

func startServer(cmd *cobra.Command, args []string) {
	log.Printf("Starting microcache at port %v\n", port)
	l, err := net.Listen("tcp", fmt.Sprintf(":%v", port))

	if err != nil {
		log.Fatalf("Error starting server: %s", err.Error())
	}

	defer l.Close()

	for {
		c, err := l.Accept()
		if err != nil {
			log.Printf("Error accepting connection: %s", err.Error())
			continue
		}

		go handleConnection(c)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	log.Printf("Connected with %s\n", conn.RemoteAddr())

	for {
		netData, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			log.Printf("Error reading: %s", err.Error())
			return
		}

		args := strings.SplitN(strings.TrimSuffix(netData, "\r\n"), " ", 3)
		cmd := strings.ToUpper(strings.TrimSpace(args[0]))
		log.Printf("Parsing command: %s\n", cmd)

		switch cmd {
		case "GET":
			processCommandGet(conn, args)
		case "PUT":
			processCommandPut(conn, args)
		case "INVALIDATE":
			processCommandInvalidate(conn, args)
		case "QUIT":
			return
		}
	}
}

func processCommandGet(conn net.Conn, args []string) {
	if !checkArguments(conn, args, 1) {
		return
	}

	value, found := cache.Cache.Get(args[1])
	var response string
	if found {
		response = fmt.Sprintf("%s\n", value.(string))
	} else {
		response = "Key not found\n"
	}

	writeResponse(conn, response)
}

func processCommandPut(conn net.Conn, args []string) {
	if !checkArguments(conn, args, 2) {
		return
	}

	cache.Cache.Put(args[1], args[2])
}

func processCommandInvalidate(conn net.Conn, args []string) {
	if !checkArguments(conn, args, 1) {
		return
	}

	cache.Cache.Invalidate(args[1])
}

func writeResponse(conn net.Conn, data string) {
	writer := bufio.NewWriter(conn)
	_, err := writer.WriteString(data)
	if err == nil {
		writer.Flush()
	}
}

func checkArguments(conn net.Conn, args []string, expected int) bool {
	if len(args)-1 != expected {
		writeResponse(conn, fmt.Sprintf("Invalid argument count. Expected %v. Found: %v\n", expected, len(args)-1))
		return false
	}

	return true
}
