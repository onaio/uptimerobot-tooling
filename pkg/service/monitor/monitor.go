package monitor

import (
	"errors"
	"fmt"
	"github.com/onaio/uptimerobot-tooling/pkg/model"
	"github.com/onaio/uptimerobot-tooling/pkg/service"
	"github.com/onaio/uptimerobot-tooling/pkg/util/httputil"
	log "github.com/sirupsen/logrus"
	"math"
	"os"
	"strconv"
	"strings"
)

var staticPropertiesToResolve = map[string]map[string]uint8{
	httputil.TypeField: {
		"HTTP":      1,
		"HTTPS":     1,
		"Keyword":   2,
		"Ping":      3,
		"Port":      4,
		"Heartbeat": 5,
	},
	httputil.SubTypeField: {
		"HTTP":  1,
		"HTTPS": 2,
		"FTP":   3,
		"SMTP":  4,
		"POP3":  5,
		"IMAP":  6,
	},
	httputil.KeywordTypeField: {
		"exists":     1,
		"not exists": 2,
	},
	httputil.KeywordCaseTypeField: {
		"case sensitive":   0,
		"case insensitive": 1,
	},
	httputil.HttpAuthTypeField: {
		"Basic":           1,
		"Digest":          2,
		"HTTP Basic Auth": 1,
	},
}

const (
	MonitorAlertContactsAttribDelimiterEnv       = "MONITOR_ALERT_CONTACTS_ATTRIB_DELIMITER"
	MonitorAlertContactsDelimiterEnv             = "MONITOR_ALERT_CONTACTS_DELIMITER"
	MonitorAlertContactsResolveByFriendlyNameEnv = "MONITOR_RESOLVE_ALERT_CONTACTS_BY_FRIENDLY_NAME"
	MonitorResolveByFriendlyNameEnv              = "MONITOR_RESOLVE_BY_FRIENDLY_NAME"
)
const (
	AlertContactsAttribDelimiter = "_"
	AlertContactsDelimiter       = "-"
)

const (
	MsgAlertContactFetchErr                   = "error occurred when fetching alert contacts: %s was returned and %s was sent"
	MsgAlertContactNotResolved                = "alert contacts %s could not resolved to an id"
	MsgFieldIsInvalid                         = "%s field invalid"
	MsgMappingNotFound                        = "mapping for %s -> %s not found"
	MsgNoAlertContactsFound                   = "no alert contacts found"
	MsgMonitorDoesNotExist                    = "monitor does not exist"
	MsgFieldMissing                           = "%s needs to be specified"
	MsgCheckIfMonitorHasRequiredFields        = "checking if provided monitor has all the required fields"
	MsgOriginalAndProviderMonitorConflictType = "original and provided monitor have conflicting types (orig: %v and provided: %v) will be recreated check https://uptimerobot.com/#editMonitorWrap"
	MsgActionNotSupported                     = "%s action is not supported"
	MsgUnknownErr                             = "unknown error %s"
)

func (service *MonitorService) HandleRequest(dataMapInterface []map[string]interface{}, action model.Args) []map[string]interface{} {
	resultArrayMap := make([]map[string]interface{}, len(dataMapInterface))
	for idx, dataMap := range dataMapInterface {
		resultArrayMap[idx] = make(map[string]interface{})
		if action == model.Create || action == model.Update {
			if err := service.isValidPayload(dataMap, action); err != nil {
				resultArrayMap = createResultObject(model.Result{ErrorResultField: err, NameResultField: dataMap[httputil.FriendlyNameField]}, idx, resultArrayMap)
				return resultArrayMap
			}

			if err := service.resolveMonitorProperties(dataMap); err != nil {
				resultArrayMap = createResultObject(model.Result{ErrorResultField: err, NameResultField: dataMap[httputil.FriendlyNameField]}, idx, resultArrayMap)
				return resultArrayMap
			}

			if action == model.Update {
				if err := service.UpdateRequest(dataMap); err != nil {
					resultArrayMap = createResultObject(model.Result{ErrorResultField: err, NameResultField: dataMap[httputil.FriendlyNameField]}, idx, resultArrayMap)
					return resultArrayMap
				}
			} else {
				if err := service.CreateRequest(dataMap); err != nil {
					resultArrayMap = createResultObject(model.Result{ErrorResultField: err, NameResultField: dataMap[httputil.FriendlyNameField]}, idx, resultArrayMap)
					return resultArrayMap
				}
			}
			resultArrayMap = createResultObject(model.Result{NameResultField: dataMap[httputil.FriendlyNameField]}, idx, resultArrayMap)
		} else if action == model.Delete {
			log.Infof("deleting %v monitor", dataMap[httputil.FriendlyNameField])

			if err := service.DeleteRequest(dataMap); err != nil {
				resultArrayMap = createResultObject(model.Result{ErrorResultField: err, NameResultField: dataMap[httputil.FriendlyNameField]}, idx, resultArrayMap)
				return resultArrayMap
			}
			resultArrayMap = createResultObject(model.Result{NameResultField: dataMap[httputil.FriendlyNameField]}, idx, resultArrayMap)

		} else {
			log.Fatalf(MsgActionNotSupported, action)
		}
	}

	return resultArrayMap
}

