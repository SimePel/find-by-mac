package main

import (
	"fmt"
	"log"
	"os"
	"strings"

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

	inter := getInterfaceByMac(clientConfig, "10.1.0.29:22", "b79e")
	desc := getDescriptionOfInterface(clientConfig, "10.1.0.29:22", inter)
	for checkIfDescriptionIsASwitch(desc) {
		sw := changeSwitchDescToAppropriateName(desc)
		inter = getInterfaceByMac(clientConfig, sw, "b79e")
		desc = getDescriptionOfInterface(clientConfig, sw, inter)
	}

	fmt.Println(desc)
}

func getDescriptionOfInterface(config *ssh.ClientConfig, host, inter string) string {
	conn, err := ssh.Dial("tcp", host, config)
	if err != nil {
		log.Fatal("unable to connect: ", err)
	}
	defer conn.Close()

	session, err := conn.NewSession()
	if err != nil {
		log.Fatal("failed to create session: ", err)
	}
	defer session.Close()

	b, err := session.Output("show interfaces description | include " + inter)
	if err != nil {
		log.Fatal("failed to run: ", err)
	}

	slices := strings.Fields(string(b))
	if len(slices) < 4 {
		log.Println("description is empty")
		return ""
	}

	return strings.Join(strings.Fields(string(b))[3:], "")
}

func getInterfaceByMac(config *ssh.ClientConfig, host, mac string) string {
	conn, err := ssh.Dial("tcp", host, config)
	if err != nil {
		log.Fatal("unable to connect: ", err)
	}
	defer conn.Close()

	session, err := conn.NewSession()
	if err != nil {
		log.Fatal("failed to create session: ", err)
	}
	defer session.Close()

	b, err := session.Output("show mac address-table | include ." + mac)
	if err != nil {
		log.Fatal("failed to run: ", err)
	}

	return strings.Fields(string(b))[3]
}

func checkIfDescriptionIsASwitch(desc string) bool {
	return strings.HasPrefix(desc, "sw-")
}

func changeSwitchDescToAppropriateName(desc string) string {
	beginning := strings.Index(desc, "sw-")
	ending := strings.LastIndex(desc, "c")
	return desc[beginning:ending+1] + ".noc.asu.ru:22"
}
