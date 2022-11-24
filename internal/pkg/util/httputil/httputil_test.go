package httputil

import (
	"errors"
	"github.com/stretchr/testify/mock"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

type testHttpUtil struct {
	mock.Mock
}

func (t *testHttpUtil) LookUpEnv(variable string) (string, bool) {
	args := t.Called(variable)
	return args.String(0), args.Bool(1)
}

func (t *testHttpUtil) newRequest(method, url string, body io.Reader) (*http.Request, error) {
	args := t.Called(method, url, body)
	return args.Get(0).(*http.Request), args.Error(1)
}

func (t *testHttpUtil) makeRequest(req *http.Request) (*http.Response, error) {
	args := t.Called(req)
	return args.Get(0).(*http.Response), args.Error(1)
}

var _ IHttpUtil = &testHttpUtil{}

func TestInitiatePostRequest(t *testing.T) {
	var testutil *testHttpUtil
	var httputil = &HttpUtil{}
	type args struct {
		endpoint string
		dataMap  map[string]interface{}
	}
	tests := []struct {
		name        string
		args        args
		setupMocks  func()
		verifyMocks func()
		want        map[string]interface{}
		wantErr     bool
	}{
		{name: "should fail if API key is undefined", setupMocks: func() {
			testutil = &testHttpUtil{}
			testutil.On("LookUpEnv", UptimeRobotApiKeyEnv).Return("", false)
			httputil.IHttpUtil = testutil

		}, verifyMocks: func() {
			testutil.AssertExpectations(t)
		}, args: args{dataMap: map[string]interface{}{}, endpoint: ""}, want: nil, wantErr: true},
		{name: "should return error string when newRequest returns error", args: args{dataMap: map[string]interface{}{}, endpoint: ""},
			setupMocks: func() {
				testutil = &testHttpUtil{}
				testutil.On("LookUpEnv", UptimeRobotApiKeyEnv).Return("api-key", true)
				testutil.On("LookUpEnv", UptimeRobotApiUrlEnv).Return("https://localhost", true)
				res := httptest.NewRequest("POST", "https://localhost", strings.NewReader("api_key=api-key&format=json"))
				testutil.On("newRequest", "POST", "https://localhost", strings.NewReader("api_key=api-key&format=json")).Return(res, errors.New("test error"))
				httputil.IHttpUtil = testutil

			}, verifyMocks: func() {
				testutil.AssertExpectations(t)
			},
			want: nil, wantErr: true},
		{name: "should return error string for non 200 response", args: args{dataMap: map[string]interface{}{}, endpoint: ""}, setupMocks: func() {
			testutil = &testHttpUtil{}
			testutil.On("LookUpEnv", UptimeRobotApiKeyEnv).Return("api-key", true)
			testutil.On("LookUpEnv", UptimeRobotApiUrlEnv).Return("https://localhost", true)
			res := httptest.NewRequest("POST", "https://localhost", strings.NewReader("api_key=api-key&format=json"))
			testutil.On("newRequest", "POST", "https://localhost", strings.NewReader("api_key=api-key&format=json")).Return(res, nil)
			testutil.On("makeRequest", res).Return(&http.Response{Body: io.NopCloser(strings.NewReader("{}")), StatusCode: http.StatusBadRequest}, nil)
			httputil.IHttpUtil = testutil

		}, verifyMocks: func() {
			testutil.AssertExpectations(t)
		}, want: nil, wantErr: true},
		{name: "should return resultMap with no error given valid inputs", args: args{dataMap: map[string]interface{}{}, endpoint: ""}, setupMocks: func() {
			testutil = &testHttpUtil{}
			testutil.On("LookUpEnv", UptimeRobotApiKeyEnv).Return("api-key", true)
			testutil.On("LookUpEnv", UptimeRobotApiUrlEnv).Return("https://localhost", true)
			res := httptest.NewRequest("POST", "https://localhost", strings.NewReader("api_key=api-key&format=json"))
			testutil.On("newRequest", "POST", "https://localhost", strings.NewReader("api_key=api-key&format=json")).Return(res, nil)
			testutil.On("makeRequest", res).Return(&http.Response{Body: io.NopCloser(strings.NewReader("{}")), StatusCode: http.StatusOK}, nil)
			httputil.IHttpUtil = testutil

		}, verifyMocks: func() {
			testutil.AssertExpectations(t)
		}, want: map[string]interface{}{}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()
			defer tt.verifyMocks()
			got, err := httputil.InitiatePostRequest(tt.args.endpoint, tt.args.dataMap)
			if (err != nil) != tt.wantErr {
				t.Errorf("InitiatePostRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("InitiatePostRequest() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidateUrl(t *testing.T) {
	type args struct {
		host string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{name: "should return true for valid url", args: args{host: "http://google.com"}, want: true},
		{name: "should return error on invalid", args: args{host: "ht:///localhost-"}, want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ValidateUrl(tt.args.host); got != tt.want {
				t.Errorf("ValidateUrl() error = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_cleanPayload(t *testing.T) {
	type args struct {
		dataMap map[string]interface{}
	}
	tests := []struct {
		name string
		args args
		want map[string]interface{}
	}{
		{name: "should remove all unwanted fields", args: args{dataMap: map[string]interface{}{
			FormatKeyField:       "xml",
			ApiKeyField:          "api-key",
			HttpMethodField:      "POST",
			PostValueField:       "test",
			PostContentTypeField: "application/xml",
		}}, want: map[string]interface{}{}},
	}
	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			cleanPayload(tt.args.dataMap)
		})

		if tt.want != nil && !reflect.DeepEqual(tt.args.dataMap, tt.want) {
			t.Errorf("cleanPayload() got = %v, want %v", tt.args.dataMap, tt.want)
		}
	}
}
