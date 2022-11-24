package service

import (
	"github.com/bennsimon/uptimerobot-tooling/pkg/model"
	"github.com/bennsimon/uptimerobot-tooling/pkg/provider"
)

type IService interface {
	provider.IConfigProvider
	HttpInitiatePostRequest(endpoint string, dataMap map[string]interface{}) (map[string]interface{}, error)
	HandleRequest(dataMapInterface []map[string]interface{}, action model.Args) []map[string]interface{}
}

type NopService struct{}

var _ IService = &NopService{}

func (s *NopService) LookUpEnv(variable string) (string, bool) {
	return "", false
}

func (s *NopService) HttpInitiatePostRequest(endpoint string, dataMap map[string]interface{}) (map[string]interface{}, error) {
	return map[string]interface{}{}, nil
}

func (s *NopService) HandleRequest([]map[string]interface{}, model.Args) []map[string]interface{} {
	return []map[string]interface{}{}
}
