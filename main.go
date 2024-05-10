package main

import (
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
	"log"
	"os"
)

func main() {
	// SSH 連接設置
	config := &ssh.ClientConfig{
		User:            "bbsu",
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	// 建立 SSH 連接
	conn, err := ssh.Dial("tcp", "ptt.cc:22", config)
	if err != nil {
		log.Fatalf("unable to connect: %s", err)
	}
	defer conn.Close()

	// 創建一個新的會話
	session, err := conn.NewSession()
	if err != nil {
		log.Fatalf("unable to create session: %s", err)
	}
	defer session.Close()

	// 連接標準輸入、輸出和錯誤
	session.Stdout = os.Stdout
	session.Stderr = os.Stderr
	session.Stdin = os.Stdin

	// 獲取本地終端的屬性
	terminalModes := ssh.TerminalModes{
		ssh.ECHO:          0,     // 禁用本地輸入的回顯
		ssh.TTY_OP_ISPEED: 14400, // 輸入速度
		ssh.TTY_OP_OSPEED: 14400, // 輸出速度
	}

	// 設置終端模式
	fileDescriptor := int(os.Stdin.Fd())
	if terminal.IsTerminal(fileDescriptor) {
		originalState, err := terminal.MakeRaw(fileDescriptor)
		if err != nil {
			panic(err)
		}
		defer terminal.Restore(fileDescriptor, originalState)

		termWidth, termHeight, err := terminal.GetSize(fileDescriptor)
		if err != nil {
			panic(err)
		}

		err = session.RequestPty("xterm-256color", termHeight, termWidth, terminalModes)
		if err != nil {
			panic(err)
		}
	}

	// 啟動一個 shell
	err = session.Shell()
	if err != nil {
		log.Fatalf("failed to start shell: %s", err)
	}

	// 等待結束
	session.Wait()
}
