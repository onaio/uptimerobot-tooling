package fileutil

import "testing"

func TestProcessInput(t *testing.T) {
	type args struct {
		data string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{name: "should return a valid Json Array", args: args{data: "../testdata/test.json"}, want: "[{\"friendly_name\":\"test\"}]", wantErr: false},
		{name: "should return an error File Not Found", args: args{data: "../testdata/testJsonObject2-1234.json"}, want: "", wantErr: true},
		{name: "should return empty Json Array when given empty input", args: args{data: ""}, want: "", wantErr: true},
		{name: "should return an invalid Json Object when given an invalid input", args: args{data: "{\"friendly_name\":\"test\""}, want: "[{\"friendly_name\":\"test\"]", wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := TransformInputToString(tt.args.data)
			if got != tt.want {
				t.Errorf("TransformInputToString() = %v, want %v", got, tt.want)
			}
			if tt.wantErr && err == nil {
				t.Errorf("TransformInputToString() = %v, want %v", got, tt.wantErr)
			}

			if !tt.wantErr && err != nil {
				t.Errorf("TransformInputToString() = %v, want %v", got, tt.wantErr)
			}
		})
	}
}

func Test_convertFileToString(t *testing.T) {
	type args struct {
		filename string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{name: "should fail and return an error", args: args{filename: ""}, want: "", wantErr: true},
		{name: "should return a valid file", args: args{filename: "../testdata/test.json"}, want: "{\"friendly_name\":\"test\"}", wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := convertFileToString(tt.args.filename)
			if (err != nil) != tt.wantErr {
				t.Errorf("convertFileToString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("convertFileToString() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_convertStringToJsonArray(t *testing.T) {
	type args struct {
		payload string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "should return empty Json Array", args: args{payload: ""}, want: "[]"},
		{name: "should return non empty Array", args: args{payload: "{}"}, want: "[{}]"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := convertStringToJsonArray(tt.args.payload); got != tt.want {
				t.Errorf("convertStringToJsonArray() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_isFile(t *testing.T) {
	type args struct {
		payload string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{name: "should return false when given empty string", args: args{payload: ""}, want: false},
		{name: "should return false non file string", args: args{payload: "{}"}, want: false},
		{name: "should return true when valid file is supplied", args: args{payload: "testJsonArray.json"}, want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isFile(tt.args.payload); got != tt.want {
				t.Errorf("isFile() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_isStringJsonArray(t *testing.T) {
	type args struct {
		payload string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{name: "should return false when given An empty string", args: args{payload: ""}, want: false},
		{name: "should return false when given a non Json Array", args: args{payload: "{}"}, want: false},
		{name: "should return true when given a Json Array", args: args{payload: "[]]"}, want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isStringJsonArray(tt.args.payload); got != tt.want {
				t.Errorf("isStringJsonArray() = %v, want %v", got, tt.want)
			}
		})
	}
}
