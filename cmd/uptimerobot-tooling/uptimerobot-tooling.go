package main

import (
	"flag"
	"github.com/bennsimon/uptime-robot-tooling/pkg/handler"
	"github.com/bennsimon/uptime-robot-tooling/pkg/model"
	log "github.com/sirupsen/logrus"
)

func main() {
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})

	payload := flag.String("d", "", "JSON file path or JSON string containing the uptimerobot-tooling robot resource(s).")
	resource := flag.String("r", "monitor", "Resource type that will be acted on. e.g monitor, alert_contact")
	action := flag.String("a", "update", "Action to be performed on the model e.g create, update, delete")
	flag.Parse()

	if resultPayload := handler.HandleRequest(*payload, *resource, *action); resultPayload != nil {
		for _, value := range resultPayload {
			if value != nil {
				if value[model.ErrorResultField] == nil {
					log.Infof("%s action on monitor %s was successful", *action, value[model.MonitorNameResultField])
				} else {
					log.Errorf("%s action on monitor %s failed with '%v'", *action, value[model.MonitorNameResultField], value[model.ErrorResultField])
				}
			}
		}
	}
}
