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

var auctionBidAddressStr string
var auctionBidBidPriceStr string
var auctionBidMaskPriceStr string
var auctionBidSalt string

// auctionBidCmd represents the auctionBid set command
var auctionBidCmd = &cobra.Command{
	Use:   "bid",
	Short: "Bid the auction for an ENS name",
	Long: `Bid the auction for a name with the Ethereum Name Service (ENS).  For example:

    ens auction bid --address=0x5FfC014343cd971B7eb70732021E26C35B744cc4 --passphrase="my secret passphrase" --bid="0.01 Ether" enstest.eth

The keystore for the address must be local (i.e. listed with 'get accounts list') and unlockable with the supplied passphrase.

In quiet mode this will return 0 if the transaction to place the bid is sent successfully, otherwise 1.`,
	Run: func(cmd *cobra.Command, args []string) {
		cli.Assert(auctionBidSalt != "", quiet, "Salt is required")
		cli.Assert(auctionBidAddressStr != "", quiet, "Address from which to send the bid is required")

		// Ensure that the name is in a suitable state
		cli.Assert(inState(args[0], "Bidding"), true, "Domain not in a suitable state to bid on an auction")

		// Fetch the wallet and account for the owner
		auctionBidAddress, err := ens.Resolve(client, auctionBidAddressStr)
		cli.ErrCheck(err, quiet, "Failed to obtain auction address")
		wallet, account, err := obtainWalletAndAccount(auctionBidAddress, passphrase)
		cli.ErrCheck(err, quiet, "Failed to obtain account details for the owner of the name")

		gasPrice, err := etherutils.StringToWei(gasPriceStr)
		cli.ErrCheck(err, quiet, "Invalid gas price")

		// Set up our session
		session := ens.CreateRegistrarSession(chainID, &wallet, account, passphrase, registrarContract, gasPrice)

		bidPrice, err := etherutils.StringToWei(auctionBidBidPriceStr)
		cli.ErrCheck(err, quiet, "Invalid bid price")
		// Start the auction
		bidMask, err := etherutils.StringToWei(auctionBidMaskPriceStr)
		if err != nil {
			bidMask = big.NewInt(0)
			bidMask.Set(bidPrice)
		} else if bidMask.Cmp(bidPrice) == -1 {
			bidMask.Set(bidPrice)
		}

		session.TransactOpts.Value = bidMask
		tx, err := ens.NewBid(session, args[0], &auctionBidAddress, *bidPrice, auctionBidSalt)
		session.TransactOpts.Value = big.NewInt(0)
		cli.ErrCheck(err, quiet, "Failed to send transaction")
		if !quiet {
			fmt.Println("Transaction ID is", tx.Hash().Hex())
		}
		log.WithFields(log.Fields{"transactionid": tx.Hash().Hex(),
			"networkid": chainID,
			"name":      args[0],
			"address":   auctionBidAddress.Hex(),
			"salt":      auctionBidSalt,
			"bid":       bidPrice,
			"mask":      bidMask}).Info("Auction bid")
	},
}

func init() {
	auctionCmd.AddCommand(auctionBidCmd)

	auctionBidCmd.Flags().StringVarP(&passphrase, "passphrase", "p", "", "Passphrase for the account that owns the bidding address")
	auctionBidCmd.Flags().StringVarP(&auctionBidAddressStr, "address", "a", "", "Address doing the bidding")
	auctionBidCmd.Flags().StringVarP(&gasPriceStr, "gasprice", "g", "4 GWei", "Gas price for the transaction")
	auctionBidCmd.Flags().StringVarP(&auctionBidBidPriceStr, "bid", "b", "0.01 Ether", "Bid price for the name")
	auctionBidCmd.Flags().StringVarP(&auctionBidMaskPriceStr, "mask", "m", "", "Amount of Ether sent in the transaction (must be at least the bid)")
	auctionBidCmd.Flags().StringVarP(&auctionBidSalt, "salt", "s", "", "Memorable phrase needed when revealing bid")
}
