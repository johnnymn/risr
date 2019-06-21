package deploy

import (
	"io/ioutil"

	"github.com/johnnymn/risr/pkg/stacks/v1alpha1"
	"gopkg.in/yaml.v2"
)

// readStackFile Reads a Stack definition file
// and return it's content as an slice of bytes.
func readStackFile(filename string) (data []byte, err error) {
	data, err = ioutil.ReadFile(filename)
	if err != nil {
		return
	}

	return
}

// parseStackFromBytes Transforms a slice of bytes
// describing a Stack definition into it's object
// representation.
func parseStackFromBytes(data []byte) (stack *v1alpha1.Stack, err error) {
	err = yaml.Unmarshal(data, &stack)
	if err != nil {
		return
	}

	return
}
