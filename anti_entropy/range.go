package anti_entropy

import (
	"bytes"
)

// Token range from (left, right]
type Range struct {
	Left  []byte
	Right []byte
}

func (r *Range) Equal(other Range) bool {
	return bytes.Equal(r.Left, other.Left) && bytes.Equal(r.Right, other.Right)
}
