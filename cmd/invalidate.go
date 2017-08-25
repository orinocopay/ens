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

	etherutils "github.com/orinocopay/go-etherutils"
	"github.com/orinocopay/go-etherutils/cli"
	"github.com/orinocopay/go-etherutils/ens"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var invalidatePassphrase string
var invalidateAddressStr string
var invalidateGasPriceStr string

// invalidateCmd represents the status command
var invalidateCmd = &cobra.Command{
	Use:   "invalidate",
	Short: "Invalidate an non-conformant ENS registration",
	Long: `Invalidate an ENS registration that is less than 7 characters. For example:

    ens invalidate --address=0x5FfC014343cd971B7eb70732021E26C35B744cc4 --passphrase="my secret passphrase" bad.eth

In quiet mode this will return 0 if the invalidate transaction has been submitted, otherwise 1.`,

	Run: func(cmd *cobra.Command, args []string) {
		// Ensure that the name is in a suitable state
		registrarContract, err := ens.RegistrarContract(client)
		state, err := ens.State(registrarContract, client, args[0])
		cli.Assert(state == "Won" || state == "Owned", quiet, "Name not in a suitable state to invalidate")

		// Fetch the wallet and account for the address
		invalidateAddress, err := ens.Resolve(client, invalidateAddressStr)
		cli.ErrCheck(err, quiet, "Failed to obtain invalidate address")
		wallet, err := cli.ObtainWallet(chainID, invalidateAddress)
		cli.ErrCheck(err, quiet, "Failed to obtain a wallet for the address")
		account, err := cli.ObtainAccount(&wallet, &invalidateAddress, invalidatePassphrase)
		cli.ErrCheck(err, quiet, "Failed to obtain an account for the address")

		gasPrice, err := etherutils.StringToWei(invalidateGasPriceStr)
		cli.ErrCheck(err, quiet, "Invalid gas price")

		// Set up our session
		session := ens.CreateRegistrarSession(chainID, &wallet, account, invalidatePassphrase, registrarContract, gasPrice)

		tx, err := ens.InvalidateName(session, args[0])
		cli.ErrCheck(err, quiet, "Failed to send transaction")
		if !quiet {
			fmt.Println("Transaction ID is", tx.Hash().Hex())
		}
		log.WithFields(log.Fields{"transactionid": tx.Hash().Hex(),
			"networkid": chainID,
			"name":      args[0]}).Info("Invalidate")

	},
}

func init() {
	RootCmd.AddCommand(invalidateCmd)
	invalidateCmd.Flags().StringVarP(&invalidatePassphrase, "passphrase", "p", "", "Passphrase for the account that will send the invalidate transaction")
	invalidateCmd.Flags().StringVarP(&invalidateAddressStr, "address", "a", "", "Address that will send the invalidate transaction")
	invalidateCmd.Flags().StringVarP(&invalidateGasPriceStr, "gasprice", "g", "4 GWei", "Gas price for the transaction")

}
