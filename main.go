package main

import (
	"fmt"
	"os"
	"os/exec"
	"log"
	"path/filepath"
	"time"

	"github.com/joho/godotenv"
	"strings"
)

const dtFormat string = "2006-01-02 15:04:05 Monday"

func Handle(err error) {
	if err != nil {
		log.Panic(err.Error())
	}
}

// Basic logging
func LogToFile(message string) {	
	f, err := os.OpenFile("databaseBackup.log", os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
	Handle(err)
	defer f.Close()

	log.SetOutput(f)
	log.Println(fmt.Sprintf(" | %s", message))
}

func UploadToGithub(archiveDir string) {
	// Add archive files to stage
	cmd := exec.Command("git", "add", ".")
	cmd.Dir = archiveDir
	_, err := cmd.Output()
	Handle(err)

	now := time.Now()

	// Commit archive files
	cmd = exec.Command("git", "commit", "-m", fmt.Sprintf("'%s''", now.Format(dtFormat)))
	cmd.Dir = archiveDir
	_, err = cmd.Output()
	Handle(err)

	// Push archive files to repository
	cmd = exec.Command("git", "push", "origin", "master", "--force")
	cmd.Dir = archiveDir
	_, err = cmd.Output()
	Handle(err)

	LogToFile("Successfully uploaded archive to github repository")
}

func main() {
	path, err := os.Getwd()
	Handle(err)
	
	// Load environment variables
	godotenv.Load(fmt.Sprintf("%s/.env", path))
	var mongoURI 				string = os.Getenv("mongoURI")
	var database_string string = os.Getenv("databases")
	var github 					string = os.Getenv("github")

	// Ensure required variables
	if mongoURI == "" || database_string == "" {
		log.Panic("Missing required .env variables: mongoURI, databases")
	}

	var databases []string = strings.Split(database_string, ", ")

	// Ensure archive directory exists // initialize repository if github provided
	archiveDir := filepath.Join(".", "archive")
	if _, err := os.Stat(archiveDir); os.IsNotExist(err) {
		err := os.Mkdir(archiveDir, os.ModePerm)
		Handle(err)

		if github != "" {
			cmd := exec.Command("git", "init")
			cmd.Dir = archiveDir
			_, err := cmd.Output()
			Handle(err)

			cmd = exec.Command("git", "remote", "add", "origin", github)
			cmd.Dir = archiveDir
			_, err = cmd.Output()
			Handle(err)
		}
	}

	// Perform mongodump to archive databases to .gzip format
	for _, db := range databases {
		archivePath := filepath.Join(archiveDir, db + ".gzip")
		cmd := exec.Command(
			"mongodump",
			"--uri=" + mongoURI,
			"--authenticationDatabase=admin",
			"--db=" + db,
			"--archive=" + archivePath,
			"--gzip",
		)
		_, err := cmd.Output()
		Handle(err)

		LogToFile(fmt.Sprintf("Successfully archived %s", db))
	}

	if github != "" {
		UploadToGithub(archiveDir)
	}

	LogToFile("------------------------ mongodb-backup-script ------------------------") // Log break for easier viewing
}