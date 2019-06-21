package v1alpha1

import (
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/golang/mock/gomock"
	mocks "github.com/johnnymn/risr/pkg/mocks/v1alpha1"
	"github.com/johnnymn/risr/pkg/stacks/v1alpha1"
	"github.com/stretchr/testify/assert"
)

// A call to NewManager should return without
// errors. The ENV defaults are good enough
// to instantiate an AWS session object even
// if the credentials are invalid/missing.
func TestNewManager(t *testing.T) {
	_, err := NewManager()

	assert.Equal(t, nil, err)
}

// DeployStack Should fail if the response
// from autoscaling.CreateLaunchConfiguration
// is an error.
func TestDeployStackInvalidLC(t *testing.T) {
	manager, _ := NewManager()
	stack := v1alpha1.Stack{}

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	autoscalingMock := mocks.NewMockAutoScalingAPI(mockCtrl)
	autoscalingMock.EXPECT().
		CreateLaunchConfiguration(gomock.Any()).
		Return(nil, awserr.New(
			autoscaling.ErrCodeAlreadyExistsFault,
			"ErrCodeAlreadyExistsFault",
			errors.New("ErrCodeAlreadyExistsFault"))).AnyTimes()

	err := manager.DeployStack(&stack)

	assert.NotEqual(t, nil, err)
}

// DeployStack Should fail if the response
// from autoscaling.CreateAutoScalingGroup is
// an error.
func TestDeployStackInvalidASG(t *testing.T) {
	manager, _ := NewManager()
	stack := v1alpha1.Stack{}

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	autoscalingMock := mocks.NewMockAutoScalingAPI(mockCtrl)
	autoscalingMock.EXPECT().
		CreateAutoScalingGroup(gomock.Any()).
		Return(nil, awserr.New(
			autoscaling.ErrCodeAlreadyExistsFault,
			"ErrCodeAlreadyExistsFault",
			errors.New("ErrCodeAlreadyExistsFault"))).AnyTimes()

	err := manager.DeployStack(&stack)

	assert.NotEqual(t, nil, err)
}

// Note: This needs to be expanded to cover
// all cases of manager.DeployStack(stack)