/*
*
Populates resultMap on an index with error and monitor name.
*/
func createResultObject(result model.Result, index int, resultArrayMap []map[string]interface{}) []map[string]interface{} {
	resultArrayMap[index][model.ErrorResultField] = result.ErrorResultField
	if result.NameResultField != nil {
		resultArrayMap[index][model.MonitorNameResultField] = fmt.Sprint(result.NameResultField)
	} else {
		resultArrayMap[index][model.MonitorNameResultField] = ""

	}
	return resultArrayMap
}

/*
*
Checks if the payload supplied is valid i.e.
On update: friendly_name or id needs to be preset.
On creation: friendly_name, type, url needs to specified; if type is port monitoring port and sub_type needs to be specified; if type is keyword monitoring keyword_type and keyword_value needs to present.
*/
func (service *MonitorService) isValidPayload(dataMap map[string]interface{}, args model.Args) error {
	var validityError error = nil
	if args == model.Update {
		if dataMap[httputil.FriendlyNameField] == nil && dataMap[httputil.IdField] == nil {
			validityError = fmt.Errorf("%s or %s is needed for an update/create", httputil.FriendlyNameField, httputil.IdField)
		}
	} else {

		if dataMap[httputil.FriendlyNameField] == nil {
			validityError = fmt.Errorf(MsgFieldMissing, httputil.FriendlyNameField)
		} else {

			if dataMap[httputil.TypeField] == nil {
				validityError = fmt.Errorf(MsgFieldMissing, httputil.TypeField)
			} else {
				typeValue := fmt.Sprint(dataMap[httputil.TypeField])
				if strings.EqualFold(httputil.PortField, typeValue) && (dataMap[httputil.SubTypeField] == nil || dataMap[httputil.PortField] == nil) {
					if dataMap[httputil.SubTypeField] == nil {
						validityError = fmt.Errorf("%s monitoring requires %s", "port", httputil.SubTypeField)
					} else if dataMap[httputil.PortField] == nil {
						validityError = fmt.Errorf("%s monitoring requires %s", "port", httputil.PortField)
					}
				} else if strings.EqualFold(httputil.KeywordField, typeValue) {
					if dataMap[httputil.KeywordTypeField] == nil {
						validityError = fmt.Errorf("%s monitoring requires %s", "keyword", httputil.KeywordTypeField)
					} else if dataMap[httputil.KeywordValueField] == nil {
						validityError = fmt.Errorf("%s monitoring requires %s", "keyword", httputil.KeywordValueField)
					}
				}

				if validityError == nil {
					if dataMap[httputil.UrlField] == nil {
						validityError = fmt.Errorf(MsgFieldMissing, httputil.UrlField)
					} else if !httputil.ValidateUrl(fmt.Sprint(dataMap[httputil.UrlField])) {
						validityError = fmt.Errorf(MsgFieldIsInvalid, httputil.UrlField)
					}
				}
			}

		}
	}
	return validityError
}

