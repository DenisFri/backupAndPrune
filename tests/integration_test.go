package integration

import (
	"backupAndPrune/internal/backup"
	"backupAndPrune/internal/config"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestIntegrationRunBackup tests the full integration of the backup process
func TestIntegrationRunBackup(t *testing.T) {
	// Create temporary directories for logs and backups
	logDir, err := ioutil.TempDir("", "logdir")
	if err != nil {
		t.Fatalf("Failed to create temp log directory: %v", err)
	}
	defer os.RemoveAll(logDir)

	backupDir, err := ioutil.TempDir("", "backupdir")
	if err != nil {
		t.Fatalf("Failed to create temp backup directory: %v", err)
	}
	defer os.RemoveAll(backupDir)

	// Create a test log file that should be backed up and pruned
	oldFilePath := filepath.Join(logDir, "old.log")
	if err := ioutil.WriteFile(oldFilePath, []byte("old log data"), 0644); err != nil {
		t.Fatalf("Failed to create old log file: %v", err)
	}
	// Modify the file's modification time to be older than the prune threshold
	oldModTime := time.Now().Add(-48 * time.Hour)
	if err := os.Chtimes(oldFilePath, oldModTime, oldModTime); err != nil {
		t.Fatalf("Failed to set modification time on old log file: %v", err)
	}

	// Create a test log file that should not be backed up or pruned
	newFilePath := filepath.Join(logDir, "new.log")
	if err := ioutil.WriteFile(newFilePath, []byte("new log data"), 0644); err != nil {
		t.Fatalf("Failed to create new log file: %v", err)
	}

	// Define the config for the backup
	cfg := &config.Config{
		PruneAfterHours: 24, // Files older than 24 hours should be backed up and pruned
		BackupPath:      backupDir,
		EnableBackup:    true,
		TargetFolder:    logDir,
		RemoteBackup:    "", // Remote backup not tested in this integration test
	}

	// Run the backup process
	err = backup.RunBackup(cfg)
	if err != nil {
		t.Fatalf("RunBackup failed: %v", err)
	}

	// Verify that the old log file was backed up
	backupOldFilePath := filepath.Join(backupDir, "old.log")
	if _, err := os.Stat(backupOldFilePath); os.IsNotExist(err) {
		t.Errorf("Expected old log file to be backed up, but it was not found in backup directory")
	}

	// Verify that the old log file was pruned
	if _, err := os.Stat(oldFilePath); !os.IsNotExist(err) {
		t.Errorf("Expected old log file to be pruned (deleted), but it still exists")
	}

	// Verify that the new log file was not backed up or pruned
	backupNewFilePath := filepath.Join(backupDir, "new.log")
	if _, err := os.Stat(backupNewFilePath); !os.IsNotExist(err) {
		t.Errorf("Expected new log file to not be backed up, but it was found in backup directory")
	}

	if _, err := os.Stat(newFilePath); os.IsNotExist(err) {
		t.Errorf("Expected new log file to not be pruned (deleted), but it is missing")
	}
}

// TestIntegrationRunBackupNoPrune ensures that when files are newer than PruneAfterHours, they are not backed up or pruned
func TestIntegrationRunBackupNoPrune(t *testing.T) {
	// Create temporary directories for logs and backups
	logDir, err := ioutil.TempDir("", "logdir")
	if err != nil {
		t.Fatalf("Failed to create temp log directory: %v", err)
	}
	defer os.RemoveAll(logDir)

	backupDir, err := ioutil.TempDir("", "backupdir")
	if err != nil {
		t.Fatalf("Failed to create temp backup directory: %v", err)
	}
	defer os.RemoveAll(backupDir)

	// Create a test log file that should not be backed up or pruned
	newFilePath := filepath.Join(logDir, "new.log")
	if err := ioutil.WriteFile(newFilePath, []byte("new log data"), 0644); err != nil {
		t.Fatalf("Failed to create new log file: %v", err)
	}
	// Modify the file's modification time to be within the PruneAfterHours threshold
	newModTime := time.Now().Add(-12 * time.Hour)
	if err := os.Chtimes(newFilePath, newModTime, newModTime); err != nil {
		t.Fatalf("Failed to set modification time on new log file: %v", err)
	}

	// Define the config for the backup
	cfg := &config.Config{
		PruneAfterHours: 24, // Files newer than 24 hours should not be backed up or pruned
		BackupPath:      backupDir,
		EnableBackup:    true,
		TargetFolder:    logDir,
		RemoteBackup:    "", // Remote backup not tested in this integration test
	}

	// Run the backup process
	err = backup.RunBackup(cfg)
	if err != nil {
		t.Fatalf("RunBackup failed: %v", err)
	}

	// Verify that the new log file was not backed up
	backupNewFilePath := filepath.Join(backupDir, "new.log")
	if _, err := os.Stat(backupNewFilePath); !os.IsNotExist(err) {
		t.Errorf("Expected new log file to not be backed up, but it was found in backup directory")
	}

	// Verify that the new log file was not pruned
	if _, err := os.Stat(newFilePath); os.IsNotExist(err) {
		t.Errorf("Expected new log file to not be pruned (deleted), but it is missing")
	}
}
