/*
Copyright (c) 2023 Infisical Inc.
*/
package cmd

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/Infisical/infisical-merge/packages/util"
	"github.com/manifoldco/promptui"
	"github.com/posthog/posthog-go"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

type VaultBackendType struct {
	Name        string
	Description string
}

var AvailableVaults = []VaultBackendType{
	{
		Name:        "auto",
		Description: "automatically select native vault on system",
	},
	{
		Name:        "file",
		Description: "encrypted file vault",
	},
}

var vaultSetCmd = &cobra.Command{
	Example:               `infisical vault set [file|auto] [option]`,
	Use:                   "set [file|auto] [option]",
	Short:                 "Used to set the type of vault backend to store your login details securely at rest",
	Long:                  "Used to set the type of vault backend to store your login details securely at rest",
	DisableFlagsInUseLine: true,
	Args:                  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		if len(args) >= 2 {
			vaultType := args[0]
			option := args[1]

			// Todo, add more vault types / configurations
			if vaultType != util.VAULT_BACKEND_FILE_MODE {
				log.Error().Msgf("No configuration options are available for vault type [%s]\n", vaultType)
				return
			}

			switch option {
			case "passphrase":
				{

					passphrasePrompt := promptui.Prompt{
						Label: "File vault passphrase",
					}

					passphrase, err := passphrasePrompt.Run()
					if err != nil {
						log.Error().Msgf("Unable to set passphrase for file vault because of [err=%s]", err)
						return
					}

					if passphrase == "" || len(passphrase) < 8 {
						log.Error().Msgf("Passphrase must be at least 8 characters long")
						return
					}
					setFileVaultPassphrase(passphrase)
				}
			default:
				log.Error().Msgf("Unknown option [%s] for vault set command", option)
			}

			return
		}

		fmt.Printf("Warning: This command has been deprecated. Please use 'infisical vault use [file|auto]' to select which vault to use.\n")
		selectVaultTypeCmd(cmd, args)
	},
}

var vaultUseCmd = &cobra.Command{
	Example:               `infisical vault use [file|auto]`,
	Use:                   "use [file|auto]",
	Short:                 "Used to set the type of vault backend to store your login details securely at rest",
	DisableFlagsInUseLine: true,
	Args:                  cobra.MinimumNArgs(1),
	Run:                   selectVaultTypeCmd,
}

// runCmd represents the run command
var vaultCmd = &cobra.Command{
	Use:                   "vault",
	Short:                 "Used to manage where your Infisical login token is saved on your machine",
	DisableFlagsInUseLine: true,
	Args:                  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		printAvailableVaultBackends()
	},
}

func setFileVaultPassphrase(passphrase string) {
	configFile, err := util.GetConfigFile()
	if err != nil {
		log.Error().Msgf("Unable to set passphrase for file vault because of [err=%s]", err)
		return
	}

	if configFile.VaultBackendType != "file" {
		log.Error().Msgf("You are not using file vault to store your login details. You can only set passphrase for file vault")
		return
	}

	// encode with base64
	encodedPassphrase := base64.StdEncoding.EncodeToString([]byte(passphrase))
	configFile.VaultBackendPassphrase = encodedPassphrase

	err = util.WriteConfigFile(&configFile)
	if err != nil {
		log.Error().Msgf("Unable to set passphrase for file vault because of [err=%s]", err)
		return
	}

	fmt.Printf("\nSuccessfully, set passphrase for file vault. You can now store your login details securely at rest\n")
}

func printAvailableVaultBackends() {
	fmt.Printf("Vaults are used to securely store your login details locally. Available vaults:")
	for _, vaultType := range AvailableVaults {
		fmt.Printf("\n- %s (%s)", vaultType.Name, vaultType.Description)
	}

	currentVaultBackend, err := util.GetCurrentVaultBackend()
	if err != nil {
		log.Error().Msgf("printAvailableVaultBackends: unable to print the available vault backend because of error [err=%s]", err)
	}

	Telemetry.CaptureEvent("cli-command:vault", posthog.NewProperties().Set("currentVault", currentVaultBackend).Set("version", util.CLI_VERSION))

	fmt.Printf("\n\nYou are currently using [%s] vault to store your login credentials\n", string(currentVaultBackend))
}

func selectVaultTypeCmd(cmd *cobra.Command, args []string) {
	wantedVaultTypeName := args[0]
	currentVaultBackend, err := util.GetCurrentVaultBackend()
	if err != nil {
		log.Error().Msgf("Unable to set vault to [%s] because of [err=%s]", wantedVaultTypeName, err)
		return
	}

	if wantedVaultTypeName == string(currentVaultBackend) {
		log.Error().Msgf("You are already on vault backend [%s]", currentVaultBackend)
		return
	}

	if wantedVaultTypeName == util.VAULT_BACKEND_AUTO_MODE || wantedVaultTypeName == util.VAULT_BACKEND_FILE_MODE {
		configFile, err := util.GetConfigFile()
		if err != nil {
			log.Error().Msgf("Unable to set vault to [%s] because of [err=%s]", wantedVaultTypeName, err)
			return
		}

		configFile.VaultBackendType = wantedVaultTypeName // save selected vault
		configFile.LoggedInUserEmail = ""                 // reset the logged in user to prompt them to re login

		err = util.WriteConfigFile(&configFile)
		if err != nil {
			log.Error().Msgf("Unable to set vault to [%s] because an error occurred when saving the config file [err=%s]", wantedVaultTypeName, err)
			return
		}

		fmt.Printf("\nSuccessfully, switched vault backend from [%s] to [%s]. Please login in again to store your login details in the new vault with [infisical login]\n", currentVaultBackend, wantedVaultTypeName)

		Telemetry.CaptureEvent("cli-command:vault set", posthog.NewProperties().Set("currentVault", currentVaultBackend).Set("wantedVault", wantedVaultTypeName).Set("version", util.CLI_VERSION))
	} else {
		var availableVaultsNames []string
		for _, vault := range AvailableVaults {
			availableVaultsNames = append(availableVaultsNames, vault.Name)
		}
		log.Error().Msgf("The requested vault type [%s] is not available on this system. Only the following vault backends are available for you system: %s", wantedVaultTypeName, strings.Join(availableVaultsNames, ", "))
	}
}

func init() {
	vaultCmd.AddCommand(vaultSetCmd)
	vaultCmd.AddCommand(vaultUseCmd)

	rootCmd.AddCommand(vaultCmd)
}