/*
*
Resolves Monitor Properties:
1. Resolves alert contact friendly name to id if friendly_name otherwise it returns the id check resolveAlertContactByFriendlyName if MONITOR_RESOLVE_ALERT_CONTACTS_BY_FRIENDLY_NAME env is true
2. Resolves monitor properties mapped on staticPropertiesToResolve.
*/
func (service *MonitorService) resolveMonitorProperties(dataMap map[string]interface{}) error {
	resolveAlertContactByFriendlyName := false
	_resolveAlertContactByFriendlyName, found := service.IService.LookUpEnv(MonitorAlertContactsResolveByFriendlyNameEnv)
	if found {
		supports, err := strconv.ParseBool(_resolveAlertContactByFriendlyName)
		if err != nil {
			return err
		}
		resolveAlertContactByFriendlyName = supports
	}
	if resolveAlertContactByFriendlyName {
		attribDelimiter, found := service.IService.LookUpEnv(MonitorAlertContactsAttribDelimiterEnv)
		if !found {
			attribDelimiter = AlertContactsAttribDelimiter
		}

		delimiter, found := service.IService.LookUpEnv(MonitorAlertContactsDelimiterEnv)
		if !found {
			delimiter = AlertContactsDelimiter
		}

		if alertContacts, exists := dataMap[httputil.AlertContactsField]; exists {
			alertContactStr := fmt.Sprint(alertContacts)
			alertContactsArr := strings.Split(alertContactStr, delimiter)
			resolvedAlertContacts := make([]string, len(alertContactsArr))
			for index, value := range alertContactsArr {
				alertContact := strings.Split(value, attribDelimiter)
				alertContactFriendlyName, err := service.resolveAlertContactByFriendlyName(alertContact[0])
				if err != nil {
					return err
				}
				resolvedAlertContacts[index] = alertContactFriendlyName

				if len(alertContact) > 1 {
					resolvedAlertContacts[index] = resolvedAlertContacts[index] + AlertContactsAttribDelimiter + strings.Join(alertContact[1:], AlertContactsAttribDelimiter)
				}
			}
			resolvedAlertContactsStr := strings.Join(resolvedAlertContacts, AlertContactsDelimiter)
			dataMap[httputil.AlertContactsField] = resolvedAlertContactsStr
		}
	}
	for property, value := range dataMap {
		if _, exists := staticPropertiesToResolve[property]; exists {
			_value := fmt.Sprint(value)
			if _, exists := staticPropertiesToResolve[property][_value]; exists {
				dataMap[property] = staticPropertiesToResolve[property][_value]
			} else {
				return fmt.Errorf(MsgMappingNotFound, property, _value)
			}
		}

	}

	return nil

}

/**
Returns alert contact id if alertContactFriendlyName = alertContactFriendlyName[-] otherwise it will return an error
*/

func (service *MonitorService) resolveAlertContactByFriendlyName(alertContactFriendlyName string) (string, error) {
	pages := -1
	offset := 0
	completed := false
	requestBody := make(map[string]interface{})
	for !completed && (pages == -1 || pages > offset) {
		requestBody[httputil.OffsetField] = offset
		resultMap, err := service.IService.HttpInitiatePostRequest(httputil.GetAlertContactsEndpoint, requestBody)
		if err != nil {
			return "", err
		}
		if resultMap != nil && resultMap[httputil.TotalField] != nil && resultMap[httputil.LimitField] != nil && resultMap[httputil.AlertContactsField] != nil {
			total, err := strconv.Atoi(fmt.Sprint(resultMap[httputil.TotalField]))
			if err != nil {
				return "", err
			}
			limit, err := strconv.Atoi(fmt.Sprint(resultMap[httputil.LimitField]))
			if err != nil {
				return "", err
			}
			if total < 1 {
				return "", errors.New(MsgNoAlertContactsFound)
			} else {
				if pages == -1 {
					pages = int(math.Ceil(float64(total / limit)))
				}
				alertContacts := resultMap[httputil.AlertContactsField].([]interface{})
				for _, alertContact := range alertContacts {
					alertContact := alertContact.(map[string]interface{})
					if alertContact[httputil.FriendlyNameField] != nil && alertContact[httputil.IdField] != nil && fmt.Sprint(alertContact[httputil.FriendlyNameField]) == alertContactFriendlyName {
						alertContactFriendlyName = fmt.Sprint(alertContact[httputil.IdField])
						completed = true
						break
					}
				}

				offset++
			}

		} else {
			return "", fmt.Errorf(MsgAlertContactFetchErr, resultMap, requestBody)

		}
	}

	_, err := strconv.Atoi(alertContactFriendlyName)
	if err != nil {
		return "", fmt.Errorf(MsgAlertContactNotResolved, alertContactFriendlyName)
	}

	return alertContactFriendlyName, nil
}

func (service *MonitorService) UpdateRequest(dataMap map[string]interface{}) error {
	var monitorArr []interface{} = nil
	if dataMap[httputil.IdField] == nil && dataMap[httputil.FriendlyNameField] != nil {
		_, err := service.monitorResolvableByFriendlyName()
		if err != nil {
			return err
		}
		_monitorArr, err := service.searchMonitorByFieldMap(map[string]interface{}{
			httputil.SearchField: dataMap[httputil.FriendlyNameField],
		})
		if err != nil {
			return err
		}
		monitorArr = _monitorArr

	} else {
		_monitorArr, err := service.searchMonitorByFieldMap(map[string]interface{}{
			httputil.MonitorsField: dataMap[httputil.IdField],
		})
		if err != nil {
			return err
		}
		monitorArr = _monitorArr
		log.Infof("%v", monitorArr)

	}
	if monitorArr != nil && len(monitorArr) > 0 {
		log.Infof("%s %s %s", "updating", dataMap[httputil.FriendlyNameField], "monitor")

		originalMonitor := monitorArr[0].(map[string]interface{})
		dataMap[httputil.IdField] = originalMonitor[httputil.IdField]

		if dataMap[httputil.TypeField] != nil && fmt.Sprint(dataMap[httputil.TypeField]) != fmt.Sprint(originalMonitor[httputil.TypeField]) {
			log.Warningf(MsgOriginalAndProviderMonitorConflictType, fmt.Sprint(originalMonitor[httputil.TypeField]), fmt.Sprint(dataMap[httputil.TypeField]))
			log.Info(MsgCheckIfMonitorHasRequiredFields)

			err := service.isValidPayload(dataMap, model.Create)
			if err != nil {
				return err
			}

			_, err = service.InitiateRequest(httputil.DeleteMonitorEndpoint, originalMonitor)
			if err != nil {
				return err
			}

			_, err = service.InitiateRequest(httputil.NewMonitorEndpoint, dataMap)
			if err != nil {
				return err
			}

		} else {
			if _, err := service.InitiateRequest(httputil.EditMonitorEndpoint, dataMap); err != nil {
				return err
			}

		}

	} else {
		if err := service.CreateRequest(dataMap); err != nil {
			return err
		}
	}
	return nil
}

