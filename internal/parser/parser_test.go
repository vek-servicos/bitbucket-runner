
package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParsePipelineConfig(t *testing.T) {
	config, err := ParsePipelineConfig("../../testdata/bitbucket-pipelines.yml")

	assert.NoError(t, err)
	assert.NotNil(t, config)

	// Add more specific assertions here based on the content of the test file
}
