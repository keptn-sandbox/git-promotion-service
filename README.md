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

## shipyard.yaml

#### Sequence Definition Sample

```yaml
- name: "delivery"
  triggeredOn:
    - event: "dev.evaluation.finished"
  tasks:
    - name: "git-promotion"
      properties:
        repository: "https://github.com/markuslackner/keptn-argo-dev"
        secretname: "my-github-secret"
        strategy: "branches"
```

#### Properties

| Property | Description | Sample |
| -------- | ----------- | ------ |
| repository | Repository URL (`https`) | https://github.com/markuslackner/keptn-argo-dev |
| secretname | Secretname for github token | testsecret |
| strategy | Promotion strategy | branches |

#### Strategies

##### branches

For every stage there **must** be a corresponding branch in the git repository. A *Pull Request* will be opened when

* there are new commits in the current stage that are not in the next stage
* there is not already an open *Pull Request* with the same source and target branches

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

# ToDos / Remark

* new strategies like `folder` or `patch` should be defined and implemented
* Test needed for scope for github token (minimal needed scope)
* Cache kubernetes client and maybe github client (implement proper backend)
* consolidate error handling and messages with other services (keptn status and keptn result) 
* Define repository and secret in a different place (service and/or project level?, event input?) and not in promotion task properties. 
* Implement unit tests
* Implement approve
* Implement bitbucket and gitlab apis
* migrate from distributor sidecar to library