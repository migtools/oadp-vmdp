package cli

// OADP-VMDP Branding Constants.
const (
	// AppName is the CLI application name.
	AppName = "oadp-vmdp"

	// AppDisplayName is a human-friendly display name for the CLI.
	AppDisplayName = "OADP VM Data Protection"

	// AppDescription is the CLI application description.
	AppDescription = "Virtual Machine Data Protection for OpenShift Virtualization"

	// AppLongDescription is the kingpin application description.
	AppLongDescription = AppDisplayName + " - " + AppDescription

	// AppAuthor is the CLI application author.
	AppAuthor = "Red Hat, Inc. <https://www.redhat.com/>"
)

// S3 Storage Constants.
const (
	// OADPPrefix is automatically prepended to all S3 storage prefixes.
	// This ensures OADP data is isolated within shared buckets.
	OADPPrefix = "oadp-vmdp/"
)

// Directory Constants.
const (
	// ConfigDirName is the directory name for configuration files.
	ConfigDirName = "oadp"

	// LogFilePrefix is the prefix for log files.
	LogFilePrefix = "oadp-"
)
