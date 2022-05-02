package handler

import "testing"

func Test_getGithubOwnerRepository(t *testing.T) {
	type args struct {
		raw string
	}
	tests := []struct {
		name           string
		args           args
		wantOwner      string
		wantRepository string
		wantErr        bool
	}{
		{
			name: "testsunshine",
			args: args{
				raw: "https://github.com/markuslackner/keptn-argo-dev",
			},
			wantOwner:      "markuslackner",
			wantRepository: "keptn-argo-dev",
			wantErr:        false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotOwner, gotRepository, err := getGithubOwnerRepository(tt.args.raw)
			if (err != nil) != tt.wantErr {
				t.Errorf("getGithubOwnerRepository() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotOwner != tt.wantOwner {
				t.Errorf("getGithubOwnerRepository() gotOwner = %v, want %v", gotOwner, tt.wantOwner)
			}
			if gotRepository != tt.wantRepository {
				t.Errorf("getGithubOwnerRepository() gotRepository = %v, want %v", gotRepository, tt.wantRepository)
			}
		})
	}
}

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
