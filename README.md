# Promotion Service

# Deployment

* Image build

## shipyard.yaml

#### Sequence Definition Sample

```yaml
- name: "delivery"
  triggeredOn:
    - event: "dev.evaluation.finished"
  tasks:
    - name: "promotion"
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
  name: meinsecret
  namespace: keptn
stringData:
  access-token: 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
```

# ToDos / Remark

* new strategies like `folder` or `patch` should be defined and implemented
* Test needed for scope for github token (minimal needed scope)
* Cache kubernetes client and maybe github client (implement proper backend)
* consolidate error handling and messages with other services (keptn status and keptn result) 
* Define repository and secret in a different place (service and/or project level?, event input?) and not in promotion task properties. 
* Implement unit tests
* Implement nicer pull request: body, labels and such things
* Implement update of pull request (only needed if there is some information in body that needs to be updated)
* Implement approve
* Implement bitbucket and gitlab apis
* migrate from distributor sidecar to library