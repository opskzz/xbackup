package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	cfg := flag.String("c", "xbackup.yml", "config file")
	ver := flag.Bool("v", false, "show version")
	flag.Parse()

	if *ver {
		fmt.Println("mbackup v1.0")
		os.Exit(0)
	}

	config := PXConfig(*cfg)

	//init backup dir
	initBackupDir(config.XDirs)
	//check mysql online
	initMySQLAlive(config.XUser, config.XPass, config.QPort, config.QHost)
	//check redis online
	initQueueAlive(config.QHost, config.QPort)

	go httpRsync(config.HAddr, config.HPort, config.XDirs)
	select {}
}
