package httputil

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/bennsimon/uptime-robot-tooling/pkg/provider"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

const (
	AlertContactsField   = "alert_contacts"
	TotalField           = "total"
	LimitField           = "limit"
	OffsetField          = "offset"
	TypeField            = "type"
	StatField            = "stat"
	SubTypeField         = "sub_type"
	PortField            = "port"
	KeywordTypeField     = "keyword_type"
	KeywordField         = "keyword"
	KeywordCaseTypeField = "keyword_case_type"
	KeywordValueField    = "keyword_value"
	MonitorsField        = "monitors"
	IdField              = "id"
	FriendlyNameField    = "friendly_name"
	UrlField             = "url"
	TypesField           = "types"
	SearchField          = "search"
	ErrorField           = "error"
	MessageField         = "message"
	ApiKeyField          = "api_key"
	FormatKeyField       = "format"
	CacheControlField    = "cache-control"
	ContentTypeField     = "content-type"
	HttpMethodField      = "http_method"
	HttpAuthTypeField    = "http_auth_type"
	PostContentTypeField = "post_content_type"
	PostValueField       = "post_value"
)

const (
	UptimeRobotApiKeyEnv = "UPTIME_ROBOT_API_KEY"
	UptimeRobotApiUrlEnv = "UPTIME_ROBOT_API_URL"
)
const (
	UptimeRobotApiUrl = "https://api.uptimerobot.com/v2/"
)

const (
	FormUrlEncodedContentType = "application/x-www-form-urlencoded"
	CacheControlValue         = "no-cache"
)

const (
	ErrorApiKeyUndefined = "API key is undefined"
)

const (
	GetMonitorsEndpoint      = "getMonitors"
	DeleteMonitorEndpoint    = "deleteMonitor"
	EditMonitorEndpoint      = "editMonitor"
	NewMonitorEndpoint       = "newMonitor"
	GetAlertContactsEndpoint = "getAlertContacts"
)

func ValidateUrl(host string) bool {
	_, err := url.ParseRequestURI(host)
	return err == nil
}

func (util *HttpUtil) InitiatePostRequest(endpoint string, dataMap map[string]interface{}) (map[string]interface{}, error) {
	return util.initiateRequest("POST", endpoint, dataMap)
}

func (util *HttpUtil) initiateRequest(method string, endpoint string, dataMap map[string]interface{}) (map[string]interface{}, error) {
	cleanPayload(dataMap)

	apiKey, found := util.IHttpUtil.LookUpEnv(UptimeRobotApiKeyEnv)
	if found {
		dataMap[ApiKeyField] = apiKey
	} else {
		return nil, errors.New(ErrorApiKeyUndefined)
	}

	uptimeRobotUrl, found := util.IHttpUtil.LookUpEnv(UptimeRobotApiUrlEnv)
	if !found {
		uptimeRobotUrl = UptimeRobotApiUrl
	}

	dataMap[FormatKeyField] = "json"
	form := url.Values{}
	for key, value := range dataMap {
		form.Add(
			key,
			fmt.Sprint(value),
		)
	}

	payload := strings.NewReader(form.Encode())
	req, err := util.IHttpUtil.newRequest(method, uptimeRobotUrl+endpoint, payload)
	if err != nil {
		return nil, err
	}
	req.Header.Add(ContentTypeField, FormUrlEncodedContentType)
	req.Header.Add(CacheControlField, CacheControlValue)
	res, err := util.IHttpUtil.makeRequest(req)

	if err != nil {
		return nil, err
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Error(err)
		}
	}(res.Body)
	body, err := io.ReadAll(res.Body)

	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, errors.New(string(body))
	}

	var resultMap map[string]interface{}
	err = json.Unmarshal(body, &resultMap)
	if err != nil {
		return nil, err
	}

	return resultMap, nil
}

func New() *HttpUtil {
	httputil := &HttpUtil{}
	httputil.IHttpUtil = httputil
	return httputil
}

func cleanPayload(dataMap map[string]interface{}) {
	delete(dataMap, FormatKeyField)
	delete(dataMap, ApiKeyField)
	delete(dataMap, HttpMethodField)
	delete(dataMap, PostValueField)
	delete(dataMap, PostContentTypeField)
}

type HttpUtil struct {
	IHttpUtil
}

func (util *HttpUtil) LookUpEnv(variable string) (string, bool) {
	return os.LookupEnv(variable)
}

func (util *HttpUtil) newRequest(method, url string, body io.Reader) (*http.Request, error) {
	return http.NewRequest(method, url, body)
}

func (util *HttpUtil) makeRequest(req *http.Request) (*http.Response, error) {
	return http.DefaultClient.Do(req)
}

var _ IHttpUtil = &HttpUtil{}
var _ provider.IConfigProvider = &HttpUtil{}

type IHttpUtil interface {
	provider.IConfigProvider
	newRequest(method, url string, body io.Reader) (*http.Request, error)
	makeRequest(req *http.Request) (*http.Response, error)
}
