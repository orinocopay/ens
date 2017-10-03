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
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	etherutils "github.com/orinocopay/go-etherutils"
	"github.com/orinocopay/go-etherutils/cli"
	"github.com/orinocopay/go-etherutils/ens"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var nameSetName string

// nameSetCmd represents the name set command
var nameSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set the ENS name for an address",
	Long: `Set the name registered with the Ethereum Name Service (ENS) for an address.  For example:

    ens name set --name=enstest.eth --passphrase="my secret passphrase" 0xED96dD3Be847b387217EF9DE5B20D8392A6cdf40

The keystore for the account that owns the name must be local (i.e. listed with 'get accounts list') and unlockable with the supplied passphrase.

In quiet mode this will return 0 if the transaction to set the name is sent successfully, otherwise 1.`,
	Run: func(cmd *cobra.Command, args []string) {
		cli.Assert(nameSetName != "", quiet, "Name is required")

		// Obtain the reverse registrar contract
		reverseRegistrar, err := ens.ReverseRegistrarContract(client)
		cli.ErrCheck(err, quiet, "Failed to obtain reverse registrar contract")

		address := common.HexToAddress(args[0])

		// Fetch the wallet and account for the address
		wallet, account, err := obtainWalletAndAccount(address, passphrase)
		cli.ErrCheck(err, quiet, fmt.Sprintf("Failed to obtain account details for the address %s", args[0]))

		gasPrice, err := etherutils.StringToWei(gasPriceStr)
		cli.ErrCheck(err, quiet, fmt.Sprintf("Invalid gas price %s", gasPriceStr))

		session := ens.CreateReverseRegistrarSession(chainID, &wallet, account, passphrase, reverseRegistrar, gasPrice)
		if nonce != -1 {
			session.TransactOpts.Nonce = big.NewInt(nonce)
		}

		// Clean up the name prior to setting
		nameSetName = ens.Normalize(nameSetName)

		tx, err := ens.SetName(session, nameSetName)
		cli.ErrCheck(err, quiet, "Failed to set name for that address")
		if !quiet {
			fmt.Println("Transaction ID is", tx.Hash().Hex())
		}
		log.WithFields(log.Fields{"transactionid": tx.Hash().Hex(),
			"networkid": chainID,
			"address":   args[0],
			"name":      nameSetName}).Info("Name set")
	},
}

func init() {
	nameCmd.AddCommand(nameSetCmd)
	nameSetCmd.Flags().StringVar(&nameSetName, "name", "", "Name to set")

	addTransactionFlags(nameSetCmd, "Passphrase for the account that owns the name")
}
