package handler

import (
	"github.com/google/go-github/github"
	"github.com/keptn/go-utils/pkg/api/models"
	"keptn/git-promotion-service/pkg/model"
	"keptn/git-promotion-service/pkg/repoaccess"
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

func Test_checkForChanges(t *testing.T) {
	type args struct {
		files  []repoaccess.RepositoryFile
		files2 []repoaccess.RepositoryFile
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "test both empty",
			args: args{
				files:  []repoaccess.RepositoryFile{},
				files2: []repoaccess.RepositoryFile{},
			},
			want: false,
		},
		{
			name: "test different size",
			args: args{
				files: []repoaccess.RepositoryFile{
					{
						Content: "hallo",
						Path:    "/mnt",
						SHA:     "sha",
					},
				},
				files2: []repoaccess.RepositoryFile{},
			},
			want: true,
		},
		{
			name: "test same content",
			args: args{
				files: []repoaccess.RepositoryFile{
					{
						Content: "hallo",
						Path:    "/mnt",
						SHA:     "sha",
					},
				},
				files2: []repoaccess.RepositoryFile{
					{
						Content: "hallo",
						Path:    "/mnt",
						SHA:     "sha",
					},
				},
			},
			want: false,
		},
		{
			name: "test different content",
			args: args{
				files: []repoaccess.RepositoryFile{
					{
						Content: "hallo",
						Path:    "/mnt",
						SHA:     "sha",
					},
				},
				files2: []repoaccess.RepositoryFile{
					{
						Content: "halleo",
						Path:    "/mnt",
						SHA:     "sha",
					},
				},
			},
			want: true,
		},
		{
			name: "test different path",
			args: args{
				files: []repoaccess.RepositoryFile{
					{
						Content: "hallo",
						Path:    "/mnt",
						SHA:     "sha",
					},
				},
				files2: []repoaccess.RepositoryFile{
					{
						Content: "hallo",
						Path:    "/mnt/test",
						SHA:     "sha",
					},
				},
			},
			want: true,
		},
		{
			name: "test different sha",
			args: args{
				files: []repoaccess.RepositoryFile{
					{
						Content: "hallo",
						Path:    "/mnt",
						SHA:     "sha",
					},
				},
				files2: []repoaccess.RepositoryFile{
					{
						Content: "hallo",
						Path:    "/mnt",
						SHA:     "shaiaeuiae",
					},
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := checkForChanges(tt.args.files, tt.args.files2); got != tt.want {
				t.Errorf("checkForChanges() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_validateConfig(t *testing.T) {
	type args struct {
		config model.PromotionConfig
	}
	tests := []struct {
		name                  string
		args                  args
		wantValidationErrrors []string
	}{
		{
			name: "valid branch config",
			args: args{
				config: model.PromotionConfig{
					APIVersion: stradr("keptn.sh/v1"),
					Kind:       stradr("GitPromotionConfig"),
					Spec: model.PromotionConfigSpec{
						Strategy: stradr("branch"),
						Target: model.Target{
							Repo:     stradr("https://github.com/test/test"),
							Secret:   stradr("hallosecret"),
							Provider: stradr("github"),
						},
					},
				},
			},
		},
		{
			name: "valid flat-pr config",
			args: args{
				config: model.PromotionConfig{
					APIVersion: stradr("keptn.sh/v1"),
					Kind:       stradr("GitPromotionConfig"),
					Spec: model.PromotionConfigSpec{
						Strategy: stradr("flat-pr"),
						Target: model.Target{
							Repo:     stradr("https://github.com/test/test"),
							Secret:   stradr("hallosecret"),
							Provider: stradr("github"),
						},
						Paths: []model.Path{
							{
								Source: stradr("hello"),
								Target: stradr("mygoodfriend"),
							},
						},
					},
				},
			},
		},
		{
			name: "flat-pr config without paths",
			args: args{
				config: model.PromotionConfig{
					APIVersion: stradr("keptn.sh/v1"),
					Kind:       stradr("GitPromotionConfig"),
					Spec: model.PromotionConfigSpec{
						Strategy: stradr("flat-pr"),
						Target: model.Target{
							Repo:     stradr("https://github.com/test/test"),
							Secret:   stradr("hallosecret"),
							Provider: stradr("github"),
						},
					},
				},
			},
			wantValidationErrrors: []string{
				"at least one path is necessary for strategy flat-pr",
			},
		},
		{
			name: "flat-pr config with contained paths",
			args: args{
				config: model.PromotionConfig{
					APIVersion: stradr("keptn.sh/v1"),
					Kind:       stradr("GitPromotionConfig"),
					Spec: model.PromotionConfigSpec{
						Strategy: stradr("flat-pr"),
						Target: model.Target{
							Repo:     stradr("https://github.com/test/test"),
							Secret:   stradr("hallosecret"),
							Provider: stradr("github"),
						},
						Paths: []model.Path{
							{
								Source: stradr("testsource"),
								Target: stradr("testtarget"),
							},
							{
								Source: stradr("testsource/hello"),
								Target: stradr("testtarget/hello"),
							},
						},
					},
				},
			},
			wantValidationErrrors: []string{
				"paths[1].target is already included in paths[0].target",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotValidationErrrors := validateConfig(tt.args.config); !reflect.DeepEqual(gotValidationErrrors, tt.wantValidationErrrors) {
				t.Errorf("validateConfig() = %v, want %v", gotValidationErrrors, tt.wantValidationErrrors)
			}
		})
	}
}

func stradr(str string) *string {
	return &str
}
