package backup

import (
	"backupAndPrune/internal/config"
	"backupAndPrune/pkg/utils"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// RunBackup handles the backup and pruning of log files based on the PruneAfterHours configuration.
func RunBackup(cfg *config.Config) error {
	if !cfg.EnableBackup {
		fmt.Println("Backup process is disabled in the configuration.")
		return nil
	}

	// Ensure the backup directory exists
	err := os.MkdirAll(cfg.BackupPath, os.ModePerm)
	if err != nil {
		return err
	}

	// Calculate the cutoff time based on PruneAfterHours
	pruneThreshold := time.Now().Add(-time.Duration(cfg.PruneAfterHours) * time.Hour)

	err = filepath.Walk(cfg.TargetFolder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Check if the file modification time is older than the prune threshold
		if !info.IsDir() && info.ModTime().Before(pruneThreshold) {
			destPath := filepath.Join(cfg.BackupPath, filepath.Base(path))

			// Backup the file if the backup flag is enabled
			if cfg.EnableBackup {
				err := utils.CopyFile(path, destPath)
				if err != nil {
					return err
				}
				fmt.Printf("Backed up %s to %s\n", path, destPath)

				// Optionally transfer the backup to a remote location
				if cfg.RemoteBackup != "" {
					err := utils.ExecuteCommand(fmt.Sprintf("scp %s %s", path, cfg.RemoteBackup))
					if err != nil {
						return err
					}
					fmt.Printf("Copied %s to remote backup at %s\n", path, cfg.RemoteBackup)
				}
			}

			// Prune (delete) the original file
			err = os.Remove(path)
			if err != nil {
				fmt.Printf("Error prunning file: %s\n", path)
				return err
			}
			fmt.Printf("Pruned (deleted) %s\n", path)
		}

		return nil
	})

	return err
}
