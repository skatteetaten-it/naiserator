package util_test

import (
	"testing"

	"github.com/nais/naiserator/pkg/util"
	"github.com/stretchr/testify/assert"
)

func TestStrTrimMiddle(t *testing.T) {

	t.Run("truncate long string", func(t *testing.T) {
		s := "This is an example of a string that will get truncated. The truncate length must be sufficiently long."
		truncated := util.StrTrimMiddle(s, 40)
		assert.Equal(t, "This is an ---[truncated]---ently long.", truncated)
		assert.True(t, len(truncated) <= 40)
	})

	t.Run("truncate short string", func(t *testing.T) {
		s := "short string"
		truncated := util.StrTrimMiddle(s, 30)
		assert.Equal(t, s, truncated)
	})

	t.Run("truncate with too short length", func(t *testing.T) {
		s := "some string"
		truncated := util.StrTrimMiddle(s, 1)
		assert.Equal(t, "s", truncated)
		assert.Len(t, truncated, 1)
	})

	t.Run("truncate with too short length", func(t *testing.T) {
		s := "longer than 17, exactly 26"
		truncated := util.StrTrimMiddle(s, 18)
		assert.Equal(t, "longer than 17,...", truncated)
		assert.Len(t, truncated, 18)
	})
}

func TestSubStr(t *testing.T) {
	t.Run("zero length substring", func(t *testing.T) {
		s := "mystring"
		substr := util.SubStr(s,0,0)
		assert.Equal(t, "", substr)
	})
	t.Run("negativ length substring", func(t *testing.T) {
		s := "mystring"
		substr := util.SubStr(s,0,-1)
		assert.Equal(t, "", substr)
	})
	t.Run("start pos higher than substring length", func(t *testing.T) {
		s := "mystring"
		substr := util.SubStr(s,10,2)
		assert.Equal(t, "", substr)
	})
	t.Run("substr longer than string", func(t *testing.T) {
		s := "mystring"
		substr := util.SubStr(s,0,1000)
		assert.Equal(t, "mystring", substr)
	})
	t.Run("substring equals string", func(t *testing.T) {
		s := "x"
		substr := util.SubStr(s,0,1)
		assert.Equal(t, "x", substr)
	})
	t.Run("unicode string", func(t *testing.T) {
		s := "øøæåøø"
		substr := util.SubStr(s,2,2)
		assert.Equal(t, "æå", substr)
	})
	t.Run("substring of string", func(t *testing.T) {
		s := "mysubstring"
		substr := util.SubStr(s,2,6)
		assert.Equal(t, "substr", substr)
	})
}