func (service *MonitorService) CreateRequest(dataMap map[string]interface{}) error {
	log.Infof("%s %s %s", "creating", dataMap[httputil.FriendlyNameField], "monitor")
	if _, err := service.InitiateRequest(httputil.NewMonitorEndpoint, dataMap); err != nil {
		return err
	}
	return nil
}

func (service *MonitorService) monitorResolvableByFriendlyName() (bool, error) {
	val, found := service.IService.LookUpEnv(MonitorResolveByFriendlyNameEnv)
	if found {
		parseBool, err := strconv.ParseBool(val)
		if err != nil {
			return false, err
		}

		if !parseBool {
			return false, fmt.Errorf("id field missing but one can enable %s env to update by friendly_name", MonitorResolveByFriendlyNameEnv)
		}

	}
	return true, nil
}

func (service *MonitorService) DeleteRequest(dataMap map[string]interface{}) error {
	if dataMap[httputil.IdField] != nil {
		if _, err := service.InitiateRequest(httputil.DeleteMonitorEndpoint, dataMap); err != nil {
			return err
		}
	} else if dataMap[httputil.FriendlyNameField] != nil {
		if _, err := service.monitorResolvableByFriendlyName(); err != nil {
			return err
		}

		monitorArr, err := service.searchMonitorByFieldMap(map[string]interface{}{
			httputil.SearchField: dataMap[httputil.FriendlyNameField],
		})
		if err != nil {
			return err
		}

		if monitorArr != nil && len(monitorArr) > 0 {
			remoteMonitor := monitorArr[0].(map[string]interface{})
			_, err = service.InitiateRequest(httputil.DeleteMonitorEndpoint, remoteMonitor)
			if err != nil {
				return err
			}
		} else {
			return errors.New(MsgMonitorDoesNotExist)
		}

	} else {
		return fmt.Errorf(MsgFieldMissing, "id or friendly_name")
	}
	return nil
}

func (service *MonitorService) searchMonitorByFieldMap(searchData map[string]interface{}) ([]interface{}, error) {
	resultMap, err := service.InitiateRequest(httputil.GetMonitorsEndpoint, searchData)
	if err != nil {
		return nil, err
	}
	if resultMap != nil {

		if resultMap[httputil.MonitorsField] != nil {
			monitorArr := resultMap[httputil.MonitorsField].([]interface{})
			return monitorArr, nil
		}

	}

	return nil, nil
}

func (service *MonitorService) InitiateRequest(endpoint string, data map[string]interface{}) (map[string]interface{}, error) {
	resultMap, err := service.IService.HttpInitiatePostRequest(endpoint, data)

	if err != nil {
		return nil, err
	}
	if resultMap != nil && resultMap[httputil.ErrorField] != nil {
		returnError := resultMap[httputil.ErrorField].(map[string]interface{})
		if returnError[httputil.MessageField] != nil {
			return nil, errors.New(fmt.Sprint(returnError[httputil.MessageField]))
		} else {
			return nil, fmt.Errorf(MsgUnknownErr, resultMap)
		}
	} else {
		return resultMap, nil
	}
}

func New() *MonitorService {
	monitorservice := &MonitorService{}
	monitorservice.IService = monitorservice
	return monitorservice
}

type MonitorService struct {
	service.IService
}

func (service *MonitorService) HttpInitiatePostRequest(endpoint string, dataMap map[string]interface{}) (map[string]interface{}, error) {
	return httputil.New().InitiatePostRequest(endpoint, dataMap)
}

func (service *MonitorService) LookUpEnv(variable string) (string, bool) {
	return os.LookupEnv(variable)
}
