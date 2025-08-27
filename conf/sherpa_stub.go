//go:build !localWhisper
// +build !localWhisper

package conf

// SherpaConfigType is a stub type used when localWhisper tag is not set.
// Keep fields minimal and YAML-compatible if necessary.
type SherpaConfigType struct {
	// add placeholder fields only if you need YAML unmarshalling to populate something
}
