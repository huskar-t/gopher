package query

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestQuery(t *testing.T) {
	cond := Where(
		And("tag1", "tag1").And("tag2", 2),
	).Or(
		And("username", "eric"),
		And("username", "jerry").And("gender", "male"),
	)

	assert.Equal(t, cond.Ands["tag1"], "tag1")
	assert.Equal(t, cond.Ands["tag2"], 2)

	assert.Equal(t, len(cond.Ands), 2)
	assert.Equal(t, len(cond.Ors), 2)
	assert.Equal(t, len(cond.Ors[1]), 2)

	assert.Equal(t, cond.Ors[0]["username"], "eric")
	assert.Equal(t, cond.Ors[1]["username"], "jerry")
	assert.Equal(t, cond.Ors[1]["gender"], "male")
}
