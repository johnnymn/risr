.PHONY: mocks

default: test

release:
	goreleaser

build:
	goreleaser --snapshot --rm-dist

GOFMT_FILES?=$$(find . -name '*.go' | grep -v vendor)
fmt:
	gofmt -w $(GOFMT_FILES)

fmtcheck:
	@sh -c "'$(CURDIR)/scripts/gofmtcheck'"

lint:
	golangci-lint run ./...

test: fmtcheck
	go test ./...

mocks:
	mockgen -destination=pkg/mocks/v1alpha1/autoscaling.go -package=v1alpha1 github.com/aws/aws-sdk-go/service/autoscaling/autoscalingiface AutoScalingAPI
	mockgen -destination=pkg/mocks/v1alpha1/elbv2.go -package=v1alpha1 github.com/aws/aws-sdk-go/service/elbv2/elbv2iface ELBV2API
