package main

import (
	"fmt"
	"log"
	"os"

	"golang.org/x/crypto/ssh"
)

const (
	login = "localadmin"
)

var (
	password = os.Getenv("CISCO_PASS")
)

func newConfigWithInsecureCiphers() (c ssh.Config) {
	c.SetDefaults()
	c.Ciphers = append(c.Ciphers, "aes128-cbc")
	return c
}

func main() {
	clientConfig := &ssh.ClientConfig{
		Config: newConfigWithInsecureCiphers(),
		User:   login,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	conn, err := ssh.Dial("tcp", "10.1.0.29:22", clientConfig)
	if err != nil {
		log.Fatal("unable to connect: ", err)
	}
	defer conn.Close()

	fmt.Printf("Connection was established\n")
}
