//go:build linux

package command

func isKeychainLockedErr(err error) bool {
	return false
}
