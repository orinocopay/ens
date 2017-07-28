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

	etherutils "github.com/orinocopay/go-etherutils"
	"github.com/orinocopay/go-etherutils/cli"
	"github.com/orinocopay/go-etherutils/ens"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var auctionStartPassphrase string
var auctionStartAddressStr string
var auctionStartGasPriceStr string
var auctionStartBidPriceStr string
var auctionStartMaskPriceStr string
var auctionStartSalt string

// auctionStartCmd represents the auctionStart set command
var auctionStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the auction for an ENS name",
	Long: `Start the auction for a name with the Ethereum Name Service (ENS).  For example:

    ens auction start --address=0x5FfC014343cd971B7eb70732021E26C35B744cc4 --passphrase="my secret passphrase" --bid="0.01 Ether" enstest.eth

The keystore for the address must be local (i.e. listed with 'get accounts list') and unlockable with the supplied passphrase.

In quiet mode this will return 0 if the transaction to start the auction is sent successfully, otherwise 1.`,
	Run: func(cmd *cobra.Command, args []string) {
		cli.Assert(auctionStartSalt != "", quiet, "Salt is required")
		cli.Assert(auctionStartAddressStr != "", quiet, "Address from which to send the bid is required")
		cli.Assert(len(args[0]) > 10, quiet, "Name must be at least 7 characters long")

		// Ensure that the name is in a suitable state
		registrarContract, err := ens.RegistrarContract(client, rpcclient)
		inState, err := ens.NameInState(registrarContract, client, args[0], "Available")
		cli.ErrAssert(inState, err, quiet, "Name not in a suitable state to start an auction")

		// Create the bid

		// Fetch the wallet and account for the address
		auctionStartAddress, err := ens.Resolve(client, auctionStartAddressStr, rpcclient)
		cli.ErrCheck(err, quiet, "Failed to obtain auction address")
		wallet, err := cli.ObtainWallet(chainID, auctionStartAddress)
		cli.ErrCheck(err, quiet, "Failed to obtain a wallet for the address")
		account, err := cli.ObtainAccount(wallet, auctionStartAddress, auctionStartPassphrase)
		cli.ErrCheck(err, quiet, "Failed to obtain an account for the address")

		gasLimit := big.NewInt(500000)
		gasPrice, err := etherutils.StringToWei(auctionStartGasPriceStr)
		cli.ErrCheck(err, quiet, "Invalid gas price")

		// Set up our session
		session := ens.CreateRegistrarSession(chainID, &wallet, account, auctionStartPassphrase, registrarContract, gasLimit, gasPrice)

		bidPrice, err := etherutils.StringToWei(auctionStartBidPriceStr)
		cli.ErrCheck(err, quiet, "Invalid bid price")
		// Start the auction
		bidMask, err := etherutils.StringToWei(auctionStartMaskPriceStr)
		if err != nil {
			bidMask = big.NewInt(0)
			bidMask.Set(bidPrice)
		} else if bidMask.Cmp(bidPrice) == -1 {
			bidMask.Set(bidPrice)
		}

		session.TransactOpts.Value = bidMask
		tx, err := ens.StartAuctionAndBid(session, args[0], &auctionStartAddress, *bidPrice, auctionStartSalt)
		session.TransactOpts.Value = big.NewInt(0)
		cli.ErrCheck(err, quiet, "Failed to send transaction")
		if !quiet {
			fmt.Println("Transaction ID is", tx.Hash().Hex())
		}
		log.WithFields(log.Fields{"transactionid": tx.Hash().Hex(),
			"name":      args[0],
			"networkid": chainID,
			"address":   auctionStartAddress.Hex(),
			"salt":      auctionStartSalt,
			"bid":       bidPrice,
			"mask":      bidMask}).Info("Auction start")
	},
}

func init() {
	auctionCmd.AddCommand(auctionStartCmd)

	auctionStartCmd.Flags().StringVarP(&auctionStartPassphrase, "passphrase", "p", "", "Passphrase for the account that owns the bidding address")
	auctionStartCmd.Flags().StringVarP(&auctionStartAddressStr, "address", "a", "", "Address doing the bidding")
	auctionStartCmd.Flags().StringVarP(&auctionStartGasPriceStr, "gasprice", "g", "20 GWei", "Gas price for the transaction")
	auctionStartCmd.Flags().StringVarP(&auctionStartBidPriceStr, "bid", "b", "0.01 Ether", "Bid price for the name")
	auctionStartCmd.Flags().StringVarP(&auctionStartMaskPriceStr, "mask", "m", "", "Amount of Ether sent in the transaction (must be at least the bid)")
	auctionStartCmd.Flags().StringVarP(&auctionStartSalt, "salt", "s", "", "Memorable phrase needed when revealing bid")

}
