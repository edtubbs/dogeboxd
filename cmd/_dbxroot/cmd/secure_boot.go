package cmd

import (
	"encoding/hex"
	"fmt"
	"os"

	"github.com/Dogebox-WG/dogeboxd/pkg/secureboot"
	"github.com/spf13/cobra"
)

var secureBootCmd = &cobra.Command{
	Use:   "secure-boot",
	Short: "Rockchip secure-boot operations via OP-TEE PTA",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var secureBootInfoCmd = &cobra.Command{
	Use:   "info",
	Short: "Read secure-boot status from OP-TEE secure-boot PTA",
	Run: func(cmd *cobra.Command, args []string) {
		client, err := secureboot.NewClient()
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to initialize secure-boot client: %v\n", err)
			os.Exit(1)
		}
		defer client.Close()

		info, err := client.GetInfo()
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to read secure-boot info: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("enabled: %t\n", info.Enabled)
		fmt.Printf("simulation: %t\n", info.Simulation)
		fmt.Printf("hash: %s\n", hex.EncodeToString(info.Hash[:]))
	},
}

var secureBootEnableCmd = &cobra.Command{
	Use:   "enable",
	Short: "Burn secure-boot hash and lock down the device",
	Run: func(cmd *cobra.Command, args []string) {
		hashInput, _ := cmd.Flags().GetString("hash")
		keySize, _ := cmd.Flags().GetInt("key-size")
		force, _ := cmd.Flags().GetBool("yes")

		if !force {
			fmt.Fprintln(os.Stderr, "refusing to enable secure boot without --yes")
			os.Exit(1)
		}

		hash, err := secureboot.ParseHash(hashInput)
		if err != nil {
			fmt.Fprintf(os.Stderr, "invalid --hash: %v\n", err)
			os.Exit(1)
		}

		if err := secureboot.ValidateKeySize(keySize); err != nil {
			fmt.Fprintf(os.Stderr, "invalid --key-size: %v\n", err)
			os.Exit(1)
		}

		client, err := secureboot.NewClient()
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to initialize secure-boot client: %v\n", err)
			os.Exit(1)
		}
		defer client.Close()

		if err := client.BurnHash(hash, uint32(keySize)); err != nil {
			fmt.Fprintf(os.Stderr, "failed to burn secure-boot hash: %v\n", err)
			os.Exit(1)
		}

		if err := client.LockdownDevice(); err != nil {
			fmt.Fprintf(os.Stderr, "failed to lock down device: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("secure boot enable command completed")
	},
}

func init() {
	rootCmd.AddCommand(secureBootCmd)
	secureBootCmd.AddCommand(secureBootInfoCmd)
	secureBootCmd.AddCommand(secureBootEnableCmd)

	secureBootEnableCmd.Flags().String("hash", "", "32-byte hash as hex (64 chars)")
	secureBootEnableCmd.Flags().Int("key-size", 2048, "RSA key size in bits (2048 or 4096)")
	secureBootEnableCmd.Flags().Bool("yes", false, "Confirm irreversible fuse operations")
	secureBootEnableCmd.MarkFlagRequired("hash")

	secureBootEnableCmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		hashInput, err := cmd.Flags().GetString("hash")
		if err != nil {
			return err
		}

		_, err = secureboot.ParseHash(hashInput)
		if err != nil {
			return err
		}

		keySize, err := cmd.Flags().GetInt("key-size")
		if err != nil {
			return err
		}

		return secureboot.ValidateKeySize(keySize)
	}
}
