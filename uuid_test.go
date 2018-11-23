package wpgx_test

import "testing"
import "github.com/shestakovda/wpgx"
import "github.com/stretchr/testify/assert"

func TestXUID(t *testing.T) {
	uuid := wpgx.UUID()
	assert.Len(t, uuid, 32)

	guid := wpgx.GUID()
	assert.Len(t, guid, 36)
}
