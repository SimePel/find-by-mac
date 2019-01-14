package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"golang.org/x/crypto/ssh"
)

const (
	rootSW    = "sw-102-0-mc"
	swPostfix = ".noc.asu.ru:22"
)

var (
	mac      string
	sw       string
	login    = os.Getenv("CISCO_LOGIN")
	password = os.Getenv("CISCO_PASS")
)

func init() {
	flag.StringVar(&mac, "mac", "b79e", "enter your searching mac address")
	flag.StringVar(&sw, "sw", rootSW, "enter nearest switch as you think")
}

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

	flag.Parse()
	sw += swPostfix
	inter, err := getInterfaceByMac(clientConfig, sw, mac)
	if err != nil {
		fmt.Printf("Мак на: %v, который не поддерживает ssh\n", sw)
		return
	}
	desc := getDescriptionOfInterface(clientConfig, sw, inter)
	for checkIfDescriptionIsASwitch(desc) {
		sw = changeSwitchDescToAppropriateName(desc)
		inter, err = getInterfaceByMac(clientConfig, sw, mac)
		if err != nil {
			fmt.Printf("Мак на: %v, который не поддерживает ssh\n", sw)
			return
		}
		desc = getDescriptionOfInterface(clientConfig, sw, inter)
	}

	fmt.Printf("Мак на: %v воткнут в %v порт. Описание порта: %v\n", sw, inter, desc)
}

func getInterfaceByMac(config *ssh.ClientConfig, host, mac string) (string, error) {
	conn, err := ssh.Dial("tcp", host, config)
	if err != nil {
		return "", err
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

	lines := strings.Split(string(b), "\n")
	return strings.Fields(lines[len(lines)-2])[3], nil
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

	lines := strings.Split(string(b), "\n")
	lastLine := strings.Fields(lines[len(lines)-2])
	desc := lastLine[3:len(lastLine)]

	if len(desc) == 0 {
		log.Println("description is empty")
		return ""
	}

	return strings.Join(desc, "")
}

func checkIfDescriptionIsASwitch(desc string) bool {
	return strings.Contains(desc, "sw-")
}

func changeSwitchDescToAppropriateName(desc string) string {
	beginning := strings.Index(desc, "sw-")
	ending := strings.LastIndex(desc, "c")
	return desc[beginning:ending+1] + swPostfix
}
