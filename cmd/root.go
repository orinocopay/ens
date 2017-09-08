// Copyright © 2017 Orinoco Payments
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"context"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/orinocopay/go-etherutils/cli"
	"github.com/orinocopay/go-etherutils/ens"
	"github.com/orinocopay/go-etherutils/ens/registrarcontract"
	"github.com/orinocopay/go-etherutils/ens/registrycontract"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string
var logFile string
var quiet bool
var connection string

var client *ethclient.Client
var chainID *big.Int

// Common command-line arguments
var passphrase string
var gasPriceStr string
var nonce int64

// Common contracts
var registryContract *registrycontract.RegistryContract
var registrarContract *registrarcontract.RegistrarContract

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:              "ens",
	Short:            "manage ENS entries",
	Long:             `Manage entries for the Ethereum Name Service (ENS).  Details of each indiidual command are available in the help files for the relevant command`,
	PersistentPreRun: persistentPreRun,
}

func persistentPreRun(cmd *cobra.Command, args []string) {
	if cmd.Name() == "help" {
		// User just wants help
		return
	}

	// Ensure that the first argument is present
	if len(args) == 0 {
		cli.Err(quiet, "This command requires a name")
	}
	if args[0] == "" {
		cli.Err(quiet, "This command requires a name")
	}

	if cmd.Name() != "nonce" {
		// Add '.eth' to the end of the name if not present
		if !strings.HasSuffix(args[0], ".eth") {
			// Might be a hex address
			if len(args[0]) == 40 || len(args[0]) == 42 {
				_, err := hex.DecodeString(args[0])
				if err != nil {
					// Might be a hex address with leading 0x
					if len(args[0]) > 2 && strings.HasPrefix(args[0], "0x") {
						_, err = hex.DecodeString(args[0][2:])
					}
					if err != nil {
						// Not a valid hex string
						args[0] += ".eth"
					}
				}
			} else {
				// Not a hex string
				args[0] += ".eth"
			}
		}
	}

	// Set the log file if set, otherwise ignore
	if logFile != "" {
		f, err := os.OpenFile(logFile, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
		cli.ErrCheck(err, quiet, "Failed to open log file")
		log.SetOutput(f)
		log.SetFormatter(&log.JSONFormatter{})
	} else {
		log.SetOutput(ioutil.Discard)
	}

	// Create a connection to an Ethereum node
	var err error
	client, err = ethclient.Dial(connection)
	cli.ErrCheck(err, quiet, "Failed to connect to Ethereum")
	// Fetch the chain ID
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	chainID, err = client.NetworkID(ctx)
	cli.ErrCheck(err, quiet, "Failed to obtain chain ID")

	// Set up the common contracts
	registrarContract, err = ens.RegistrarContract(client)
	cli.ErrCheck(err, quiet, "Cannot obtain ENS registrar contract")
	registryContract, err = ens.RegistryContract(client)
	cli.ErrCheck(err, quiet, "Cannot obtain ENS registry contract")
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.cmd.yaml)")
	RootCmd.PersistentFlags().StringVarP(&logFile, "log", "l", "", "log activity to the named file")
	RootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "no output")
	RootCmd.PersistentFlags().StringVarP(&connection, "connection", "c", "https://api.orinocopay.com:8546/", "path to the Ethereum connection")
	RootCmd.PersistentFlags().Int64VarP(&nonce, "nonce", "n", -1, "Nonce for the transaction")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".cmd" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".cmd")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

//
// Helpers
//

// Add flags for commands that carry out transactions
func addTransactionFlags(cmd *cobra.Command, passphraseExplanation string) {
	cmd.Flags().StringVarP(&passphrase, "passphrase", "p", "", passphraseExplanation)
	cmd.Flags().StringVarP(&gasPriceStr, "gasprice", "g", "4 GWei", "Gas price for the transaction")
	cmd.Flags().Int64VarP(&nonce, "nonce", "n", -1, "Nonce for the transaction; -1 is auto-select")
}

func inState(name string, state string) (inState bool) {
	// Ensure that the name is in a suitable state
	inState, err := ens.NameInState(registrarContract, client, name, state)
	if err != nil {
		inState = false
	}
	return
}

func obtainWalletAndAccount(address common.Address, passphrase string) (wallet accounts.Wallet, account *accounts.Account, err error) {
	wallet, err = cli.ObtainWallet(chainID, address)
	if err == nil {
		account, err = cli.ObtainAccount(&wallet, &address, passphrase)
	}
	return wallet, account, err
}
