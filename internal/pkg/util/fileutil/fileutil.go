package fileutil

import (
	"encoding/json"
	"errors"
	log "github.com/sirupsen/logrus"
	"os"
	"strings"
)

func isFile(payload string) bool {
	return strings.HasSuffix(strings.TrimSpace(payload), ".json")
}

func isStringJsonArray(payload string) bool {
	return strings.HasPrefix(payload, "[") && strings.HasSuffix(payload, "]")
}
func convertStringToJsonArray(payload string) string {
	return "[" + payload + "]"
}

func convertFileToString(filename string) (string, error) {
	data, err := os.ReadFile(filename)
	return string(data), err
}

func TransformInputToString(data string) (string, error) {
	data = strings.TrimSpace(data)
	if data != "" {
		if isFile(data) {
			str, err := convertFileToString(data)
			if err != nil {
				return "", err
			} else {
				data = str
			}
		}

		if !isStringJsonArray(data) {
			data = convertStringToJsonArray(data)
		}

		return data, nil
	} else {
		return "", errors.New("input is blank")
	}
}

func TransformStringToMapInterface(data string) ([]map[string]interface{}, error) {
	var dataMaps = make([]map[string]interface{}, 0)

	err := json.Unmarshal([]byte(data), &dataMaps)

	if err != nil {
		log.Error(err)
		return nil, err
	}

	return dataMaps, nil
}
