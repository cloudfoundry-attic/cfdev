package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os/exec"
	"strings"
	"encoding/json"
	"os"
)

func main() {
	//write resolv.conf
	dns, err := exec.Command("powershell.exe", "-Command", "get-dnsclientserveraddress -family ipv4 | select-object -expandproperty serveraddresses").Output()
	if err != nil {
		log.Fatal(err)
	}

	dns_file := ""

	scanner := bufio.NewScanner(bytes.NewReader(dns))
	for scanner.Scan() {
		line := scanner.Text()
		dns_file += fmt.Sprintf("nameserver %s\n", line)
	}

	ioutil.WriteFile("resolv.conf", []byte(dns_file), 0600)

	//write dhcp.json
	dhcp, err := exec.Command("powershell.exe", "-Command", "get-dnsclient | select-object -expandproperty connectionspecificsuffix").Output()
	if err != nil {
		log.Fatal(err)
	}

	var output struct {
		SearchDomains []string `json:"searchDomains"`
		DomainName string `json:"domainName"`
	}

	scanner = bufio.NewScanner(bytes.NewReader(dhcp))
	for scanner.Scan() {
		if line := scanner.Text(); strings.TrimSpace(line) != "" {
			output.SearchDomains = append(output.SearchDomains, line)
		}

		if len(output.SearchDomains) > 0 {
			output.DomainName = output.SearchDomains[len(output.SearchDomains) - 1]
		}
	}

	file, err := os.Create("dhcp.json")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	json.NewEncoder(file).Encode(&output)
}
