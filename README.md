# Risr

## Building the binary
Download [goreleaser](https://github.com/goreleaser/goreleaser) and run:
```
make build
```
in the root folder of the repo.

If you want to contribute/test, download all the dev tools:
- [golangci-lint](https://github.com/golangci/golangci-lint) Wrapper around common golang linters.
- [mockgen](https://github.com/golang/mock) Mocking framework paired with a code generator.

and make sure to check the [Makefile](Makefile).

## Usage
Make sure you have AWS credentials on your ENV or `~/.aws/config`.

Create a `.yaml` file with the definitions for your Stack. Example:
```yaml
name: test-stack
ami: "ami-xxxxxxx"
subnetIDs:
  - subnet-xxxxxxxx
  - subnet-xxxxxxxy
  - subnet-xxxxxxxz
securityGroupIDs:
  - sg-xxxxxxxx
instanceType: "t2.micro"
replicas: 3
targetGroupARN: "arn:aws:elasticloadbalancing:us-west-2:xxxxxxxxxxxx:targetgroup/risr-test/xxxxxxxxxxxxxxxx"
tags:
  Env: "staging"
  Group: "servers"
  App: "risr"

```

and then run:
```
risr deploy example-stack.yaml
```

`risr` will take care of doing a blue/green deployment of your Stack by creating an AutoScaling group for the new version of your servers, waiting till is healthy, and then dropping the old ones.

A full reference of the possible configurations of `stack.yaml` is available [here](pkg/stacks/v1alpha1/stack.go)
