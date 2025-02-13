//go:build windows

package command

func isKeychainLockedErr(err error) bool {
	return false
}
