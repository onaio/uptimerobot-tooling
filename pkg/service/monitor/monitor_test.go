package monitor

import (
	"errors"
	"fmt"
	"github.com/bennsimon/uptime-robot-tooling/internal/pkg/util/httputil"
	"github.com/bennsimon/uptime-robot-tooling/pkg/model"
	"github.com/bennsimon/uptime-robot-tooling/pkg/service"
	"github.com/stretchr/testify/mock"
	"reflect"
	"testing"
)

type testMonitorService struct {
	service.IService
	mock.Mock
}

func (t *testMonitorService) HandleRequest(dataMapInterface []map[string]interface{}, action model.Args) []map[string]interface{} {
	args := t.Called(dataMapInterface, action)
	return args.Get(0).([]map[string]interface{})
}

func (t *testMonitorService) HttpInitiatePostRequest(endpoint string, dataMap map[string]interface{}) (map[string]interface{}, error) {
	args := t.Called(endpoint, dataMap)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func (t *testMonitorService) LookUpEnv(variable string) (string, bool) {
	args := t.Called(variable)
	return args.String(0), args.Bool(1)
}

func Test_createResultBody(t *testing.T) {
	type args struct {
		result         model.Result
		count          int
		resultArrayMap []map[string]interface{}
	}

	tests := []struct {
		name string
		args args
		want []map[string]interface{}
	}{

		{
			name: "should return map with a non-nil error field",
			args: args{model.Result{ErrorResultField: errors.New(""), NameResultField: ""}, 0, []map[string]interface{}{{}}},
			want: []map[string]interface{}{
				{model.ErrorResultField: errors.New(""), model.MonitorNameResultField: ""},
			},
		},
		{
			name: "should return map with a nil error field",
			args: args{model.Result{ErrorResultField: nil, NameResultField: ""}, 0, []map[string]interface{}{{}}},
			want: []map[string]interface{}{
				{model.ErrorResultField: nil, model.MonitorNameResultField: ""},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := createResultObject(tt.args.result, tt.args.count, tt.args.resultArrayMap); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("createResultObject() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMonitorService_HandleRequest(t *testing.T) {
	var testmonitorservice *testMonitorService
	monitorservice := &MonitorService{}

	type args struct {
		data     []map[string]interface{}
		argument model.Args
	}
	tests := []struct {
		name        string
		args        args
		setupMocks  func()
		verifyMocks func()
		want        []map[string]interface{}
	}{
		{name: "should return nil when empty map is provided", setupMocks: func() {
			testmonitorservice = &testMonitorService{}
			monitorservice.IService = testmonitorservice
		}, verifyMocks: func() {
			testmonitorservice.AssertExpectations(t)
		}, args: args{data: []map[string]interface{}{}, argument: model.Create}, want: []map[string]interface{}{}},
		{name: "should return nil when given JSON Object with missing fields", setupMocks: func() {
			testmonitorservice = &testMonitorService{}
			monitorservice.IService = testmonitorservice
		}, verifyMocks: func() {
			testmonitorservice.AssertExpectations(t)
		}, args: args{[]map[string]interface{}{{
			httputil.FriendlyNameField: "test",
		}}, model.Create}, want: []map[string]interface{}{{
			model.ErrorResultField:       fmt.Errorf(MsgFieldMissing, httputil.TypeField),
			model.MonitorNameResultField: "test",
		}}},
		{name: "should return result payload with nil err field", setupMocks: func() {
			testmonitorservice = &testMonitorService{}
			testmonitorservice.On("HttpInitiatePostRequest", httputil.NewMonitorEndpoint, mock.IsType(map[string]interface{}{})).Return(map[string]interface{}{}, nil)
			testmonitorservice.On("LookUpEnv", MonitorAlertContactsResolveByFriendlyNameEnv).Return("", false)
			monitorservice.IService = testmonitorservice
		}, verifyMocks: func() {
			testmonitorservice.AssertExpectations(t)
		}, args: args{[]map[string]interface{}{{
			httputil.FriendlyNameField: "test",
			httputil.TypeField:         "HTTP",
			httputil.UrlField:          "https://localhost",
		}}, model.Create}, want: []map[string]interface{}{{model.ErrorResultField: nil, model.MonitorNameResultField: "test"}}},
		{name: "should return result payload with non-nil err field", setupMocks: func() {
			testmonitorservice = &testMonitorService{}
			testmonitorservice.On("HttpInitiatePostRequest", httputil.NewMonitorEndpoint, mock.IsType(map[string]interface{}{})).Return(map[string]interface{}{}, errors.New("unknown error"))
			testmonitorservice.On("LookUpEnv", MonitorAlertContactsResolveByFriendlyNameEnv).Return("", false)
			monitorservice.IService = testmonitorservice
		}, verifyMocks: func() {
			testmonitorservice.AssertExpectations(t)
		}, args: args{[]map[string]interface{}{{
			httputil.FriendlyNameField: "test",
			httputil.TypeField:         "HTTP",
			httputil.UrlField:          "https://localhost",
		}}, model.Create}, want: []map[string]interface{}{{model.ErrorResultField: errors.New("unknown error"), model.MonitorNameResultField: "test"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()
			defer tt.verifyMocks()
			if got := monitorservice.HandleRequest(tt.args.data, tt.args.argument); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("HandleRequest() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMonitorService_isValidPayload(t *testing.T) {

	type args struct {
		dataMap   map[string]interface{}
		arguments model.Args
	}
	tests := []struct {
		name string
		args args
		want error
	}{
		{name: "should return error when with friendly_name or id is missing on update", args: args{dataMap: map[string]interface{}{
			httputil.UrlField:  "http://test.localhost",
			httputil.TypeField: "HTTP",
		}, arguments: model.Update}, want: fmt.Errorf("%s or %s is needed for an update/create", httputil.FriendlyNameField, httputil.IdField)},
		{name: "should return error when with friendly_name is missing on create", args: args{dataMap: map[string]interface{}{
			httputil.UrlField:  "http://test.localhost",
			httputil.TypeField: "HTTP",
		}, arguments: model.Create}, want: fmt.Errorf(MsgFieldMissing, httputil.FriendlyNameField)},
		{name: "should return error when with type is missing on create", args: args{dataMap: map[string]interface{}{
			httputil.FriendlyNameField: "tester",
			httputil.UrlField:          "http://test.localhost",
		}, arguments: model.Create}, want: fmt.Errorf(MsgFieldMissing, httputil.TypeField)},
		{name: "should return error in port monitoring if port is missing on create", args: args{dataMap: map[string]interface{}{
			httputil.FriendlyNameField: "tester",
			httputil.TypeField:         "Port",
			httputil.SubTypeField:      "HTTPS",
			httputil.UrlField:          "http://test.localhost",
		}, arguments: model.Create}, want: fmt.Errorf("%s monitoring requires %s", "port", httputil.PortField)},
		{name: "should return error in port monitoring if sub_type is missing on create", args: args{dataMap: map[string]interface{}{
			httputil.FriendlyNameField: "tester",
			httputil.TypeField:         "Port",
			httputil.PortField:         "8080",
			httputil.UrlField:          "http://test.localhost",
		}, arguments: model.Create}, want: fmt.Errorf("%s monitoring requires %s", "port", httputil.SubTypeField)},
		{name: "should return error in keyword monitoring if keyword_type field is missing on create", args: args{dataMap: map[string]interface{}{
			httputil.FriendlyNameField: "tester",
			httputil.TypeField:         "Keyword",
			httputil.KeywordValueField: "welcome",
			httputil.UrlField:          "http://test.localhost",
		}, arguments: model.Create}, want: fmt.Errorf("%s monitoring requires %s", "keyword", httputil.KeywordTypeField)},
		{name: "should return error in keyword monitoring if keyword_value field is missing on create", args: args{dataMap: map[string]interface{}{
			httputil.FriendlyNameField: "tester",
			httputil.TypeField:         "Keyword",
			httputil.KeywordTypeField:  "exists",
			httputil.UrlField:          "http://test.localhost",
		}, arguments: model.Create}, want: fmt.Errorf("%s monitoring requires %s", "keyword", httputil.KeywordValueField)},
		{name: "should return error if url field is missing on create", args: args{dataMap: map[string]interface{}{
			httputil.FriendlyNameField: "tester",
			httputil.TypeField:         "Keyword",
			httputil.KeywordTypeField:  "exists",
			httputil.KeywordValueField: "welcome",
		}, arguments: model.Create}, want: fmt.Errorf(MsgFieldMissing, httputil.UrlField)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			monitorservice := &MonitorService{}
			got := monitorservice.isValidPayload(tt.args.dataMap, tt.args.arguments)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("isValidPayload() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMonitorService_UpdateRequest(t *testing.T) {
	var testmonitorservice *testMonitorService
	monitorservice := &MonitorService{}

	type args struct {
		dataMap map[string]interface{}
	}
	tests := []struct {
		name        string
		args        args
		setupMocks  func()
		verifyMocks func()
		wantErr     bool
		wantDataMap map[string]interface{}
	}{
		{name: "should resolve monitor by friendly_name", setupMocks: func() {
			testmonitorservice = &testMonitorService{}
			testmonitorservice.On("HttpInitiatePostRequest", httputil.GetMonitorsEndpoint, mock.IsType(map[string]interface{}{})).Return(map[string]interface{}{
				httputil.MonitorsField: []interface{}{
					map[string]interface{}{
						httputil.FriendlyNameField: "tester",
						httputil.TypeField:         "HTTP",
						httputil.IdField:           "3",
					},
				},
			}, nil)
			testmonitorservice.On("HttpInitiatePostRequest", httputil.EditMonitorEndpoint, mock.IsType(map[string]interface{}{})).Return(map[string]interface{}{}, nil)
			testmonitorservice.On("LookUpEnv", MonitorResolveByFriendlyNameEnv).Return("", false)
			monitorservice.IService = testmonitorservice
		}, verifyMocks: func() {
			testmonitorservice.AssertExpectations(t)
		}, args: args{dataMap: map[string]interface{}{
			httputil.FriendlyNameField: "tester",
			httputil.UrlField:          "https://localhost",
			httputil.TypeField:         "HTTP",
		}}, wantDataMap: map[string]interface{}{
			httputil.FriendlyNameField: "tester",
			httputil.UrlField:          "https://localhost",
			httputil.TypeField:         "HTTP",
			httputil.IdField:           "3",
		}, wantErr: false},
		{name: "should fail to resolve monitor by friendly_name when config set to false and return an error", setupMocks: func() {
			testmonitorservice = &testMonitorService{}
			testmonitorservice.On("LookUpEnv", MonitorResolveByFriendlyNameEnv).Return("false", true)
			monitorservice.IService = testmonitorservice
		}, verifyMocks: func() {
			testmonitorservice.AssertExpectations(t)
		}, args: args{dataMap: map[string]interface{}{
			httputil.FriendlyNameField: "tester",
			httputil.UrlField:          "https://localhost",
			httputil.TypeField:         "HTTP",
		}}, wantDataMap: map[string]interface{}{
			httputil.FriendlyNameField: "tester",
			httputil.UrlField:          "https://localhost",
			httputil.TypeField:         "HTTP",
		}, wantErr: true},
		{name: "should resolve monitor by id without an error", setupMocks: func() {
			testmonitorservice = &testMonitorService{}
			testmonitorservice.On("HttpInitiatePostRequest", httputil.GetMonitorsEndpoint, mock.IsType(map[string]interface{}{})).Return(map[string]interface{}{
				httputil.MonitorsField: []interface{}{
					map[string]interface{}{
						httputil.FriendlyNameField: "tester",
						httputil.TypeField:         "HTTP",
						httputil.IdField:           "3",
					},
				},
			}, nil)
			testmonitorservice.On("HttpInitiatePostRequest", httputil.EditMonitorEndpoint, mock.IsType(map[string]interface{}{})).Return(map[string]interface{}{}, nil)
			monitorservice.IService = testmonitorservice

		}, verifyMocks: func() {
			testmonitorservice.AssertExpectations(t)
		}, args: args{dataMap: map[string]interface{}{
			httputil.FriendlyNameField: "tester",
			httputil.UrlField:          "https://localhost",
			httputil.TypeField:         "HTTP",
			httputil.IdField:           "3",
		}}, wantDataMap: map[string]interface{}{
			httputil.FriendlyNameField: "tester",
			httputil.UrlField:          "https://localhost",
			httputil.TypeField:         "HTTP",
			httputil.IdField:           "3",
		}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()
			defer tt.verifyMocks()
			if err := monitorservice.UpdateRequest(tt.args.dataMap); (err != nil) != tt.wantErr {
				t.Errorf("UpdateRequest() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantDataMap != nil && !reflect.DeepEqual(tt.wantDataMap, tt.args.dataMap) {
				t.Errorf("UpdateRequest() got = %v, want %v", tt.args.dataMap, tt.wantDataMap)

			}
		})
	}
}

func TestMonitorService_resolveAlertContactByFriendlyName(t *testing.T) {
	var testmonitorservice *testMonitorService
	monitorservice := &MonitorService{}

	type args struct {
		alertContactFriendlyName string
	}
	tests := []struct {
		name        string
		args        args
		setupMocks  func()
		verifyMocks func()
		want        string
		wantErr     bool
	}{
		{name: "should fail to resolve alertContactByFriendlyName", setupMocks: func() {
			testmonitorservice = &testMonitorService{}
			testmonitorservice.On("HttpInitiatePostRequest", httputil.GetAlertContactsEndpoint, mock.IsType(map[string]interface{}{})).Return(map[string]interface{}{
				"stat":  "ok",
				"limit": float64(1),
				"total": float64(2),
				"alert_contacts": []interface{}{
					map[string]interface{}{
						"id":            "1",
						"friendly_name": "test@email.com",
					},
				},
			}, nil)

			monitorservice.IService = testmonitorservice
		}, verifyMocks: func() {
			testmonitorservice.AssertExpectations(t)
		}, args: args{alertContactFriendlyName: "test@email.com"}, want: "1", wantErr: false},
		{name: "should resolve alertContactByFriendlyName after pagination", setupMocks: func() {
			testmonitorservice = &testMonitorService{}
			testmonitorservice.On("HttpInitiatePostRequest", httputil.GetAlertContactsEndpoint, map[string]interface{}{
				httputil.OffsetField: 0,
			}).Return(map[string]interface{}{
				"stat":  "ok",
				"limit": float64(1),
				"total": float64(2),
				"alert_contacts": []interface{}{
					map[string]interface{}{
						"id":            "1",
						"friendly_name": "tester@email.com",
					},
				},
			}, nil)
			testmonitorservice.On("HttpInitiatePostRequest", httputil.GetAlertContactsEndpoint, map[string]interface{}{
				httputil.OffsetField: 1,
			}).Return(map[string]interface{}{
				"stat":  "ok",
				"limit": float64(1),
				"total": float64(2),
				"alert_contacts": []interface{}{
					map[string]interface{}{
						"id":            "2",
						"friendly_name": "test@email.com",
					},
				},
			}, nil)

			monitorservice.IService = testmonitorservice
		}, verifyMocks: func() {
			testmonitorservice.AssertExpectations(t)
		}, args: args{alertContactFriendlyName: "test@email.com"}, want: "2", wantErr: false},
		{name: "should fail to resolve alertContact by FriendlyName", setupMocks: func() {
			testmonitorservice = &testMonitorService{}
			testmonitorservice.On("HttpInitiatePostRequest", httputil.GetAlertContactsEndpoint, map[string]interface{}{
				httputil.OffsetField: 0,
			}).Return(map[string]interface{}{
				"stat":  "ok",
				"limit": float64(1),
				"total": float64(2),
				"alert_contacts": []interface{}{
					map[string]interface{}{
						"id":            "1",
						"friendly_name": "tester@email.com",
					},
				},
			}, nil)
			testmonitorservice.On("HttpInitiatePostRequest", httputil.GetAlertContactsEndpoint, map[string]interface{}{
				httputil.OffsetField: 1,
			}).Return(map[string]interface{}{
				"stat":  "ok",
				"limit": float64(1),
				"total": float64(2),
				"alert_contacts": []interface{}{
					map[string]interface{}{
						"id":            "2",
						"friendly_name": "testsd@email.com",
					},
				},
			}, nil)

			monitorservice.IService = testmonitorservice
		}, verifyMocks: func() {
			testmonitorservice.AssertExpectations(t)
		}, args: args{alertContactFriendlyName: "test@email.com"}, want: "", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()
			defer tt.verifyMocks()
			got, err := monitorservice.resolveAlertContactByFriendlyName(tt.args.alertContactFriendlyName)
			if (err != nil) != tt.wantErr {
				t.Errorf("resolveAlertContactByFriendlyName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("resolveAlertContactByFriendlyName() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMonitorService_resolveMonitorProperties(t *testing.T) {
	var testmonitorservice *testMonitorService

	monitorservice := &MonitorService{}
	type args struct {
		dataMap map[string]interface{}
	}
	tests := []struct {
		name        string
		args        args
		setupMocks  func()
		verifyMocks func()
		wantErr     bool
		mapResult   map[string]interface{}
	}{
		{name: "should resolve alert contact correctly", setupMocks: func() {
			testmonitorservice = &testMonitorService{}
			testmonitorservice.On("LookUpEnv", MonitorAlertContactsResolveByFriendlyNameEnv).Return("true", true)
			testmonitorservice.On("LookUpEnv", MonitorAlertContactsAttribDelimiterEnv).Return("_", true)
			testmonitorservice.On("LookUpEnv", MonitorAlertContactsDelimiterEnv).Return("-", true)
			testmonitorservice.On("HttpInitiatePostRequest", httputil.GetAlertContactsEndpoint, mock.IsType(map[string]interface{}{})).Return(map[string]interface{}{
				"stat":  "ok",
				"limit": float64(1),
				"total": float64(2),
				"alert_contacts": []interface{}{
					map[string]interface{}{
						"id":            "1",
						"friendly_name": "tester@email.com",
					},
				},
			}, nil)

			monitorservice.IService = testmonitorservice
		}, verifyMocks: func() {
			testmonitorservice.AssertExpectations(t)
		}, args: args{dataMap: map[string]interface{}{
			httputil.AlertContactsField: "tester@email.com",
			httputil.FriendlyNameField:  "tester@email.com",
			httputil.UrlField:           "tester.localhost",
		}}, wantErr: false, mapResult: map[string]interface{}{
			httputil.AlertContactsField: "1",
			httputil.FriendlyNameField:  "tester@email.com",
			httputil.UrlField:           "tester.localhost",
		}},
		{name: "should resolve alert contact with attrib correctly(friendly_name used)", setupMocks: func() {
			testmonitorservice = &testMonitorService{}
			testmonitorservice.On("LookUpEnv", MonitorAlertContactsResolveByFriendlyNameEnv).Return("true", true)
			testmonitorservice.On("LookUpEnv", MonitorAlertContactsAttribDelimiterEnv).Return("_", true)
			testmonitorservice.On("LookUpEnv", MonitorAlertContactsDelimiterEnv).Return("-", true)
			testmonitorservice.On("HttpInitiatePostRequest", httputil.GetAlertContactsEndpoint, mock.IsType(map[string]interface{}{})).Return(map[string]interface{}{
				"stat":  "ok",
				"limit": float64(1),
				"total": float64(2),
				"alert_contacts": []interface{}{
					map[string]interface{}{
						"id":            "1",
						"friendly_name": "tester@email.com",
					},
				},
			}, nil)

			monitorservice.IService = testmonitorservice
		}, verifyMocks: func() {
			testmonitorservice.AssertExpectations(t)
		}, args: args{dataMap: map[string]interface{}{
			httputil.AlertContactsField: "tester@email.com_9_9",
			httputil.FriendlyNameField:  "tester@email.com",
			httputil.UrlField:           "tester.localhost",
		}}, wantErr: false, mapResult: map[string]interface{}{
			httputil.AlertContactsField: "1_9_9",
			httputil.FriendlyNameField:  "tester@email.com",
			httputil.UrlField:           "tester.localhost",
		}},
		{name: "should resolve alert contact with attrib correctly (id used)", setupMocks: func() {
			testmonitorservice = &testMonitorService{}
			testmonitorservice.On("LookUpEnv", MonitorAlertContactsResolveByFriendlyNameEnv).Return("true", true)
			testmonitorservice.On("LookUpEnv", MonitorAlertContactsAttribDelimiterEnv).Return("_", true)
			testmonitorservice.On("LookUpEnv", MonitorAlertContactsDelimiterEnv).Return("-", true)
			testmonitorservice.On("HttpInitiatePostRequest", httputil.GetAlertContactsEndpoint, mock.IsType(map[string]interface{}{})).Return(map[string]interface{}{
				"stat":  "ok",
				"limit": float64(1),
				"total": float64(2),
				"alert_contacts": []interface{}{
					map[string]interface{}{
						"id":            "1",
						"friendly_name": "tester@email.com",
					},
				},
			}, nil)

			monitorservice.IService = testmonitorservice
		}, verifyMocks: func() {
			testmonitorservice.AssertExpectations(t)
		}, args: args{dataMap: map[string]interface{}{
			httputil.AlertContactsField: "1_9_9",
			httputil.FriendlyNameField:  "tester@email.com",
			httputil.UrlField:           "tester.localhost",
		}}, wantErr: false, mapResult: map[string]interface{}{
			httputil.AlertContactsField: "1_9_9",
			httputil.FriendlyNameField:  "tester@email.com",
			httputil.UrlField:           "tester.localhost",
		}},
		{name: "Should Resolve Alert Contacts With Attrib Correctly", setupMocks: func() {
			testmonitorservice = &testMonitorService{}
			testmonitorservice.On("LookUpEnv", MonitorAlertContactsResolveByFriendlyNameEnv).Return("true", true)
			testmonitorservice.On("LookUpEnv", MonitorAlertContactsAttribDelimiterEnv).Return("_", true)
			testmonitorservice.On("LookUpEnv", MonitorAlertContactsDelimiterEnv).Return("-", true)
			testmonitorservice.On("HttpInitiatePostRequest", httputil.GetAlertContactsEndpoint, mock.IsType(map[string]interface{}{})).Return(map[string]interface{}{
				"stat":  "ok",
				"limit": float64(1),
				"total": float64(2),
				"alert_contacts": []interface{}{
					map[string]interface{}{
						"id":            "1",
						"friendly_name": "tester@email.com",
					},
					map[string]interface{}{
						"id":            "2",
						"friendly_name": "testerr@email.com",
					},
				},
			}, nil)

			monitorservice.IService = testmonitorservice
		}, verifyMocks: func() {
			testmonitorservice.AssertExpectations(t)
		}, args: args{dataMap: map[string]interface{}{
			httputil.AlertContactsField: "tester@email.com_9_8-testerr@email.com_7_3",
			httputil.FriendlyNameField:  "tester@email.com",
			httputil.UrlField:           "tester.localhost",
		}}, wantErr: false, mapResult: map[string]interface{}{
			httputil.AlertContactsField: "1_9_8-2_7_3",
			httputil.FriendlyNameField:  "tester@email.com",
			httputil.UrlField:           "tester.localhost",
		}},
		{name: "should fail to resolve properties (alertContact not found)", setupMocks: func() {
			testmonitorservice = &testMonitorService{}
			testmonitorservice.On("LookUpEnv", MonitorAlertContactsResolveByFriendlyNameEnv).Return("true", true)
			testmonitorservice.On("LookUpEnv", MonitorAlertContactsAttribDelimiterEnv).Return("_", true)
			testmonitorservice.On("LookUpEnv", MonitorAlertContactsDelimiterEnv).Return("-", true)
			testmonitorservice.On("HttpInitiatePostRequest", httputil.GetAlertContactsEndpoint, mock.IsType(map[string]interface{}{})).Return(map[string]interface{}{
				"stat":  "ok",
				"limit": float64(1),
				"total": float64(2),
				"alert_contacts": []interface{}{
					map[string]interface{}{
						"id":            "1",
						"friendly_name": "tester@email.com",
					},
				},
			}, nil)

			monitorservice.IService = testmonitorservice
		}, verifyMocks: func() {
			testmonitorservice.AssertExpectations(t)
		}, args: args{dataMap: map[string]interface{}{
			httputil.AlertContactsField: "testerer@email.com",
			httputil.FriendlyNameField:  "tester@email.com",
			httputil.UrlField:           "tester.localhost",
		}}, wantErr: true},
		{name: "should fail to resolve properties when alert contacts are not present but set on payload", setupMocks: func() {
			testmonitorservice = &testMonitorService{}
			testmonitorservice.On("LookUpEnv", MonitorAlertContactsResolveByFriendlyNameEnv).Return("true", true)
			testmonitorservice.On("LookUpEnv", MonitorAlertContactsAttribDelimiterEnv).Return("_", true)
			testmonitorservice.On("LookUpEnv", MonitorAlertContactsDelimiterEnv).Return("-", true)
			testmonitorservice.On("HttpInitiatePostRequest", httputil.GetAlertContactsEndpoint, mock.IsType(map[string]interface{}{})).Return(map[string]interface{}{
				"stat":           "ok",
				"limit":          float64(1),
				"total":          float64(0),
				"alert_contacts": []interface{}{},
			}, nil)

			monitorservice.IService = testmonitorservice
		}, verifyMocks: func() {
			testmonitorservice.AssertExpectations(t)
		}, args: args{dataMap: map[string]interface{}{
			httputil.AlertContactsField: "tester@email.com",
			httputil.FriendlyNameField:  "tester@email.com",
			httputil.UrlField:           "tester.localhost",
		}}, wantErr: true},
		{name: "should fail to resolve properties when response is missing total property", setupMocks: func() {
			testmonitorservice = &testMonitorService{}
			testmonitorservice.On("LookUpEnv", MonitorAlertContactsResolveByFriendlyNameEnv).Return("true", true)
			testmonitorservice.On("LookUpEnv", MonitorAlertContactsAttribDelimiterEnv).Return("_", true)
			testmonitorservice.On("LookUpEnv", MonitorAlertContactsDelimiterEnv).Return("-", true)
			testmonitorservice.On("HttpInitiatePostRequest", httputil.GetAlertContactsEndpoint, mock.IsType(map[string]interface{}{})).Return(map[string]interface{}{
				"stat":           "ok",
				"limit":          float64(1),
				"alert_contacts": []interface{}{},
			}, nil)

			monitorservice.IService = testmonitorservice
		}, verifyMocks: func() {
			testmonitorservice.AssertExpectations(t)
		}, args: args{dataMap: map[string]interface{}{
			httputil.AlertContactsField: "tester@email.com",
			httputil.FriendlyNameField:  "tester@email.com",
			httputil.UrlField:           "tester.localhost",
		}}, wantErr: true},
		{name: "should fail to resolve properties when static property mapping is absent and missing key set ", setupMocks: func() {
			testmonitorservice = &testMonitorService{}
			testmonitorservice.On("LookUpEnv", MonitorAlertContactsResolveByFriendlyNameEnv).Return("true", true)
			testmonitorservice.On("LookUpEnv", MonitorAlertContactsAttribDelimiterEnv).Return("_", true)
			testmonitorservice.On("LookUpEnv", MonitorAlertContactsDelimiterEnv).Return("-", true)
			testmonitorservice.On("HttpInitiatePostRequest", httputil.GetAlertContactsEndpoint, mock.IsType(map[string]interface{}{})).Return(map[string]interface{}{
				"stat":  "ok",
				"limit": float64(1),
				"total": float64(0),
				"alert_contacts": []interface{}{
					map[string]interface{}{
						"id":            "1",
						"friendly_name": "tester@email.com",
					},
				},
			}, nil)

			monitorservice.IService = testmonitorservice
		}, verifyMocks: func() {
			testmonitorservice.AssertExpectations(t)
		}, args: args{dataMap: map[string]interface{}{
			httputil.AlertContactsField: "tester@email.com",
			httputil.FriendlyNameField:  "tester@email.com",
			httputil.UrlField:           "tester.localhost",
			httputil.TypeField:          "gRPC",
		}}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()
			defer tt.verifyMocks()
			if err := monitorservice.resolveMonitorProperties(tt.args.dataMap); (err != nil) != tt.wantErr {
				t.Errorf("resolveMonitorProperties() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.mapResult != nil && !(reflect.DeepEqual(tt.mapResult, tt.args.dataMap)) {
				t.Errorf("resolveMonitorProperties() args = %v, mapResult %v", tt.mapResult, tt.args.dataMap)
			}
		})
	}
}

func TestMonitorService_DeleteRequest(t *testing.T) {
	var testmonitorservice *testMonitorService

	monitorservice := &MonitorService{}

	type args struct {
		dataMap map[string]interface{}
	}
	tests := []struct {
		name        string
		args        args
		setupMocks  func()
		verifyMocks func()
		wantErr     bool
	}{
		{name: "should return error when id or friendly_name field is absent", setupMocks: func() {
			testmonitorservice = &testMonitorService{}
			//testmonitorservice.On("HttpInitiatePostRequest", mock.Anything, mock.IsType(map[string]interface{}{})).Return(map[string]interface{}{}, nil)
			monitorservice.IService = testmonitorservice
		}, verifyMocks: func() {
			testmonitorservice.AssertExpectations(t)
		}, args: args{dataMap: map[string]interface{}{}}, wantErr: true},
		{name: "should not return error when id is present", setupMocks: func() {
			testmonitorservice = &testMonitorService{}
			testmonitorservice.On("HttpInitiatePostRequest", mock.Anything, mock.IsType(map[string]interface{}{})).Return(map[string]interface{}{}, nil)
			monitorservice.IService = testmonitorservice
		}, verifyMocks: func() {
			testmonitorservice.AssertExpectations(t)
		}, args: args{dataMap: map[string]interface{}{
			httputil.IdField: "2",
		}}, wantErr: false},
		{name: "should not return error when friendly_name is present", setupMocks: func() {
			testmonitorservice = &testMonitorService{}
			testmonitorservice.On("HttpInitiatePostRequest", httputil.DeleteMonitorEndpoint, mock.IsType(map[string]interface{}{})).Return(map[string]interface{}{}, nil)
			testmonitorservice.On("HttpInitiatePostRequest", httputil.GetMonitorsEndpoint, mock.IsType(map[string]interface{}{})).Return(map[string]interface{}{
				"monitors": []interface{}{
					map[string]interface{}{
						httputil.FriendlyNameField: "test",
					},
				},
			}, nil)
			testmonitorservice.On("LookUpEnv", mock.Anything).Return("true", true)
			monitorservice.IService = testmonitorservice
		}, verifyMocks: func() {
			testmonitorservice.AssertExpectations(t)
		}, args: args{dataMap: map[string]interface{}{
			httputil.FriendlyNameField: "test",
		}}, wantErr: false},
		{name: "should return error when friendly name monitor does not exist", setupMocks: func() {
			testmonitorservice = &testMonitorService{}
			testmonitorservice.On("HttpInitiatePostRequest", httputil.GetMonitorsEndpoint, mock.IsType(map[string]interface{}{})).Return(map[string]interface{}{
				"monitors": []interface{}{},
			}, nil)
			testmonitorservice.On("LookUpEnv", mock.Anything).Return("", false)
			monitorservice.IService = testmonitorservice
		}, verifyMocks: func() {
			testmonitorservice.AssertExpectations(t)
		}, args: args{dataMap: map[string]interface{}{
			httputil.FriendlyNameField: "test",
		}}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()
			defer tt.verifyMocks()
			if err := monitorservice.DeleteRequest(tt.args.dataMap); (err != nil) != tt.wantErr {
				t.Errorf("DeleteRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMonitorService_searchMonitorByFieldMap(t *testing.T) {
	var testmonitorservice *testMonitorService

	monitorservice := &MonitorService{}
	type args struct {
		searchData map[string]interface{}
	}
	tests := []struct {
		name        string
		args        args
		setupMocks  func()
		verifyMocks func()
		want        []interface{}
		wantErr     bool
	}{
		{name: "should return nil with no error with empty resultmap", setupMocks: func() {
			testmonitorservice = &testMonitorService{}
			testmonitorservice.On("HttpInitiatePostRequest", mock.Anything, mock.IsType(map[string]interface{}{})).Return(nil, nil)
			monitorservice.IService = testmonitorservice
		}, verifyMocks: func() {
			testmonitorservice.AssertExpectations(t)
		}, args: args{}, want: nil, wantErr: false},
		{name: "should return empty array payload with no error with empty monitors array", setupMocks: func() {
			testmonitorservice = &testMonitorService{}
			testmonitorservice.On("HttpInitiatePostRequest", mock.Anything, mock.IsType(map[string]interface{}{})).Return(map[string]interface{}{
				httputil.MonitorsField: []interface{}{},
			}, nil)
			monitorservice.IService = testmonitorservice
		}, verifyMocks: func() {
			testmonitorservice.AssertExpectations(t)
		}, args: args{}, want: []interface{}{}, wantErr: false},
		{name: "should return non empty array with no error with non empty monitors array", setupMocks: func() {
			testmonitorservice = &testMonitorService{}
			testmonitorservice.On("HttpInitiatePostRequest", mock.Anything, mock.IsType(map[string]interface{}{})).Return(map[string]interface{}{
				"monitors": []interface{}{
					map[string]interface{}{
						httputil.FriendlyNameField: "test",
					},
				},
			}, nil)
			monitorservice.IService = testmonitorservice
		}, verifyMocks: func() {
			testmonitorservice.AssertExpectations(t)
		}, args: args{}, want: []interface{}{
			map[string]interface{}{
				httputil.FriendlyNameField: "test",
			},
		}, wantErr: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()
			defer tt.verifyMocks()
			got, err := monitorservice.searchMonitorByFieldMap(tt.args.searchData)
			if (err != nil) != tt.wantErr {
				t.Errorf("searchMonitorByFieldMap() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("searchMonitorByFieldMap() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMonitorService_InitiateRequest(t *testing.T) {
	var testmonitorservice *testMonitorService
	type args struct {
		endpoint string
		data     map[string]interface{}
	}
	monitorservice := &MonitorService{}

	tests := []struct {
		name        string
		args        args
		setupMocks  func()
		verifyMocks func()
		want        map[string]interface{}
		wantErr     bool
	}{
		{name: "should return error when request returns error", setupMocks: func() {
			testmonitorservice = &testMonitorService{}
			testmonitorservice.On("HttpInitiatePostRequest", mock.Anything, mock.IsType(map[string]interface{}{})).Return(nil, errors.New("error"))
			monitorservice.IService = testmonitorservice
		}, verifyMocks: func() {
			testmonitorservice.AssertExpectations(t)
		}, args: args{}, want: nil, wantErr: true},
		{name: "should returns no error when request returns no error", setupMocks: func() {
			testmonitorservice = &testMonitorService{}
			testmonitorservice.On("HttpInitiatePostRequest", mock.Anything, mock.IsType(map[string]interface{}{})).Return(map[string]interface{}{}, nil)
			monitorservice.IService = testmonitorservice
		}, verifyMocks: func() {
			testmonitorservice.AssertExpectations(t)
		}, args: args{}, want: map[string]interface{}{}, wantErr: false},
		{name: "should return error when error returns error with message field", setupMocks: func() {
			testmonitorservice = &testMonitorService{}
			testmonitorservice.On("HttpInitiatePostRequest", mock.Anything, mock.IsType(map[string]interface{}{})).Return(map[string]interface{}{
				httputil.ErrorField: map[string]interface{}{
					httputil.MessageField: "message x",
				},
			}, nil)
			monitorservice.IService = testmonitorservice
		}, verifyMocks: func() {
			testmonitorservice.AssertExpectations(t)
		}, args: args{}, want: nil, wantErr: true},
		{name: "should return error with no message field", setupMocks: func() {
			testmonitorservice = &testMonitorService{}
			testmonitorservice.On("HttpInitiatePostRequest", mock.Anything, mock.IsType(map[string]interface{}{})).Return(map[string]interface{}{
				httputil.ErrorField: map[string]interface{}{},
			}, nil)
			monitorservice.IService = testmonitorservice
		}, verifyMocks: func() {
			testmonitorservice.AssertExpectations(t)
		}, args: args{}, want: nil, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()
			defer tt.verifyMocks()
			got, err := monitorservice.InitiateRequest(tt.args.endpoint, tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("InitiateRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("InitiateRequest() got = %v, want %v", got, tt.want)
			}
		})
	}
}
