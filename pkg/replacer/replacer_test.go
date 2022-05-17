package replacer

import "testing"

func TestReplace(t *testing.T) {
	type args struct {
		fileData string
		tags     map[string]string
	}
	tests := []struct {
		name       string
		args       args
		wantResult string
	}{
		{
			name: "first try",
			args: args{
				fileData: `
test: ihllo
  kl: iaeuiaeui # {"keptn.git-promotion.replacewith":"data.erster"}
iaeie:
 - uiaeuiae: oiaeue # {"keptn.git-promotion.replacewith":"data.hallo.zweiter"}
`,
				tags: map[string]string{
					"data.erster":        "replace1",
					"data.hallo.zweiter": "replace2",
				},
			},
			wantResult: `
test: ihllo
  kl: replace1 # {"keptn.git-promotion.replacewith":"data.erster"}
iaeie:
 - uiaeuiae: replace2 # {"keptn.git-promotion.replacewith":"data.hallo.zweiter"}
`,
		},
		{
			name: "error with shkeptncontext",
			args: args{
				fileData: `hallo: du # {"keptn.git-promotion.replacewith":"shkeptncontext"}`,
				tags: map[string]string{
					"shkeptncontext": "1ed398e3-641f-4de2-8663-73e303f78f96",
				},
			},
			wantResult: `hallo: 1ed398e3-641f-4de2-8663-73e303f78f96 # {"keptn.git-promotion.replacewith":"shkeptncontext"}`,
		},
		{
			name: "deployment replacer",
			args: args{
				fileData: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ .Chart.Name }}
  namespace: {{ .Values.stage }}-{{ .Chart.Name }}
  labels:
    app.kubernetes.io/component: {{ .Chart.Name }}
spec:
  replicas: {{ .Values.replicas }}
  selector:
    matchLabels:
      app.kubernetes.io/component: {{ .Chart.Name }}
  template:
    metadata:
      labels:
        app.kubernetes.io/component: {{ .Chart.Name }}
      annotations:
        shkeptncontext: unknown # {"keptn.git-promotion.replacewith":"shkeptncontext"}
`,
				tags: map[string]string{
					"data.labels.argocd_url":  "https://34.71.16.176/applications/dev-flat-pr-podtato",
					"data.labels.servicename": "dev-flat-pr-podtato",
					"data.labels.version":     "b9675facb0d4656054b0f30c2b8548b4d7abc1c3",
					"data.project":            "git-promotion-test-prj",
					"data.result":             "pass",
					"data.service":            "flat-pr-svc",
					"data.stage":              "dev",
					"data.status":             "succeeded",
					"data.temporaryData.distributor.subscriptionID": "bb00968f-fc3d-4165-8497-05f523272b0e",
					"gitcommitid":        "2ff57d30afc96f68a3f064237d3fe8431e59980e",
					"id":                 "8e04a78e-74e0-444c-bda1-d96b7ce12e59",
					"shkeptncontext":     "mykeptncontext",
					"shkeptnspecversion": "latest",
					"source":             "shipyard-controller",
					"specversion":        "1.0",
					"triggeredid":        ""},
			},
			wantResult: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ .Chart.Name }}
  namespace: {{ .Values.stage }}-{{ .Chart.Name }}
  labels:
    app.kubernetes.io/component: {{ .Chart.Name }}
spec:
  replicas: {{ .Values.replicas }}
  selector:
    matchLabels:
      app.kubernetes.io/component: {{ .Chart.Name }}
  template:
    metadata:
      labels:
        app.kubernetes.io/component: {{ .Chart.Name }}
      annotations:
        shkeptncontext: mykeptncontext # {"keptn.git-promotion.replacewith":"shkeptncontext"}
`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotResult := Replace(tt.args.fileData, tt.args.tags); gotResult != tt.wantResult {
				t.Errorf("Replace() = %v, want %v", gotResult, tt.wantResult)
			}
		})
	}
}
