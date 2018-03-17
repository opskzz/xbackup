package g

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

// xtrabackupAll all backup
func xtrabackupAll(user, pass, host, backupDir string, port int) error {
	var toLsn, completFlag = 0, false
	args := []string{"--compress", "--no-lock", "--stream=xbstream",
		fmt.Sprintf("--user=%s", user),
		fmt.Sprintf("--password=%s", pass),
		fmt.Sprintf("--host=%s", host),
		fmt.Sprintf("--port=%d", port),
		fmt.Sprintf("%s", backupDir),
		fmt.Sprintf("--throttle=%d", XtrabackupThrottle),
		fmt.Sprintf("--compress-threads=%d", XtrabackupThreads),
	}
	cmd := exec.Command("innobackupex", args...)
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("xtrabackup error: %v", err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("xtrabackup finished error: %v", err)
	}

	localIP := GetLocalIPAddr()
	localTimestamp := TimeToYmdHms()
	xtrabackupFullFilename := localIP + "_" + localTimestamp + ".xbstream"
	outputFile, err := os.Create(xtrabackupFullFilename)
	defer outputFile.Close()

	if err != nil {
		return fmt.Errorf("create xtrabackup file error: %v", err)
	}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("xtrabackup finished error: %v", err)
	}
	go io.Copy(outputFile, stdout)
	scanner := bufio.NewScanner(stderr)
	for scanner.Scan() {
		text := scanner.Text()
		if strings.Contains(text, "innobackupex: completed OK!") {
			completFlag = true
		} else if strings.Contains(text, "The latest check point (for incremental)") {
			if _, err := fmt.Sscanf(text, "xtrabackup: The latest check point (for incremental): '%d'\n", &toLsn); err != nil {
				return fmt.Errorf("Innobackupex finished error: %v", err)
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("xtrabackup finished error: %v", err)
	}

	if completFlag == false {
		return fmt.Errorf("xtrabackup finished error: %v", "Unknow error")
	}
	xtrabackupCheckpoints := fmt.Sprintf("to_lsn = %d", toLsn)
	xtrabackupCheckpointsFilePath := fmt.Sprintf("%s/xtrabackup_checkpoints", backupDir)
	_, err = WriteString(xtrabackupCheckpointsFilePath, xtrabackupCheckpoints)
	if err != nil {
		return fmt.Errorf("Innobackupex finished but write xtrabackupCheckpoints error: %v", err)
	}
	return nil
}

// xtrabackupInc incr backup
func xtrabackupInc(user, pass, host, backupDir, checkpoints string, port int) error {
	var toLsn, completFlag = 0, false
	point, err := ReadToString(checkpoints)
	if err != nil {
		return fmt.Errorf("Xtrabackup read checkpoints error: %v", err)
	}
	fromLsn := strings.Split(point, " ")[2]
	args := []string{"--compress", "--incremental", "--stream=xbstream",
		fmt.Sprintf("--user=%s", user),
		fmt.Sprintf("--password=%s", pass),
		fmt.Sprintf("--host=%s", host),
		fmt.Sprintf("--port=%d", port),
		fmt.Sprintf("%s", backupDir),
		fmt.Sprintf("--incremental-lsn=%s", fromLsn),
		fmt.Sprintf("--parallel=%d", 48),
		fmt.Sprintf("--throttle=%d", XtrabackupThrottle),
		fmt.Sprintf("--compress-threads=%d", XtrabackupThreads),
	}
	cmd := exec.Command("innobackupex", args...)
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("xtrabackup incremental error: %v", err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("xtrabackup incremental finished error: %v", err)
	}

	localIP := GetLocalIPAddr()
	localTimestamp := TimeToYmdHms()
	xtrabackupIncFilename := localIP + "_" + localTimestamp + ".increase.xbstream"
	outputFile, err := os.Create(xtrabackupIncFilename)
	defer outputFile.Close()

	if err != nil {
		return fmt.Errorf("create xtrabackup incremental file error: %v", err)
	}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("xtrabackup incremental finished error: %v", err)
	}
	go io.Copy(outputFile, stdout)
	scanner := bufio.NewScanner(stderr)
	for scanner.Scan() {
		text := scanner.Text()
		if strings.Contains(text, "innobackupex: completed OK!") {
			completFlag = true
		} else if strings.Contains(text, "The latest check point (for incremental)") {
			if _, err := fmt.Sscanf(text, "xtrabackup: The latest check point (for incremental): '%d'\n", &toLsn); err != nil {
				return fmt.Errorf("Innobackupex finished error: %v", err)
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("xtrabackup incremental finished error: %v", err)
	}
	if completFlag == false {
		return fmt.Errorf("xtrabackup incremental finished error: %v", "Unknow error")
	}
	xtrabackupCheckpoints := fmt.Sprintf("to_lsn = %d", toLsn)
	xtrabackupCheckpointsFilePath := fmt.Sprintf("%s/xtrabackup_checkpoints", backupDir)
	_, err = WriteString(xtrabackupCheckpointsFilePath, xtrabackupCheckpoints)
	if err != nil {
		return fmt.Errorf("Innobackupex incremental finished but write xtrabackupCheckpoints error: %v", err)
	}
	return nil
}
