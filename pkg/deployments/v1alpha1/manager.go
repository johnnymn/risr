package v1alpha1

import (
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/autoscaling/autoscalingiface"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/aws/aws-sdk-go/service/elbv2/elbv2iface"
	"github.com/google/uuid"
	"github.com/johnnymn/risr/pkg/stacks/v1alpha1"
	"github.com/johnnymn/risr/pkg/util"
	"github.com/rs/zerolog/log"
)

const risrStackName = "v1alpha1.risr.stack"

type ManagerInterface interface {
	DeployStack(stack *v1alpha1.Stack) error
}

// Manager takes care of translating a Stack
// instance into AWS resources, and exposes an
// API to apply the necessary operations to reach
// the desired state for a Stack.
type Manager struct {
	autoscaling autoscalingiface.AutoScalingAPI
	elbv2       elbv2iface.ELBV2API
}

// MakeManager Returns a new instance of Manager.
func NewManager() (*Manager, error) {
	session, err := session.NewSession()
	if err != nil {
		return nil, err
	}

	autoscalingClient := autoscaling.New(session)
	elbv2Client := elbv2.New(session)

	manager := MakeManager(autoscalingClient, elbv2Client)
	return manager, nil
}

// MakeManager Returns a new instance of Manager
// created using the supplied AutoScaling and ELBV2
// clients.
func MakeManager(auto autoscalingiface.AutoScalingAPI, elbv2 elbv2iface.ELBV2API) *Manager {
	// This function is useful for testing because we
	// can mock and inject the clients.
	return &Manager{
		autoscaling: auto,
		elbv2:       elbv2,
	}
}

// DeployStack Deploys a given Stack using a
// blue/green deployment strategy. This means
// we first create a new group of servers, confirm
// they are healthy, and then drop the old group.
func (manager *Manager) DeployStack(stack *v1alpha1.Stack) error {
	createLCInput := generateCreateLCInput(stack)
	log.Info().Msg("Creating Launch Configuration: " + *createLCInput.LaunchConfigurationName)
	_, err := manager.autoscaling.CreateLaunchConfiguration(createLCInput)
	if err != nil {
		return err
	}

	createASGInput := generateCreateASGroupInput(stack, *createLCInput.LaunchConfigurationName)
	log.Info().Msg("Creating AutoScaling Group: " + *createASGInput.AutoScalingGroupName)
	_, err = manager.autoscaling.CreateAutoScalingGroup(createASGInput)
	if err != nil {
		return err
	}

	// Waiting for all the instances in the new
	// ASG to be healthy.
	log.Info().Msg("Waiting for ASG to be healthy")
	attempts := 0

	// Note: this should be able to handle a
	// timeout.
	for {
		attempts++
		log.Info().Msg("Attempts: " + strconv.Itoa(attempts))

		isASGHealthy, err := manager.isASGHealthy(stack, *createASGInput.AutoScalingGroupName)
		if err != nil {
			return err
		}

		if isASGHealthy {
			log.Info().Msg("ASG healthy, dropping old groups")
			err := manager.dropOldASGs(stack.Name, *createASGInput.AutoScalingGroupName)
			if err != nil {
				return err
			}

			break

		} else {
			// Note: we add have a rollback flow.
			log.Info().Msg("ASG not healthy, sleeping for 10 seconds")
			// Sleep for 10 seconds then retry
			time.Sleep(10 * time.Second)
		}
	}

	return nil
}

// isASGHealthy Returns true if all the instances
// of the ASG are healthy.
func (manager *Manager) isASGHealthy(stack *v1alpha1.Stack, asgName string) (healthy bool, err error) {
	// If the ASG is attached to a target group,
	// then check the health using the Target Groups
	// Health API.
	if stack.TargetGroupARN != nil {
		log.Info().Msg("Checking health of ASG using Target Groups API")

		healthy, err = manager.getTargetsHealth(asgName, *stack.TargetGroupARN)
		if err != nil {
			return false, err
		}
	} else {
		// If not, use the EC2 instance checks.
		log.Info().Msg("Checking health of ASG using EC2 instance checks")

		healthy, err = manager.getASGHealth(asgName)
		if err != nil {
			return false, err
		}
	}

	return
}

