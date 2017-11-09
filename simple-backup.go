package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/spf13/viper"

)

func trimString(s string) string {
	re := regexp.MustCompile(`\r?\n`)
	s = strings.TrimSpace(s)
	return re.ReplaceAllString(s, " ")
}


var	remoteUser   string
var	remoteServer string
var	remotePath   string

var	localBasePath string
var	localPaths    []string
var	logFile       io.WriteCloser

var	verbosity int


func Init() {
	viper.SetConfigName("simple-backup") // name of config file (without extension)
	viper.AddConfigPath("/etc")          // path to look for the config file in
	viper.AddConfigPath("$HOME/.local")  // call multiple times to add many search paths
	viper.AddConfigPath(".")

	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {             // Handle errors reading the config file
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}

	remoteServer = viper.GetString("remote.server")
	remotePath   = viper.GetString("remote.path")
	remoteUser   = viper.GetString("remote.user")

	localBasePath = viper.GetString("local.basepath")
	localPaths    = viper.GetStringSlice("local.paths")

	logPath := viper.GetString("local.logpath")

	logFile, err = os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	if err != nil {
		panic(err)
	}
	if verbosity > 0 {
		logWriter := io.MultiWriter(os.Stdout, logFile)
		log.SetOutput(logWriter)
	} else {
		log.SetOutput(logFile)
	}
	log.Println("Backup started")
	// Assume success
}

func Close() {
	// Recover from panics
	r := recover()
	if r == nil {
		log.Println("Backup finished")
	} else {
		log.Println("Backup terminated because of errors")
		if verbosity == 0 {
			// Output something so user knows
			fmt.Fprintln(os.Stderr, "Fatal error:", r)
		}
	}
	if logFile != nil {
		logFile.Close()
	}
}

func remoteDirectoryExists(path string) bool {
	_, exitCode := remoteExec("test", "-d", path)
	return exitCode == 0
}

func getLastBackupPath() string {

	// Find all folders in remotePath but not itself
	output, exitCode := remoteExec("find", remotePath, "-mindepth", "1", "-maxdepth", "1", "-type", "d")
	// panic on exitCode !== 0
	if exitCode != 0 {
		output = trimString(output)
		log.Panicln(output)
	}
	paths := strings.Split(output, "\n")
	sort.Strings(paths)
	if len(paths) > 0 {
		return paths[len(paths)-1]
	}
	return ""
}

func remoteExec(remoteCommand string, remoteArguments ...string) (string, int) {
	arguments := []string{"-o", "PasswordAuthentication=no", remoteUser + "@" + remoteServer, remoteCommand}
	arguments = append(arguments, remoteArguments...)
	output, exitCode := RunCommand("ssh", arguments...)
	return output, exitCode
}

func remoteExecOrPanic(remoteCommand string, remoteArguments ...string) {
	output, exitCode := remoteExec(remoteCommand, remoteArguments...)
	if exitCode != 0 {
		log.Panicln(trimString(output))
	}
}

func Backup() {
	latestBackupPath := getLastBackupPath()

	// This really sucks or does it
	now := time.Now().Format("2006-01-02_150405")
	nextBackupPath := remotePath + "/" + now

	nextBackupTempPath := remotePath + "/temp"
	if latestBackupPath == "" {
		// No previous backups
		log.Println("First backup")
		remoteExecOrPanic("mkdir", "-d", nextBackupTempPath)

	} else if strings.HasSuffix(latestBackupPath, "/temp") {
		// Previous backup not finished do nothing
		log.Println("Continue previous failed backup")
	} else {
		remoteExecOrPanic("cp", "-al", latestBackupPath, nextBackupTempPath)
	}

	for _, path := range localPaths {
		backupFolder(path, nextBackupTempPath)
	}

	remoteExecOrPanic("mv", nextBackupTempPath, nextBackupPath)
	log.Println("Backup written to", nextBackupPath)

}

func backupFolder(path, nextBackupPath string) {
	sourcePath := localBasePath + "/" + path + "/"
	if fileInfo, err := os.Stat(sourcePath); err != nil || !fileInfo.IsDir() {
		log.Panicln("Not a directory", sourcePath)
	}
	destinationPath := nextBackupPath + "/" + path

	if !remoteDirectoryExists(destinationPath) {
		remoteExecOrPanic("mkdir", "-p", destinationPath)
	}

	destination := remoteUser + "@" + remoteServer + ":" + destinationPath
	output, exitCode := RunCommand("nice", "rsync", "-az", "--delete", sourcePath, destination)
	if exitCode == 0 {
		log.Println("backed up", sourcePath)
	} else {
		log.Panicln("backup of", sourcePath, "failed", trimString(output))
	}
}

func main() {
	verbose := flag.Bool("v", false, "Verbose output")
	veryVerbose := flag.Bool("vv", false, "Very verbose output")

	flag.Parse()

	verbosity = 0
	if *veryVerbose {
		verbosity = 2
	} else if *verbose {
		verbosity = 1
	}



	Init()

	defer Close()

	Backup()
}
