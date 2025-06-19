package httpfile

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPathName(t *testing.T) {

	testPathName := `20250528三期简易版户型折页（C2-1栋及C3-1栋.pdf`
	result := validateFileDirName(testPathName)
	assert.Equal(t, true, result)

}
