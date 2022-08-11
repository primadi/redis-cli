package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"

	"primadi.setiawan/redis-cli/client"
)

var host string
var port string

func init() {
	flag.StringVar(&host, "h", "localhost", "host")
	flag.StringVar(&port, "p", "9163", "port")
}

func main() {
	flag.Parse()

	intPort, err := strconv.Atoi(port)

	if err != nil {
		log.Fatal(err)
	}
	client := client.NewClient(host, intPort)
	err = client.Connect()

	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	for {
		fmt.Printf("%s:%s>", host, port)

		bio := bufio.NewReader(os.Stdin)
		input, _, err := bio.ReadLine()

		if err != nil {
			log.Fatal(err)
		}

		fields := bytes.Fields(input)
		_, err = client.DoRequest(fields[0], fields[1:]...)
		if err != nil {
			log.Fatal(err)
		}

		reply, err := client.GetReply()
		if err != nil {
			log.Fatal(err)
		}

		for _, s := range reply.Format() {
			fmt.Println(s)
		}

	}
}
