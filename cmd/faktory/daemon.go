package main

import (
	"fmt"
	"log"
	"time"

	"github.com/hunter-io/faktory/cli"
	"github.com/hunter-io/faktory/client"
	"github.com/hunter-io/faktory/util"
	"github.com/hunter-io/faktory/webui"
)

func logPreamble() {
	log.SetFlags(0)
	log.Println(client.Name, client.Version)
	log.Println(fmt.Sprintf("Copyright © %d Contributed Systems LLC", time.Now().Year()))
	log.Println(client.Licensing)
}

func main() {
	logPreamble()

	opts := cli.ParseArguments()
	util.InitLogger(opts.LogLevel)
	util.Debugf("Options: %v", opts)

	s, stopper, err := cli.BuildServer(opts)
	if stopper != nil {
		defer stopper()
	}

	if err != nil {
		util.Error("Unable to create Faktory server", err)
		return
	}

	err = s.Boot()
	if err != nil {
		util.Error("Unable to boot the command server", err)
		return
	}

	s.Register(webui.Subsystem(opts.WebBinding))

	go cli.HandleSignals(s)
	go s.Run()

	<-s.Stopper()
	s.Stop(nil)
}
