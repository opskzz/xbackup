package main

import (
	"flag"
	"fmt"
	"os"

	"./g"
)

func main() {
	cfg := flag.String("c", "xbackup.yml", "config file")
	ver := flag.Bool("v", false, "show version")
	flag.Parse()

	if *ver {
		fmt.Println(g.Version)
		os.Exit(0)
	}
	config := g.PXConfig(*cfg)
	//init backup dir
	g.InitBackupDir(config.XDirs)
	//check mysql online
	g.InitMySQLAlive(config.XUser, config.XPass, config.QPort, config.QHost)
	//check redis online
	g.InitQueueAlive(config.QHost, config.QPort)

	//backup progress on
	go g.DoBackup(config.QHost, config.QPort, config.QName)
	//init http rsync
	go g.HTTPRsync(config.HAddr, config.HPort, config.XDirs)
	select {}
}
