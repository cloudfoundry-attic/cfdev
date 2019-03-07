package main

import (
	"code.cloudfoundry.org/cfdev/servicew/config"
	"code.cloudfoundry.org/cfdev/servicew/program"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	configPath, err := configPath()
	expectNoError(err)

	contents, err := ioutil.ReadFile(configPath)
	expectNoError(err)

	var conf config.Config
	err = yaml.Unmarshal(contents, &conf)
	expectNoError(err)

	prog, err := program.New(conf, os.Stdout)
	expectNoError(err)

	if len(os.Args) == 1 {
		err = prog.Service.Run()
		expectNoError(err)

		os.Exit(0)
	}

	switch os.Args[1] {
	case "status":
		fmt.Println(strings.Title(prog.Status()))
	case "install":
		err = prog.Install()
	case "start":
		err = prog.StartService()
	case "stop":
		err = prog.StopService()
	case "uninstall":
		err = prog.Uninstall()
	default:
		err = fmt.Errorf("unsupported command: '%q'", os.Args[1])
	}

	expectNoError(err)
}

func configPath() (string, error) {
	fullexecpath, err := os.Executable()
	if err != nil {
		return "", err
	}

	dir, execname := filepath.Split(fullexecpath)
	ext := filepath.Ext(execname)
	name := execname[:len(execname)-len(ext)]

	return filepath.Join(dir, name+".yml"), nil
}

func expectNoError(err error) {
	if err != nil {
		log.Fatalln("Error:", err)
	}
}
