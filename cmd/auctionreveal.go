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

var auctionRevealPassphrase string
var auctionRevealAddressStr string
var auctionRevealGasPriceStr string
var auctionRevealBidPriceStr string
var auctionRevealSalt string

// auctionRevealCmd represents the auctionReveal set command
var auctionRevealCmd = &cobra.Command{
	Use:   "reveal",
	Short: "Reveal a bid in an auction for an ENS name",
	Long: `Reveal the auction for a name with the Ethereum Name Service (ENS).  For example:

    ens auction reveal --address=0x5FfC014343cd971B7eb70732021E26C35B744cc4 --passphrase="my secret passphrase" --bid="0.01 Ether" enstest.eth

The keystore for the address must be local (i.e. listed with 'get accounts list') and unlockable with the supplied passphrase.

In quiet mode this will return 0 if the transaction to reveal the bid is sent successfully, otherwise 1.`,
	Run: func(cmd *cobra.Command, args []string) {
		cli.Assert(auctionRevealSalt != "", quiet, "Salt is required")

		// Ensure that the name is in a suitable state
		registrarContract, err := ens.RegistrarContract(client, rpcclient)
		inState, err := ens.NameInState(registrarContract, args[0], "Revealing")
		cli.ErrAssert(inState, err, quiet, "Name not in a suitable state for bid to be revealed")

		// Fetch the wallet and account for the address
		auctionRevealAddress, err := ens.Resolve(client, auctionRevealAddressStr, rpcclient)
		cli.ErrCheck(err, quiet, "Failed to obtain auction address")
		wallet, err := cli.ObtainWallet(chainID, auctionRevealAddress)
		cli.ErrCheck(err, quiet, "Failed to obtain a wallet for the address")
		account, err := cli.ObtainAccount(wallet, auctionRevealAddress, auctionRevealPassphrase)
		cli.ErrCheck(err, quiet, "Failed to obtain an account for the address")

		gasLimit := big.NewInt(500000)
		gasPrice, err := etherutils.StringToWei(auctionRevealGasPriceStr)
		cli.ErrCheck(err, quiet, "Invalid gas price")

		// Set up our session
		session := ens.CreateRegistrarSession(chainID, &wallet, account, auctionRevealPassphrase, registrarContract, gasLimit, gasPrice)

		bidPrice, err := etherutils.StringToWei(auctionRevealBidPriceStr)
		cli.ErrCheck(err, quiet, "Invalid bid price")

		// Reveal the bid
		tx, err := ens.RevealBid(session, args[0], &auctionRevealAddress, *bidPrice, auctionRevealSalt)
		cli.ErrCheck(err, quiet, "Failed to send transaction")
		if !quiet {
			fmt.Println("Transaction ID is", tx.Hash().Hex())
		}
		log.WithFields(log.Fields{"transactionid": tx.Hash().Hex(),
			"networkid": chainID,
			"name":      args[0],
			"address":   auctionRevealAddress.Hex(),
			"salt":      auctionRevealSalt,
			"bid":       bidPrice}).Info("Auction reveal")

	},
}

func init() {
	auctionCmd.AddCommand(auctionRevealCmd)

	auctionRevealCmd.Flags().StringVarP(&auctionRevealPassphrase, "passphrase", "p", "", "Passphrase for the account that owns the bidding address")
	auctionRevealCmd.Flags().StringVarP(&auctionRevealAddressStr, "address", "a", "", "Address doing the bidding")
	auctionRevealCmd.Flags().StringVarP(&auctionRevealGasPriceStr, "gasprice", "g", "20 GWei", "Gas price for the transaction")
	auctionRevealCmd.Flags().StringVarP(&auctionRevealBidPriceStr, "bid", "b", "0.01 Ether", "Bid price for the name")
	auctionRevealCmd.Flags().StringVarP(&auctionRevealSalt, "salt", "s", "", "Memorable phrase needed when revealing bid")
}
