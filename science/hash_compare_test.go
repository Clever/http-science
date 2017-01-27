package science

import (
	"fmt"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHashBasicValues(t *testing.T) {
	hashInt64, err := computeJSONTreeHash(int64(5))
	assert.NoError(t, err)
	assert.Equal(t, fmt.Sprintf("%x", hashInt64), "3e27b3aa6b89137cce48b3379a2a6610")

	hashFloat, err := computeJSONTreeHash(float64(math.Pi))
	assert.NoError(t, err)
	assert.Equal(t, fmt.Sprintf("%x", hashFloat), "acd5d6588d65420cb64a22e37b888aac")

	hashStr, err := computeJSONTreeHash("testing")
	assert.NoError(t, err)
	assert.Equal(t, fmt.Sprintf("%x", hashStr), "ae2b1fca515949e5d54fb22b8ed95575")

	hashTrue, err := computeJSONTreeHash(true)
	assert.NoError(t, err)
	assert.Equal(t, fmt.Sprintf("%x", hashTrue), "01000000000000000000000000000000")

	hashFalse, err := computeJSONTreeHash(false)
	assert.NoError(t, err)
	assert.Equal(t, fmt.Sprintf("%x", hashFalse), "00000000000000000000000000000000")

	_, err = computeJSONTreeHash(nil)
	assert.Error(t, err)
}
