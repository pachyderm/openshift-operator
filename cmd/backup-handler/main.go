package main

import (
	"log"

	"github.com/pachyderm/openshift-operator/cmd/backup-handler/pkg/command"
	"github.com/pachyderm/openshift-operator/cmd/backup-handler/pkg/restapi"
)

func main() {
	// Start the REST api in a goroutine
	go func() {
		if err := restapi.Start(); err != nil {
			log.Panic("error starting REST api", err)
		}

	}()

	command.BackupDispatch()
}
