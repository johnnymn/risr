package v1alpha1

// Stack is a repesentation of the
// desired state of a given Stack.
type Stack struct {
	// Stack name. We will use this
	// as ID because we are not persisting
	// this to any storage, so we will assume
	// the name of the Stack is unique.
	Name string `json:"name" yaml:"name"`

	// ID of the AMI.
	AMI string `json:"ami" yaml:"ami"`

	// The ID's of the subnets where
	// we want the EC2 instances to be
	// created.
	// Note: we can figure out the VPC
	// from this.
	SubnetIDs []string `json:"subnetIDs,omitempty" yaml:"subnetIDs,omitempty"`

	// A list of all the SG's we want
	// to attach to the EC2 instances.
	SecurityGroupIDs []*string `json:"securityGroupIDs,omitempty" yaml:"securityGroupIDs,omitempty"`

	// EC2 instance type.
	InstanceType string `json:"instanceType" yaml:"instanceType"`

	// Name of the EC2 Key Pair that is going
	// to be used to SSH into the EC2 instances.
	KeyName *string `json:"keyName,omitempty" yaml:"keyName,omitempty"`

	// UserData script to be executed on boot
	// by the EC2 instances.
	UserData *string `json:"userData,omitempty" yaml:"userData,omitempty"`

	// Desired number of servers
	// in the stack.
	Replicas int64 `json:"replicas" yaml:"replicas"`

	// Name of the IAM instance profile
	// that will be applied to the EC2
	// instances.
	IamInstanceProfile *string `json:"iamInstanceProfile,omitempty" yaml:"iamInstanceProfile,omitempty"`

	// ARN of the TargetGroup that
	// will be used for the EC2 instances.
	TargetGroupARN *string `json:"targetGroupARN,omitempty" yaml:"targetGroupARN,omitempty"`

	// The amount of time, in seconds, that
	// Amazon EC2 Auto Scaling waits before
	// checking the health status of an EC2
	// instance that has come into service
	HealthCheckGracePeriod *int64 `json:"healthCheckGracePeriod,omitempty" yaml:"healthCheckGracePeriod,omitempty"`

	// Map of tags that will be applied on
	// all the resources (where possible).
	Tags map[string]string `json:"tags,omitempty" yaml:"tags,omitempty"`
}
