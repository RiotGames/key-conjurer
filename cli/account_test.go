package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAccountFuncs(t *testing.T) {
	test := &Account{
		ID:    "12345",
		Name:  "AWS - Test Account",
		Alias: "",
	}

	test.DefaultAlias()

	assert.Equal(t, test.Alias, "test", "AWS - Test Account should become `test`")

	test.SetAlias("supercooltestalias")

	assert.Equal(t, test.Alias, "supercooltestalias", "Alias should have been set")

	test.SetAlias("secondalias")

	assert.Equal(t, test.Alias, "secondalias", "Alias should have been reassigned")

	assert.Equal(t, test.IsNameMatch("Test Account"), true, "Should be able to name match with normalized name")

	assert.Equalf(t, test.IsNameMatch("secondalias"), true, "Should be able to name match %s with alias %s", "secondalias", test.Alias)

	assert.Equal(t, test.NormalizeName(), "Test Account", true, "Should match normalized name")

}
