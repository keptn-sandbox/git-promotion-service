package config

import (
	"keptn/git-promotion-service/pkg/model"
	"reflect"
	"testing"
)

func Test_Validate(t *testing.T) {
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
			validator := NewValidator()
			if gotValidationErrrors := validator.Validate(tt.args.config); !reflect.DeepEqual(gotValidationErrrors, tt.wantValidationErrrors) {
				t.Errorf("validateConfig() = %v, want %v", gotValidationErrrors, tt.wantValidationErrrors)
			}
		})
	}
}

func stradr(str string) *string {
	return &str
}
