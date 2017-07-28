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

	etherutils "github.com/orinocopay/go-etherutils"
	"github.com/orinocopay/go-etherutils/cli"
	"github.com/orinocopay/go-etherutils/ens"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var auctionFinishPassphrase string
var auctionFinishGasPriceStr string

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
		registrarContract, err := ens.RegistrarContract(client, rpcclient)
		inState, err := ens.NameInState(registrarContract, client, args[0], "Won")
		cli.ErrAssert(inState, err, quiet, "Auction not in a suitable state to finish")
		// Obtain the registry contract
		registryContract, err := ens.RegistryContract(client, rpcclient)
		cli.ErrCheck(err, quiet, "Failed to obtain registry contract")
		// Fetch the owner of the name - must be 0 if this auction has not been finalised
		nameHash, err := ens.NameHash(args[0])
		cli.ErrCheck(err, quiet, "Invalid name")
		owner, err := registryContract.Owner(nil, nameHash)
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

		// Fetch the wallet and account for the address
		wallet, err := cli.ObtainWallet(chainID, deedOwner)
		cli.ErrCheck(err, quiet, "Failed to obtain a wallet for the address")
		account, err := cli.ObtainAccount(wallet, deedOwner, auctionFinishPassphrase)
		cli.ErrCheck(err, quiet, "Failed to obtain an account for the address")

		gasLimit := big.NewInt(500000)
		gasPrice, err := etherutils.StringToWei(auctionFinishGasPriceStr)
		cli.ErrCheck(err, quiet, "Invalid gas price")

		// Set up our session
		session := ens.CreateRegistrarSession(chainID, &wallet, account, auctionFinishPassphrase, registrarContract, gasLimit, gasPrice)

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

	auctionFinishCmd.Flags().StringVarP(&auctionFinishPassphrase, "passphrase", "p", "", "Passphrase for the account that owns the winning address")
	auctionFinishCmd.Flags().StringVarP(&auctionFinishGasPriceStr, "gasprice", "g", "20 GWei", "Gas price for the transaction")
}
