package main

import (
	_ "embed"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

//go:embed server/server.go
var serverSource []byte

//go:embed server/modfile
var modSource []byte

//go:embed server/sumfile
var sumSource []byte

func main() {
	if len(os.Args) < 4 || (os.Args[2] != "-o" && os.Args[2] != "--output") || os.Args[1] == "" || os.Args[3] == "" {
		programName := os.Args[0]
		fmt.Println("Usage:")
		fmt.Printf("  %s <input_folder> --output <output_binary>\n", programName)
		fmt.Printf("  %s <input_folder> -o <output_binary>\n", programName)
		os.Exit(1)
	}

	buildBinary(os.Args[1], os.Args[3])
}

func buildBinary(inputFolderPath string, outputBinaryName string) error {
	tempDir, err := os.MkdirTemp(os.TempDir(), "build")
	if err != nil {
		return fmt.Errorf("failed to create temporary directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	publicDir := filepath.Join(tempDir, "public")
	if err := copyDir(inputFolderPath, publicDir); err != nil {
		return fmt.Errorf("failed to copy input folder: %w", err)
	}

	serverSourceDestination := filepath.Join(tempDir, "server.go")
	if err := os.WriteFile(serverSourceDestination, serverSource, 0664); err != nil {
		return fmt.Errorf("failed to copy server source: %w", err)
	}

	serverModDestination := filepath.Join(tempDir, "go.mod")
	if err := os.WriteFile(serverModDestination, modSource, 0664); err != nil {
		return fmt.Errorf("failed to copy server source: %w", err)
	}

	serverSumDestination := filepath.Join(tempDir, "go.sum")
	if err := os.WriteFile(serverSumDestination, sumSource, 0664); err != nil {
		return fmt.Errorf("failed to copy server source: %w", err)
	}

	cmd := exec.Command("go", "build", "-o", "binary", ".")
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to build binary: %w", err)
	}

	binarySource := filepath.Join(tempDir, "binary")
	if err := copyFile(binarySource, outputBinaryName); err != nil {
		return fmt.Errorf("failed to copy binary to output: %w", err)
	}

	return nil
}

func copyDir(src string, dst string) error {
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dst, 0755); err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}
	return nil
}

func copyFile(sourcePath string, destinationPath string) error {
	sourceFile, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	sourceFileInfo, err := sourceFile.Stat()
	if err != nil {
		return err
	}
	sourceFilePermissions := sourceFileInfo.Mode()

	destinationFile, err := os.Create(destinationPath)
	if err != nil {
		return err
	}
	defer destinationFile.Close()

	if _, err = io.Copy(destinationFile, sourceFile); err != nil {
		return err
	}

	if err = os.Chmod(destinationPath, sourceFilePermissions); err != nil {
		return err
	}

	return destinationFile.Close()
}
