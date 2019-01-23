package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"golang.org/x/crypto/ssh"
)

const (
	domain     = ".noc.asu.ru"
	rootSwitch = "sw-102-0-mc"
)

var (
	sw  string
	mac string
)

func init() {
	flag.StringVar(&mac, "mac", "b79e", "enter your searching mac address")
	flag.StringVar(&sw, "sw", rootSwitch, "enter nearest switch as you think")
}

func newConfigWithInsecureCiphers() (c ssh.Config) {
	c.SetDefaults()
	c.Ciphers = append(c.Ciphers, "aes128-cbc")
	return c
}

func main() {
	clientConfig := &ssh.ClientConfig{
		Config: newConfigWithInsecureCiphers(),
		User:   os.Getenv("CISCO_LOGIN"),
		Auth: []ssh.AuthMethod{
			ssh.Password(os.Getenv("CISCO_PASS")),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	flag.Parse()
	sw += domain
	inter := getInterfaceByMac(clientConfig, sw, mac)
	desc := getDescriptionOfInterface(clientConfig, sw, inter)
	for descriptionIsASwitch(desc) {
		sw = changeSwitchDescToAppropriateName(desc)
		if httpConnectionIsAvailable(sw) {
			fmt.Printf("Мак на: %v, который не поддерживает ssh\n", sw)
			return
		}
		inter = getInterfaceByMac(clientConfig, sw, mac)
		desc = getDescriptionOfInterface(clientConfig, sw, inter)
	}

	fmt.Printf("Мак на: %v воткнут в %v порт. Описание порта: %v\n", sw, inter, desc)
}

func getInterfaceByMac(config *ssh.ClientConfig, host, mac string) string {
	conn, err := ssh.Dial("tcp", host+":22", config)
	if err != nil {
		log.Fatal("failed to connect by ssh: ", err)
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

	if !strings.Contains(string(b), mac) {
		fmt.Println("Этот мак адрес не удалось найти.")
		os.Exit(0)
	}

	lines := strings.Split(string(b), "\n")
	return strings.Fields(lines[len(lines)-2])[3]
}

func getDescriptionOfInterface(config *ssh.ClientConfig, host, inter string) string {
	conn, err := ssh.Dial("tcp", host+":22", config)
	if err != nil {
		log.Fatal("failed to connect by ssh: ", err)
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
		fmt.Printf("Мак на: %v воткнут в %v порт, но описание отсутствует.\n", host, inter)
		os.Exit(0)
	}

	return strings.Join(desc, "")
}

func httpConnectionIsAvailable(host string) bool {
	resp, err := http.Get("http://" + host)
	if err != nil {
		return false
	}
	if resp.StatusCode == http.StatusOK {
		return true
	}
	log.Printf("Однако не 200 код\n")
	return false
}

func descriptionIsASwitch(desc string) bool {
	return strings.Contains(desc, "sw-")
}

func changeSwitchDescToAppropriateName(desc string) string {
	beginning := strings.Index(desc, "sw-")
	ending := strings.LastIndex(desc, "c")
	return desc[beginning:ending+1] + domain
}
