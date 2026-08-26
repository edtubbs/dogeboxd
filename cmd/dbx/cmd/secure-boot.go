package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

var secureBootCmd = &cobra.Command{
	Use:   "secure-boot",
	Short: "Run Rockchip secure-boot commands through _dbxroot",
	Run: func(cmd *cobra.Command, args []string) {
		runSecureBootCommand(args)
	},
}

var secureBootInfoCmd = &cobra.Command{
	Use:   "info",
	Short: "Read secure-boot status",
	Run: func(cmd *cobra.Command, args []string) {
		runSecureBootCommand(append([]string{"info"}, args...))
	},
}

var secureBootEnableCmd = &cobra.Command{
	Use:   "enable",
	Short: "Burn secure-boot hash and lock down the device",
	Run: func(cmd *cobra.Command, args []string) {
		runSecureBootCommand(append([]string{"enable"}, args...))
	},
}

func runSecureBootCommand(args []string) {
	cmdArgs := append([]string{"_dbxroot", "secure-boot"}, args...)
	c := exec.Command("sudo", cmdArgs...)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	c.Stdin = os.Stdin
	if err := c.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "secure-boot command failed: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(secureBootCmd)
	secureBootCmd.AddCommand(secureBootInfoCmd)
	secureBootCmd.AddCommand(secureBootEnableCmd)
}
