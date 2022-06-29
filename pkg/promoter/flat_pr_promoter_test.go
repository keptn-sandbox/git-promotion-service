package promoter

import (
	"keptn/git-promotion-service/pkg/repoaccess"
	"testing"
)

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
