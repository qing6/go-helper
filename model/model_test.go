package model

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestNumSpan_Split(t *testing.T) {
	numSpan := NumSpan{5, 17}
	d := int64(3)
	assert.Equal(t, []int64{3, 6, 9, 12, 15}, numSpan.Split(d))

	numSpan = NumSpan{4, 14}
	d = int64(2)
	assert.Equal(t, []int64{4, 6, 8, 10, 12}, numSpan.Split(d))

	numSpan = NumSpan{3, 21}
	d = int64(4)
	assert.Equal(t, []int64{0, 4, 8, 12, 16, 20}, numSpan.Split(d))
}