// getTargetsHealth Returns true if all the instances
// of the ASG are registered on the configured Target
// Group, and their status is healthy.
func (manager *Manager) getTargetsHealth(asgName string, targetGroupARN string) (bool, error) {
	describeASGInput := autoscaling.DescribeAutoScalingGroupsInput{
		AutoScalingGroupNames: aws.StringSlice([]string{asgName}),
	}

	describeASGOutput, err := manager.autoscaling.DescribeAutoScalingGroups(&describeASGInput)
	if err != nil {
		return false, err
	}

	if len(describeASGOutput.AutoScalingGroups) == 0 {
		return false, &AutoScalingGroupNotFoundError{"AutoScaling group " + asgName + " not found"}
	}

	asg := describeASGOutput.AutoScalingGroups[0]

	describeTargetHealthInput := elbv2.DescribeTargetHealthInput{
		TargetGroupArn: aws.String(targetGroupARN),
	}

	// This are the details about the health of
	// the instances registered on the Target Group.
	targetsHealth, err := manager.elbv2.DescribeTargetHealth(&describeTargetHealthInput)
	if err != nil {
		return false, err
	}

	// Checking the health of each instance that is
	// attached to the Target Group, and that belongs
	// to the new ASG.
	var numberOfInstances int64 = 0
	for _, thDescription := range targetsHealth.TargetHealthDescriptions {

		// We need to check if the Target is a member
		// of the new ASG.
		if thDescription.Target.Id != nil &&
			util.Contains(*thDescription.Target.Id, getInstanceIDsValues(asg.Instances)) {

			// If is not healthy, return false.
			if thDescription.TargetHealth != nil && *thDescription.TargetHealth.State != "healthy" {
				return false, nil
			}

			numberOfInstances += 1
		}

	}

	if numberOfInstances < *asg.DesiredCapacity {
		log.Info().Msg("Target Group health check: not at full capacity")
		return false, nil
	}

	return true, nil
}

// isASGHealthy Returns true if all the instances
// of the ASG are healthy.
func (manager *Manager) getASGHealth(asgName string) (bool, error) {
	describeASGInput := autoscaling.DescribeAutoScalingGroupsInput{
		AutoScalingGroupNames: aws.StringSlice([]string{asgName}),
	}

	describeASGOutput, err := manager.autoscaling.DescribeAutoScalingGroups(&describeASGInput)
	if err != nil {
		return false, err
	}

	if len(describeASGOutput.AutoScalingGroups) == 0 {
		return false, &AutoScalingGroupNotFoundError{"AutoScaling group " + asgName + " not found"}
	}

	asg := describeASGOutput.AutoScalingGroups[0]

	// Checking the EC2 health status for each
	// instance that belongs to the new ASG.
	var numberOfInstances int64 = 0
	for _, instance := range asg.Instances {
		if *instance.HealthStatus != "Healthy" || *instance.LifecycleState != "InService" {
			return false, nil
		}

		numberOfInstances += 1
	}

	if numberOfInstances < *asg.DesiredCapacity {
		log.Info().Msg("EC2 health check: not at full capacity")
		return false, nil
	}

	return true, nil
}

