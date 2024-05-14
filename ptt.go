package main

import (
	"fmt"
	"github.com/nttu-ysc/patt/utils/xterm"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
	"golang.org/x/term"
	"io"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

type ptt struct {
	sessionStdin         io.WriteCloser
	sessionCh            chan bool
	toggleReloadComments chan bool
	reloadSecond         float64
}

func NewDefaultPtt() *ptt {
	sec, err := strconv.ParseFloat(os.Getenv("ptt_reload_sec"), 32)
	if err != nil {
		sec = 1
	}
	return &ptt{
		sessionCh:            make(chan bool, 1),
		toggleReloadComments: make(chan bool, 1),
		reloadSecond:         sec,
	}
}

func (ptt *ptt) connect() {
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

	// 視窗大小改變時發送 SIGWINCH 信號
	sigwinch := make(chan os.Signal, 1)
	signal.Notify(sigwinch, syscall.SIGWINCH)
	go func() {
		for {
			<-sigwinch
			width, height, err := xterm.TerminalSize()
			if err == nil {
				session.WindowChange(height, width)
			}
		}
	}()

	ptt.sessionStdin, _ = session.StdinPipe()

	session.Stdout = os.Stdout
	session.Stderr = os.Stderr

	// 獲取本地終端的屬性
	terminalModes := ssh.TerminalModes{
		ssh.ECHO:          0,     // 禁用本地輸入的回顯
		ssh.TTY_OP_ISPEED: 14400, // 輸入速度
		ssh.TTY_OP_OSPEED: 14400, // 輸出速度
	}

	// 設置終端模式
	fileDescriptor := int(os.Stdin.Fd())
	if terminal.IsTerminal(fileDescriptor) {
		originalState, err := term.MakeRaw(fileDescriptor)
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

	ptt.autoLogin()

	go func() {
		session.Wait()
		ptt.sessionCh <- true
	}()

	ptt.toggleReloadComments = make(chan bool, 1)

	go func() {
		for {
		RELOAD:
			active := <-ptt.toggleReloadComments
			if !active {
				continue
			}
			for {
				select {
				case <-ptt.toggleReloadComments:
					goto RELOAD
				default:
					ptt.sessionStdin.Write([]byte("ql$"))
					time.Sleep(time.Duration(ptt.reloadSecond) * time.Second)
				}
			}
		}
	}()

	ptt.detectKeyboard()
}

func (ptt *ptt) detectKeyboard() {
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		fmt.Println("Error setting terminal mode:", err)
		return
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	keyCh := make(chan rune)
	go func() {
		var buf [1]byte
		for {
			_, err := os.Stdin.Read(buf[:])
			if err != nil {
				return
			}
			keyCh <- rune(buf[0])
		}
	}()

	for {
		select {
		case key := <-keyCh:
			if key == 16 {
				ptt.toggleReloadComments <- true
				break
			} else {
				ptt.toggleReloadComments <- false
			}

			ptt.sessionStdin.Write([]byte(string(key)))
		case <-ptt.sessionCh:
			return
		}
	}
}

func (ptt *ptt) autoLogin() {
	account := os.Getenv("account")
	password := os.Getenv("password")
	if account == "" || password == "" {
		return
	}
	ptt.sessionStdin.Write([]byte(account + "\r"))
	ptt.sessionStdin.Write([]byte(password + "\r"))
}
