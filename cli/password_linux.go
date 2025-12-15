package cli

import (
	"github.com/alecthomas/kingpin/v2"
)

func (c *App) setupOSSpecificKeychainFlags(svc appServices, app *kingpin.Application) {
	// OADP: Changed from KOPIA_USE_KEYRING to OADP_USE_KEYRING, updated terminology
	app.Flag("use-keyring", "Use Gnome Keyring for storing BSL password.").Default("false").Envar(svc.EnvName("OADP_USE_KEYRING")).BoolVar(&c.keyRingEnabled)
}
