package g

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/go-redis/redis"
)

//ListScript get backup script file
func listScript(dir string) ([]string, error) {
	var filename []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		//skip dir
		if info.IsDir() {
			return nil
		}
		filename = append(filename, path)
		return nil
	})
	if err != nil {
		return filename, err
	}
	return filename, nil
}

//execute cmd and capture the output
//args maybe the shell script
func execute(script string) (*Xbackup, error) {
	xbackups := &Xbackup{}
	var stdoutBuf, stderrBuf bytes.Buffer
	var errStderr, errStdout error

	cmd := exec.Command("bash", "-c", script)
	stderrIn, _ := cmd.StderrPipe()
	stdoutIn, _ := cmd.StdoutPipe()

	stdout := io.MultiWriter(os.Stdout, &stdoutBuf)
	stderr := io.MultiWriter(os.Stderr, &stderrBuf)

	if err := cmd.Start(); err != nil {
		log.Fatalf("Cmd.Start() faild with: %x\n", err)
	}
	go func() {
		_, errStdout = io.Copy(stdout, stdoutIn)
	}()
	go func() {
		_, errStderr = io.Copy(stderr, stderrIn)
	}()
	if err := cmd.Wait(); err != nil {
		log.Fatalf("Cmd.Wait() faild with: %x\n", err)
	}
	if errStdout != nil || errStderr != nil {
		log.Fatalf("Faild to capture stdout or stderr\n")
	}
	// outStr, errStr := string(stdoutBuf.Bytes()), string(stderrBuf.Bytes())
	// result := fmt.Sprintf("Out:%s,\tErr:%s", outStr, errStr)
	// return result, nil
	err := json.Unmarshal(stdoutBuf.Bytes(), &xbackups)
	if err != nil {
		return xbackups, err
	}
	return xbackups, nil
}

//sendMsg send backup message into redis queue
func sendMsg(host, port, queue string, msg *Xbackup) error {
	client := redis.NewClient(&redis.Options{
		Addr: host + ":" + port,
	})
	_, err := client.LPush(queue, msg).Result()
	if err != nil {
		return err
	}
	return nil
}

//DoBackup execute backup script and push msg to redis
func DoBackup(qhost, qport, qqueue string) {
	scripts, err := listScript(ScriptsDir)
	if err != nil {
		log.Fatalf("Script dir: %s, backup script does't exist", ScriptsDir)
	}
	for _, script := range scripts {
		backupMsg, err := execute(script)
		if err != nil {
			log.Fatalf("Run backup script faild")
		}
		err = sendMsg(qhost, qport, qqueue, backupMsg)
		if err != nil {
			log.Fatalf("Push backup msg to queue faild")
		}
	}
}
