package main

import (
	"fmt"
	"github.com/joho/godotenv"
	"github.com/nttu-ysc/patt/utils/file"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"
	"os"
	"syscall"
)

func init() {
	homeDir, _ := os.UserHomeDir()
	if err := godotenv.Load(homeDir + "/.patt"); err != nil {
		return
	}
}

func main() {
	rootCmd := &cobra.Command{
		Use: "patt",
		Run: func(cmd *cobra.Command, args []string) {
			p := NewDefaultPtt()
			p.Connect()
		},
	}
	rootCmd.AddCommand(&cobra.Command{
		Use:   "config",
		Short: "Start an interactive config session",
		Run: func(cmd *cobra.Command, args []string) {
			homeDir, _ := os.UserHomeDir()
			var account, password string
			var reloadSecond float64
			fmt.Print("請輸入你的帳號? ")
			fmt.Scanln(&account)
			fmt.Print("請輸入你的密碼? ")
			bytePassword, _ := terminal.ReadPassword(syscall.Stdin)
			password = string(bytePassword)
			fmt.Println()
			for {
				fmt.Print("請輸入自動更新推文間隔秒數？ (最少為 1 秒) ")
				fmt.Scanln(&reloadSecond)
				if reloadSecond >= 1 {
					break
				}
			}
			if err := file.Write(
				homeDir+"/.patt",
				[]byte(
					fmt.Sprintf(
						"account=%s\npassword=%s\nptt_reload_sec=%f",
						account,
						password,
						reloadSecond,
					),
				),
			); err != nil {
				fmt.Printf("\n設定失敗: %s\n", err)
			} else {
				fmt.Printf("\n設定完成...\n")
			}
		},
	})
	rootCmd.AddCommand(&cobra.Command{
		Use:   "help",
		Short: "Help",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	})

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
