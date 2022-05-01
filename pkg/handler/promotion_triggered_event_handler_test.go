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
			wantOwner: "markuslackner",
			wantRepository: "keptn-argo-dev",
			wantErr: false,
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
