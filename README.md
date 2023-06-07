> # Interested in Gitops with Keptn? We recommend you investigate the [Keptn Lifecycle Toolkit](https://lifecycle.keptn.sh).

# Git Promotion Service

# Artifacts

Image and helm chart are published into github packages:

#### Docker image

Github link: https://github.com/keptn-sandbox/git-promotion-service/pkgs/container/git-promotion-service

#### Helm Chart

Github link: https://github.com/keptn-sandbox/git-promotion-service/pkgs/container/git-promotion-service-chart

# Deployment

For dev cluster use for example

```
helm template \
  --namespace my-namespace \
  --set externalUrl="https://my-keptn-url" \
  oci://ghcr.io/keptn-sandbox/git-promotion-service-chart --version 0.0.1 \
  | kubectl apply -f -
```

> Add `--set pubSubUrl='nats://keptn-nats-cluster' for keptn version < 0.14

# Test

```
keptn trigger delivery --project=temp-project --service=temp-service --stage=dev --image=test --tag=1.1
```

# Configuration

## `shipyard.yaml`

#### Sequence Definition Sample

```yaml
- name: "delivery"
  triggeredOn:
    - event: "dev.evaluation.finished"
  tasks:
    - name: "git-promotion"      
```

## `git-promotion.yaml`

### Sample

```yaml

apiVersion: keptn.sh/v1
kind: GitPromotionConfig
metadata:
  name: ${project}-${service}-${stage}
spec:
  strategy: flat-pr
  target:
    repo: https://github.com/test/gke-${project}-${service}
    secret: gke-${project}
    provider: github
  paths:
    - target: ${stage}
```

#### Available Placeholders

This placeholders can be used in `promotion-config.yaml` with `${<name>}` syntax.

| Name        | Description   |
|-------------|---------------|
| `project`   | project name  |
| `service`   | service name  |
| `stage`     | current stage |
| `nextstage` | next stage    |

#### Configuration description

| Property             | Description                                                              | Sample                                            |
|----------------------|--------------------------------------------------------------------------|---------------------------------------------------|
| apiVersion           | API Version                                                              | `keptn.sh/v1`                                     |
| kind                 | Name of type                                                             | `GitPromotionConfig`                              |
| metadata.name        | Resource name                                                            | `${project}-${service}-${stage}`                  |
| spec.strategy        | Strategy to use (`branch` or `flat-pr`)                                  | `branch`                                          |
| spec.target.repo     | Target Repository                                                        | https://github.com/test/gke-${project}-${service} |
| spec.target.secret   | Secretname for token                                                     | `testsecret`                                      |
| spec.target.provider | Name of the provider                                                     | `github`                                          |
| spec.[]paths         | Paths for sync/modification. Only allowed with `spec.strategy` *flat-pr* |                                                   |
| spec.[]paths.target  | Folder to process (replace contents with placeholders)                   | `${nextstage}`                                    |
| spec.[]paths.source  | Folder to sync contents from (optional)                                  | `${stage}`                                        |

#### Strategies

##### `branch`

For every stage there **must** be a corresponding branch in the git repository. A *Pull Request* will be opened when

* there are new commits in the current stage that are not in the next stage
* there is not already an open *Pull Request* with the same source and target branches

#### `flat-pr`

A new *PullRequest* with base branch `main` and branch name `promote/<source-stage>_<target-stage>` is opened for promotion. In the 
configuration multiple paths (at least one) can be defined. 

There are two possibilities:

* `source` is empty => All files in `target` are templated
* `source` and `target` are available => All files will be synced form source to target and the target folder ist templated afterwards.


###### Placeholder replacements in files

All files in *target* path are processed and placeholders are replaced. All values of the processed *cloud event* can be accessed and used in the files.
The values must be annotated with a comment:

**Example:**
```yaml
test: default # {"keptn.git-promotion.replacewith":"shkeptncontext"}
```
In this example `default` will be replaced with the `keptncontext` in the cloud event. All data is accessible within the hierarchy through separating the names with a *dot*.

Example cloud event:

```json
{
  "data": {
    "evaluation": {
      "gitCommit": "",
      "indicatorResults": null,
      "result": "pass",
      "score": 0,
      "sloFileContent": "",
      "timeEnd": "2022-05-17T12:39:29.387Z",
      "timeStart": "2022-05-17T12:34:29.387Z",
      "timeframe": "5m"
    },
    "message": "",
    "project": "git-promotion-test-prj",
    "result": "pass",
    "service": "my-test-service",
    "stage": "dev",
    "status": "succeeded"
  },
  "gitcommitid": "27b9e0b3c8f440200b3a799cf8e54b25c2ae4502",
  "id": "9db333ce-5b16-4f4e-9036-6e61c5ec4c00",
  "shkeptncontext": "b30f2864-116a-45f5-90df-75fe05e580a3",
  "shkeptnspecversion": "latest",
  "source": "shipyard-controller",
  "specversion": "1.0",
  "time": "2022-05-17T12:39:34.907Z",
  "type": "sh.keptn.event.git-promotion.triggered"
}
```
Some sample values:

| Placeholder            | Content                                   |
|------------------------|-------------------------------------------|
| data.evaluation.result | pass                                      |
| gitcommitid            | 27b9e0b3c8f440200b3a799cf8e54b25c2ae4502  |
| data.status            | pass                                      |

####### Known Limitations

* The placeholder mechanism can only handle *string* and *int* json values at the moment. Arrays and float will most probably lead to problems
* The annotation has to be formatted **exactly** as shown in the sample. Additional spaces or missing " - although probably ok from a json/yaml point of view - will lead to problems.

###### Sample Configuration

```yaml
apiVersion: keptn.sh/v1
kind: GitPromotionConfig
metadata:
  name: ${project}-${service}-${stage}
spec:
  strategy: flat-pr
  target:
    repo: https://github.com/markuslackner/gke-${project}-${service}
    secret: gke-${project}
    provider: github
  paths:
    - source: ${stage}  
      target: ${nextstage}
```

#### Secret for github token

The secret must be available in the same namespace as the *promotion-service*. The *access-token* must be generated for a github user in
*Settings -> Developer Settings -> Personal Access Token* with `repo` Scope (TODO: must be tested if reduced scope is also working).

```yaml
apiVersion: v1
kind: Secret
metadata:  
  name: my-secret
  namespace: my-namespace
stringData:
  access-token: xxxxxxxxxxxxxxxxx
```

# Testevent

```json
{
  "data": {
    "project": "test-git-promotion",
    "service": "my-test-service",
    "stage": "dev"
  },
  "shkeptnspecversion": "0.2.4",
  "source": "curl",
  "specversion": "1.0",
  "type": "sh.keptn.event.dev.evaluation.triggered"
}
```

```bash
curl -H "accept: application/json" -H "content-type: application/json" -H "x-token: xxxxxxxxx" -d '{
"data": {
"project": "test-git-promotion",
"service": "my-test-service",
"stage": "dev"
},
"shkeptnspecversion": "0.2.4",
"source": "curl",
"specversion": "1.0",
"type": "sh.keptn.event.dev.evaluation.triggered"
}' http://localhost:8080/api/event
``` 
