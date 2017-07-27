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
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	etherutils "github.com/orinocopay/go-etherutils"
	"github.com/orinocopay/go-etherutils/cli"
	"github.com/orinocopay/go-etherutils/ens"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var nameSetPassphrase string
var nameSetName string
var nameSetGasPriceStr string

// nameSetCmd represents the address set command
var nameSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set the ENS name for an address",
	Long: `Set the name registered with the Ethereum Name Service (ENS) for an address.  For example:

    ens name set --name=enstest.eth --passphrase="my secret passphrase" 0xe40626310e0726e45041ac34094037f30d2a9cc3

The keystore for the account that owns the name must be local (i.e. listed with 'get accounts list') and unlockable with the supplied passphrase.

In quiet mode this will return 0 if the transaction to set the name is sent successfully, otherwise 1.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Ensure that the name is in a suitable state
		registrarContract, err := ens.RegistrarContract(client, rpcclient)
		inState, err := ens.NameInState(registrarContract, nameSetName, "Owned")
		cli.ErrAssert(inState, err, quiet, "Name not in a suitable state to set an address")

		// Obtain the reverse registrar contract
		reverseRegistrar, err := ens.ReverseRegistrar(client, rpcclient)
		cli.ErrCheck(err, quiet, "Failed to obtain reverse registrar contract")

		nameSetAddress := common.HexToAddress(args[0])
		cli.Assert(bytes.Compare(nameSetAddress.Bytes(), ens.UnknownAddress.Bytes()) != 0, quiet, "Address is invalid")

		// Fetch the wallet and account for the owner
		wallet, err := cli.ObtainWallet(chainID, nameSetAddress)
		cli.ErrCheck(err, quiet, "Failed to obtain a wallet for the address")
		account, err := cli.ObtainAccount(wallet, nameSetAddress, nameSetPassphrase)
		cli.ErrCheck(err, quiet, "Failed to obtain an account for the address")

		gasLimit := big.NewInt(150000)
		gasPrice, err := etherutils.StringToWei(nameSetGasPriceStr)
		cli.ErrCheck(err, quiet, "Invalid gas price")

		session := ens.CreateReverseRegistrarSession(chainID, &wallet, account, nameSetPassphrase, reverseRegistrar, gasLimit, gasPrice)
		tx, err := ens.SetName(session, nameSetName)
		cli.ErrCheck(err, quiet, "Failed to set name for that address")
		if !quiet {
			fmt.Println("Transaction ID is", tx.Hash().Hex())
		}
		log.WithFields(log.Fields{"transactionid": tx.Hash().Hex(),
			"networkid": chainID,
			"name":      nameSetName,
			"address":   nameSetAddress.Hex()}).Info("Name set")
	},
}

func init() {
	nameCmd.AddCommand(nameSetCmd)

	nameSetCmd.Flags().StringVarP(&nameSetPassphrase, "passphrase", "p", "", "Passphrase for the account that owns the name")
	nameSetCmd.Flags().StringVarP(&nameSetName, "name", "a", "", "Name to resolve the address to")
	nameSetCmd.Flags().StringVarP(&nameSetGasPriceStr, "gasprice", "g", "20 GWei", "Gas price for the transaction")
}
