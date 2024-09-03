package handler

import (
	"github.com/onaio/uptimerobot-tooling/internal/pkg/util/fileutil"
	"github.com/onaio/uptimerobot-tooling/pkg/model"
	"github.com/onaio/uptimerobot-tooling/pkg/service/monitor"
	log "github.com/sirupsen/logrus"
	"strings"
)

func HandleRequest(payload string, resource string, action string) []map[string]interface{} {
	input, err := fileutil.TransformInputToString(payload)
	var resultPayload []map[string]interface{} = nil
	if err == nil {
		if mapInterface, err := fileutil.TransformStringToMapInterface(input); err == nil {
			if strings.EqualFold(resource, model.Monitor) {

				monitorService := monitor.New()
				resultPayload = monitorService.HandleRequest(mapInterface, model.Args(action))
			} else if strings.EqualFold(resource, model.AlertContact) {
				//TODO add alert-contact service
				log.Error("alert contact not supported yet")
			} else {
				log.Errorf("invalid resource %s specified", resource)
			}
		} else {
			log.Error(err)
		}
	} else {
		log.Error(err)
	}
	return resultPayload
}
