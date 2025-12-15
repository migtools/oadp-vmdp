package cli

type commandRepository struct {
	connect          commandRepositoryConnect
	create           commandRepositoryCreate
	disconnect       commandRepositoryDisconnect
	repair           commandRepositoryRepair
	setClient        commandRepositorySetClient
	setParameters    commandRepositorySetParameters
	changePassword   commandRepositoryChangePassword
	status           commandRepositoryStatus
	syncTo           commandRepositorySyncTo
	throttle         commandRepositoryThrottle
	validateProvider commandRepositoryValidateProvider
	upgrade          commandRepositoryUpgrade
}

func (c *commandRepository) setup(svc advancedAppServices, parent commandParent) {
	// OADP: Renamed from "repository" to "bsl" (Backup Storage Location)
	cmd := parent.Command("bsl", "Commands to manage Backup Storage Location (BSL).")

	// OADP: Only include subcommands needed for VM users
	c.connect.setup(svc, cmd)
	c.create.setup(svc, cmd)
	c.disconnect.setup(svc, cmd)
	c.status.setup(svc, cmd)
	c.changePassword.setup(svc, cmd)
}
