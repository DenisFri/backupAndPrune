package utils

import (
	"io"
	"os"
	"os/exec"
)

func CopyFile(src, dest string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}

	return destFile.Sync()
}

func ExecuteCommand(command string) error {
	cmd := exec.Command("sh", "-c", command)
	return cmd.Run()
}
