package handler

import (
	"fmt"
	"github.com/bennsimon/uptime-robot-tooling/internal/pkg/util/httputil"
	"github.com/bennsimon/uptime-robot-tooling/pkg/model"
	"github.com/bennsimon/uptime-robot-tooling/pkg/service/monitor"
	"reflect"
	"testing"
)

func TestHandleRequest(t *testing.T) {
	type args struct {
		payload  string
		resource string
		action   string
	}
	tests := []struct {
		name string
		args args
		want []map[string]interface{}
	}{
		{name: "Should Return Nil When Given Empty Payload", args: args{payload: "", action: model.Create, resource: model.Monitor}, want: nil},
		{name: "Should Return Nil When Given File Not Found", args: args{payload: "tester-123.json", action: model.Create, resource: model.Monitor}, want: nil},
		{name: "Should Return Nil When Action Is Not Supported", args: args{payload: "{\"friendly_name\":\"test\"}", action: model.Create, resource: model.AlertContact}, want: nil},
		{name: "Should Return Result Payload When Given Valid JSON Payload", args: args{payload: "{\"friendly_name\":\"test\"}", action: "create", resource: "monitor"}, want: []map[string]interface{}{{
			model.ErrorResultField:       fmt.Errorf(monitor.MsgFieldMissing, httputil.TypeField),
			model.MonitorNameResultField: "test",
		}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HandleRequest(tt.args.payload, tt.args.resource, tt.args.action); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("HandleRequest() = %v, want %v", got, tt.want)
			}
		})
	}
}
