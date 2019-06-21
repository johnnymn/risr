package v1alpha1

type AutoScalingGroupNotFoundError struct {
	err string
}

func (e *AutoScalingGroupNotFoundError) Error() string {
	return e.err
}
