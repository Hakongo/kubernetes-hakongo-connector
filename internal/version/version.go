package version

// Version is the current version of the connector.
// This value is set during build using ldflags.
var Version = "dev"

// GetVersion returns the current version of the connector.
func GetVersion() string {
	return Version
}