// dropOldASGs Drops old the ASG's that are tagged with
// stackName, except the ones that match the newASG name.
func (manager *Manager) dropOldASGs(stackName string, newASG string) error {
	describeASGInput := autoscaling.DescribeAutoScalingGroupsInput{}
	var forDeletion []*string

	pageNum := 0

	// Iterating through the pages of the list of ASG's
	err := manager.autoscaling.DescribeAutoScalingGroupsPages(&describeASGInput,
		func(page *autoscaling.DescribeAutoScalingGroupsOutput, lastPage bool) bool {
			pageNum++

			for _, asg := range page.AutoScalingGroups {
				for _, tag := range asg.Tags {
					// If the ASG is tagged with stackName but
					// it's name doesn't match newASG add it to the
					// forDeletion listbecause is an old deployment.
					if *tag.Key == risrStackName && *tag.Value == stackName && *asg.AutoScalingGroupName != newASG {
						forDeletion = append(forDeletion, asg.AutoScalingGroupName)
					}
				}
			}

			return true
		})
	if err != nil {
		return err
	}

	// We delete in this loop to avoid
	// complicated error handling inside the
	// DescribeAutoScalingGroupsPages closure.
	for _, del := range forDeletion {
		log.Info().Msg("Deleting old ASG: " + *del)

		_, err = manager.autoscaling.DeleteAutoScalingGroup(&autoscaling.DeleteAutoScalingGroupInput{
			AutoScalingGroupName: del,
			ForceDelete:          aws.Bool(true),
		})
		if err != nil {
			return err
		}
	}

	return nil
}

// generateCreateLCInput Creates the input object
// needed to register a new EC2 Launch Configuration
// based on a Stack difinition.
func generateCreateLCInput(stack *v1alpha1.Stack) *autoscaling.CreateLaunchConfigurationInput {
	return &autoscaling.CreateLaunchConfigurationInput{
		LaunchConfigurationName: aws.String(stack.Name + "-" + uuid.New().String()[:6]),
		IamInstanceProfile:      stack.IamInstanceProfile,
		ImageId:                 aws.String(stack.AMI),
		InstanceType:            aws.String(stack.InstanceType),
		KeyName:                 stack.KeyName,
		SecurityGroups:          stack.SecurityGroupIDs,
	}
}

// generateCreateASGroupInput Creates the input object
// needed to register a new EC2 Autoscaling Group based on
// a Stack definition.
func generateCreateASGroupInput(stack *v1alpha1.Stack,
	lauchConfigurationName string) *autoscaling.CreateAutoScalingGroupInput {

	// Check if a healthcheck grace period is specified on
	// the Stack, if not use 60 seconds as default.
	var healthCheckGracePeriod *int64 = aws.Int64(60)

	if stack.HealthCheckGracePeriod != nil {
		healthCheckGracePeriod = stack.HealthCheckGracePeriod
	}

	return &autoscaling.CreateAutoScalingGroupInput{
		AutoScalingGroupName:    aws.String(stack.Name + "-" + uuid.New().String()[:6]),
		DesiredCapacity:         aws.Int64(stack.Replicas),
		HealthCheckGracePeriod:  healthCheckGracePeriod,
		HealthCheckType:         aws.String("ELB"),
		LaunchConfigurationName: aws.String(lauchConfigurationName),
		MaxSize:                 aws.Int64(stack.Replicas),
		MinSize:                 aws.Int64(stack.Replicas),
		Tags:                    generateASGTags(stack),
		TargetGroupARNs:         []*string{stack.TargetGroupARN},
		VPCZoneIdentifier:       aws.String(strings.Join(stack.SubnetIDs, ",")),
	}
}

// generateASGTags Creates an slice of Autoscaling Tags
// for an ASG based on a Stack definition. Also inserts
// the control tags that risr expects to find on the ASG.
func generateASGTags(stack *v1alpha1.Stack) (tags []*autoscaling.Tag) {
	for key, value := range stack.Tags {
		tags = append(tags, &autoscaling.Tag{
			Key:               aws.String(key),
			Value:             aws.String(value),
			PropagateAtLaunch: aws.Bool(true),
		})
	}

	tags = append(tags, &autoscaling.Tag{
		Key:               aws.String(risrStackName),
		Value:             aws.String(stack.Name),
		PropagateAtLaunch: aws.Bool(true),
	})

	return
}

// getInstanceIDsValues Returns the string values
// of the IDs of an slice of autoscaling instances.
func getInstanceIDsValues(instances []*autoscaling.Instance) (ids []string) {
	for _, instance := range instances {
		ids = append(ids, *instance.InstanceId)
	}

	return
}
