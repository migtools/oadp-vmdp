package cli

type commandSnapshot struct {
	copyHistory commandSnapshotCopyMoveHistory
	moveHistory commandSnapshotCopyMoveHistory
	create      commandSnapshotCreate
	delete      commandSnapshotDelete
	estimate    commandSnapshotEstimate
	expire      commandSnapshotExpire
	fix         commandSnapshotFix
	list        commandSnapshotList
	migrate     commandSnapshotMigrate
	pin         commandSnapshotPin
	restore     commandSnapshotRestore
	verify      commandSnapshotVerify
}

func (c *commandSnapshot) setup(svc advancedAppServices, parent commandParent) {
	// OADP: Renamed from "snapshot" to "backup"
	cmd := parent.Command("backup", "Commands to manage backups.")

	// OADP: Only include subcommands needed for VM users
	c.create.setup(svc, cmd)
	c.delete.setup(svc, cmd)
	c.list.setup(svc, cmd)
	c.restore.setup(svc, cmd)
}
