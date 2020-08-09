package version

import "fmt"

// Version represents a version of httpie-go
type Version struct {
	major int
	minor int
	patch int
}

func (v *Version) String() string {
	return fmt.Sprintf("%d.%d.%d", v.major, v.minor, v.patch)
}

// Current returns current version of httpie-go
func Current() *Version {
	return &Version{major: 0, minor: 7, patch: 0}
}
