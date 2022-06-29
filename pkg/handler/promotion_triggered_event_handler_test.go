package handler

import (
	"github.com/google/go-github/github"
	"github.com/keptn/go-utils/pkg/api/models"
	"keptn/git-promotion-service/pkg/model"
	"reflect"
	"testing"
)

func Test_buildBody(t *testing.T) {
	type args struct {
		keptncontext string
		projectName  string
		serviceName  string
		stage        string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "first test",
			args: args{
				keptncontext: "f229b32b-963f-4ce0-a916-284ac59ac730",
				projectName:  "temp-project",
				serviceName:  "temp-service",
				stage:        "dev",
			},
			want: `Opened by cloud-automation sequence [f229b32b-963f-4ce0-a916-284ac59ac730](/bridge/project/temp-project/sequence/f229b32b-963f-4ce0-a916-284ac59ac730/stage/dev).

Project: *temp-project* 
Service: *temp-service* 
Stage: *dev*`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := buildBody(tt.args.keptncontext, tt.args.projectName, tt.args.serviceName, tt.args.stage); got != tt.want {
				t.Errorf("buildBody() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_readAndMergeResource(t *testing.T) {
	type args struct {
		target          model.PromotionConfig
		getResourceFunc func() (resource *models.Resource, err error)
	}
	tests := []struct {
		name    string
		args    args
		wantRet model.PromotionConfig
	}{
		{
			name: "first test",
			args: args{
				target: model.PromotionConfig{},
				getResourceFunc: func() (resource *models.Resource, err error) {
					return &models.Resource{
						ResourceContent: `
spec:
  strategy: "mystrategy"
  target:
    repo: "myrepo"
    secret: "mysecret"
    provider: "github"
  paths:
    - target: /hallo
      source: /test
`,
						ResourceURI: github.String("myresourceuri"),
					}, nil
				},
			},
			wantRet: model.PromotionConfig{
				Spec: model.PromotionConfigSpec{
					Target: model.Target{
						Repo:     github.String("myrepo"),
						Secret:   github.String("mysecret"),
						Provider: github.String("github"),
					},
					Strategy: github.String("mystrategy"),
					Paths: []model.Path{
						{
							Target: github.String("/hallo"),
							Source: github.String("/test"),
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotRet := readAndMergeResource(tt.args.target, tt.args.getResourceFunc); !reflect.DeepEqual(gotRet, tt.wantRet) {
				t.Errorf("readAndMergeResource() = %+v, want %+v", gotRet, tt.wantRet)
			}
		})
	}
}

func Test_replacePlaceHolders(t *testing.T) {
	type args struct {
		placeholders map[string]string
		p            *string
	}
	tests := []struct {
		name       string
		args       args
		wantResult *string
	}{
		{
			args: args{
				placeholders: map[string]string{
					"stage":   "mystage",
					"service": "myservice",
					"project": "myproject",
				},
				p: github.String("${project}/${service}/${stage} => project: ${project} service: ${service} stage: ${stage}"),
			},
			wantResult: github.String("myproject/myservice/mystage => project: myproject service: myservice stage: mystage"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotResult := replacePlaceHolders(tt.args.placeholders, tt.args.p); !reflect.DeepEqual(gotResult, tt.wantResult) {
				t.Errorf("replacePlaceHolders() = %v, want %v", gotResult, tt.wantResult)
			}
		})
	}
}
