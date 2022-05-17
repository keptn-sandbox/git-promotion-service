package replacer

import (
	v2 "github.com/cloudevents/sdk-go/v2"
	"reflect"
	"testing"
)

func TestConvertToMap(t *testing.T) {
	evt := v2.NewEvent()
	if err := evt.SetData(v2.ApplicationJSON, map[string]string{"hallo": "test", "hallo2": "test2"}); err != nil {
		t.Errorf("err: %s", err)
	}
	type args struct {
		event v2.Event
	}
	tests := []struct {
		name    string
		args    args
		wantRes map[string]string
	}{
		{
			name: "firsttry",
			args: args{
				event: evt,
			},
			wantRes: map[string]string{
				"data.hallo":  "test",
				"data.hallo2": "test2",
				"source":      "",
				"specversion": "1.0",
				"id":          "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotRes := ConvertToMap(tt.args.event); !reflect.DeepEqual(gotRes, tt.wantRes) {
				t.Errorf("ConvertToMap() = %v, want %v", gotRes, tt.wantRes)
			}
		})
	}
}

func TestConvertToMap2(t *testing.T) {
	evt := v2.NewEvent()

	if err := evt.SetData(v2.ApplicationJSON, map[string]interface{}{
		"hallo": map[string]string{
			"test":  "test",
			"hallo": "hallo",
		},
	}); err != nil {
		t.Errorf("err: %s", err)
	}
	type args struct {
		event v2.Event
	}
	tests := []struct {
		name    string
		args    args
		wantRes map[string]string
	}{
		{
			name: "firsttry",
			args: args{
				event: evt,
			},
			wantRes: map[string]string{
				"data.hallo.test":  "test",
				"data.hallo.hallo": "hallo",
				"source":           "",
				"specversion":      "1.0",
				"id":               "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotRes := ConvertToMap(tt.args.event); !reflect.DeepEqual(gotRes, tt.wantRes) {
				t.Errorf("ConvertToMap() = %v, want %v", gotRes, tt.wantRes)
			}
		})
	}
}

func TestConvertToMapWithInt(t *testing.T) {
	evt := v2.NewEvent()

	if err := evt.SetData(v2.ApplicationJSON, map[string]interface{}{
		"hallo": map[string]string{
			"test":  "test",
			"hallo": "hallo",
		},
		"test2": map[string]interface{}{
			"meininnt": 1234,
			"string":   "hallo",
		},
	}); err != nil {
		t.Errorf("err: %s", err)
	}
	type args struct {
		event v2.Event
	}
	tests := []struct {
		name    string
		args    args
		wantRes map[string]string
	}{
		{
			name: "firsttry",
			args: args{
				event: evt,
			},
			wantRes: map[string]string{
				"data.hallo.test":     "test",
				"data.hallo.hallo":    "hallo",
				"data.test2.meininnt": "1234",
				"data.test2.string":   "hallo",
				"source":              "",
				"specversion":         "1.0",
				"id":                  "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotRes := ConvertToMap(tt.args.event); !reflect.DeepEqual(gotRes, tt.wantRes) {
				t.Errorf("ConvertToMap() = %v, want %v", gotRes, tt.wantRes)
			}
		})
	}
}

func TestConvertToMapWithNilEntry(t *testing.T) {
	evt := v2.NewEvent()

	if err := evt.SetData(v2.ApplicationJSON, map[string]interface{}{
		"hallo": map[string]string{
			"test":  "test",
			"hallo": "hallo",
		},
		"test2": map[string]interface{}{
			"meininnt": 1234,
			"string":   "hallo",
			"nilentry": nil,
		},
	}); err != nil {
		t.Errorf("err: %s", err)
	}
	type args struct {
		event v2.Event
	}
	tests := []struct {
		name    string
		args    args
		wantRes map[string]string
	}{
		{
			name: "firsttry",
			args: args{
				event: evt,
			},
			wantRes: map[string]string{
				"data.hallo.test":     "test",
				"data.hallo.hallo":    "hallo",
				"data.test2.meininnt": "1234",
				"data.test2.string":   "hallo",
				"source":              "",
				"specversion":         "1.0",
				"id":                  "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotRes := ConvertToMap(tt.args.event); !reflect.DeepEqual(gotRes, tt.wantRes) {
				t.Errorf("ConvertToMap() = %v, want %v", gotRes, tt.wantRes)
			}
		})
	}
}
