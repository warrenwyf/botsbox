package runtime

import (
	"fmt"
)

const (
	VersionMajor int = 0
	VersionMinor int = 0
	VersionPatch int = 1
)

func GetVersion() string {
	return fmt.Sprintf("%d.%d.%d", VersionMajor, VersionMinor, VersionPatch)
}
