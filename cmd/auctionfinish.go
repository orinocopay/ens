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

	etherutils "github.com/orinocopay/go-etherutils"
	"github.com/orinocopay/go-etherutils/cli"
	"github.com/orinocopay/go-etherutils/ens"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// auctionFinishCmd represents the auction reveal command
var auctionFinishCmd = &cobra.Command{
	Use:   "finish",
	Short: "Finish an auction for an ENS name",
	Long: `Finish an auction for a name with the Ethereum Name Service (ENS).  For example:

    ens auction finish --address=0x5FfC014343cd971B7eb70732021E26C35B744cc4 --passphrase="my secret passphrase" enstest.eth

The keystore for the address must be local (i.e. listed with 'get accounts list') and unlockable with the supplied passphrase.

In quiet mode this will return 0 if the transaction to finish the auction is sent successfully, otherwise 1.`,
	Run: func(cmd *cobra.Command, args []string) {

		// Ensure that the name is in a suitable state
		cli.Assert(inState(args[0], "Won"), quiet, "Domain not in a suitable state to finish the auction")

		// Fetch the owner of the name - must be 0 if this auction has not been finalised
		owner, err := registryContract.Owner(nil, ens.NameHash(args[0]))
		cli.ErrCheck(err, quiet, "Cannot obtain owner")
		cli.Assert(bytes.Compare(owner.Bytes(), ens.UnknownAddress.Bytes()) == 0, quiet, "Auction already finished")

		_, deedAddress, _, _, _, err := ens.Entry(registrarContract, client, args[0])
		cli.ErrCheck(err, quiet, "Cannot obtain information for that auction")

		// Fetch the owner of the deed that won the address
		// Deed
		deedContract, err := ens.DeedContract(client, &deedAddress)
		cli.ErrCheck(err, quiet, "Failed to obtain deed contract")
		// Deed owner
		deedOwner, err := deedContract.Owner(nil)
		cli.ErrCheck(err, quiet, "Failed to obtain deed owner")

		// Fetch the wallet and account for the owner
		wallet, account, err := obtainWalletAndAccount(deedOwner, passphrase)
		cli.ErrCheck(err, quiet, "Failed to obtain account details for the owner of the name")

		gasPrice, err := etherutils.StringToWei(gasPriceStr)
		cli.ErrCheck(err, quiet, "Invalid gas price")

		// Set up our session
		session := ens.CreateRegistrarSession(chainID, &wallet, account, passphrase, registrarContract, gasPrice)

		// Finish the bid
		tx, err := ens.FinishAuction(session, args[0])
		cli.ErrCheck(err, quiet, "Failed to send transaction")
		if !quiet {
			fmt.Println("Transaction ID is", tx.Hash().Hex())
		}
		log.WithFields(log.Fields{"transactionid": tx.Hash().Hex(),
			"networkid": chainID,
			"name":      args[0]}).Info("Auction finish")

	},
}

func init() {
	auctionCmd.AddCommand(auctionFinishCmd)

	auctionFinishCmd.Flags().StringVarP(&passphrase, "passphrase", "p", "", "Passphrase for the account that owns the winning address")
	auctionFinishCmd.Flags().StringVarP(&gasPriceStr, "gasprice", "g", "4 GWei", "Gas price for the transaction")
}
