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
	"bytes"
	"fmt"
	"strings"

	etherutils "github.com/orinocopay/go-etherutils"
	"github.com/orinocopay/go-etherutils/cli"
	"github.com/orinocopay/go-etherutils/ens"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var transferPassphrase string
var transferAddressStr string
var transferGasPriceStr string

// transferCmd represents the transfer set command
var transferCmd = &cobra.Command{
	Use:   "transfer",
	Short: "Transfer an ENS name",
	Long: `Transfer an Ethereum Name Service (ENS) name's ownership to another address.  For example:

    ens transfer --address=0x5FfC014343cd971B7eb70732021E26C35B744cc4 --passphrase="my secret passphrase" enstest.eth

The keystore for the address must be local (i.e. listed with 'get accounts list') and unlockable with the supplied passphrase.

In quiet mode this will return 0 if the transaction to transfer the name is sent successfully, otherwise 1.`,
	Run: func(cmd *cobra.Command, args []string) {
		cli.Assert(transferAddressStr != "", quiet, "Address to which to transfer ownership of the name is required")
		cli.Assert(len(args[0]) > 10, quiet, "Name must be at least 7 characters long")
		cli.Assert(len(strings.Split(args[0], ".")) == 2, quiet, "Name must not contain . (except for ending in .eth)")

		// Ensure that the name is in a suitable state
		registrarContract, err := ens.RegistrarContract(client)
		inState, err := ens.NameInState(registrarContract, client, args[0], "Owned")
		cli.ErrAssert(inState, err, quiet, "Name not in a suitable state to transfer")

		// Obtain the registry contract
		registryContract, err := ens.RegistryContract(client)

		// Fetch the owner of the name
		cli.ErrCheck(err, quiet, "Invalid name")
		owner, err := registryContract.Owner(nil, ens.NameHash(args[0]))
		cli.ErrCheck(err, quiet, "Cannot obtain owner")
		cli.Assert(bytes.Compare(owner.Bytes(), ens.UnknownAddress.Bytes()) != 0, quiet, "Owner is not set")

		// Fetch the wallet and account for the owner
		wallet, err := cli.ObtainWallet(chainID, owner)
		cli.ErrCheck(err, quiet, "Failed to obtain a wallet for the owner")
		account, err := cli.ObtainAccount(&wallet, &owner, transferPassphrase)
		cli.ErrCheck(err, quiet, "Failed to obtain an account for the owner")

		gasPrice, err := etherutils.StringToWei(transferGasPriceStr)
		cli.ErrCheck(err, quiet, "Invalid gas price")

		// Set up our session
		session := ens.CreateRegistrarSession(chainID, &wallet, account, transferPassphrase, registrarContract, gasPrice)

		// Transfer the deed
		transferAddress, err := ens.Resolve(client, transferAddressStr)
		cli.ErrCheck(err, quiet, "Failed to obtain transfer address")
		tx, err := ens.Transfer(session, args[0], transferAddress)
		cli.ErrCheck(err, quiet, "Failed to send transaction")
		if !quiet {
			fmt.Println("Transaction ID is", tx.Hash().Hex())
		}
		log.WithFields(log.Fields{"transactionid": tx.Hash().Hex(),
			"name":      args[0],
			"networkid": chainID,
			"address":   transferAddress.Hex()}).Info("Transfer")
	},
}

func init() {
	RootCmd.AddCommand(transferCmd)

	transferCmd.Flags().StringVarP(&transferPassphrase, "passphrase", "p", "", "Passphrase for the account that owns the name")
	transferCmd.Flags().StringVarP(&transferAddressStr, "address", "a", "", "Address to which to transfer the ownership of the name")
	transferCmd.Flags().StringVarP(&transferGasPriceStr, "gasprice", "g", "4 GWei", "Gas price for the transaction")
}
