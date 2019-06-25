package keyconjurer

import (
	"testing"

	"log"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func init() {
	logger := logrus.New()
	logger.SetOutput(os.Stderr)
	level, err := logrus.ParseLevel("debug")
	if err != nil {
		log.Fatal(err)
	}
	logger.SetLevel(level)

	Logger = logger
}

func TestAccountFuncs(t *testing.T) {
	test := &Account{
		ID:    uint(12345),
		Name:  "AWS - Test Account",
		Alias: "",
	}

	test.defaultAlias()

	assert.Equal(t, test.Alias, "test", "AWS - Test Account should become `test`")

	test.setAlias("supercooltestalias")

	assert.Equal(t, test.Alias, "supercooltestalias", "Alias should have been set")

	test.setAlias("secondalias")

	assert.Equal(t, test.Alias, "secondalias", "Alias should have been reassigned")

	assert.Equal(t, test.isNameMatch("Test Account"), true, "Should be able to name match with normalized name")

	assert.Equalf(t, test.isNameMatch("secondalias"), true, "Should be able to name match %s with alias %s", "secondalias", test.Alias)

	assert.Equal(t, test.normalizeName(), "Test Account", true, "Should match normalized name")

}
