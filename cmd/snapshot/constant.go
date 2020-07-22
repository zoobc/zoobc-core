package snapshot

import "github.com/spf13/cobra"

var (
	snapshotHeight uint32
	dbPath, dbName string
	snapshotFile   string
	dump           bool

	snapshotCmd = &cobra.Command{
		Use:   "snapshot",
		Short: "Root command of snapshot behavior",
	}
	newSnapshotCommand = &cobra.Command{
		Use:   "new",
		Short: "Snapshot sub command for generate new snapshot file",
		Long:  "Snapshot sub command that aim to generating new snapshot file based on database target",
	}
	importSnapshotCommand = &cobra.Command{
		Use:   "import",
		Short: "Snapshot sub command simulation for import from snapshot file and storing snapshot payload into a database target",
	}
)
