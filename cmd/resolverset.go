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
	"github.com/spf13/cobra"
)

var resolverAddressStr string

// resolverSetCmd represents the resolver set command
var resolverSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set the resolver of an ENS name",
	Long: `Set the resolver of a name registered with the Ethereum Name Service (ENS).  For example:

    ens resolver set --address=0x5FfC014343cd971B7eb70732021E26C35B744cc4 --passphrase="my secret passphrase" enstest.eth

If the address is not supplied then the public resolver for the network will be used.

The keystore for the account that owns the name must be local (i.e. listed with 'get accounts list') and unlockable with the supplied passphrase.

In quiet mode this will return 0 if the transaction to set the resolver is sent successfully, otherwise 1.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Ensure that the name is in a suitable state
		if ens.DomainLevel(args[0]) == 1 {
			cli.Assert(inState(args[0], "Owned"), true, "Domain not in a suitable state to set a resolver")
		}

		// Fetch the owner of the name
		owner, err := registryContract.Owner(nil, ens.NameHash(args[0]))
		cli.ErrCheck(err, quiet, "Cannot obtain owner")
		cli.Assert(bytes.Compare(owner.Bytes(), ens.UnknownAddress.Bytes()) != 0, quiet, "Owner is not set")

		// Fetch the wallet and account for the owner
		wallet, account, err := obtainWalletAndAccount(owner, passphrase)
		cli.ErrCheck(err, quiet, "Failed to obtain account details for the owner of the domain")

		gasPrice, err := etherutils.StringToWei(gasPriceStr)
		cli.ErrCheck(err, quiet, "Invalid gas price")

		// Set up our session
		session := ens.CreateRegistrySession(chainID, &wallet, account, passphrase, registryContract, gasPrice)
		if err != nil {
			// No registry
			return
		}

		// Set the resolver from either command-line or default
		resolverAddress, err := ens.Resolve(client, resolverAddressStr)
		if err != nil {
			resolverAddress, err = ens.PublicResolver(client)
			cli.ErrCheck(err, quiet, "No public resolver for that network")
		}
		tx, err := ens.SetResolver(session, args[0], &resolverAddress)
		cli.ErrCheck(err, quiet, "Failed to send transaction")
		if !quiet {
			fmt.Println("Transaction ID is", tx.Hash().Hex())
		}
	},
}

func init() {
	resolverCmd.AddCommand(resolverSetCmd)

	resolverSetCmd.Flags().StringVarP(&passphrase, "passphrase", "p", "", "Passphrase for the account that owns the name")
	resolverSetCmd.Flags().StringVarP(&resolverAddressStr, "address", "a", "", "Address of the resolver")
	resolverSetCmd.Flags().StringVarP(&gasPriceStr, "gasprice", "g", "4 GWei", "Gas price for the transaction")
}
