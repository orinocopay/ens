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

var abiSetAbi string
var abiSetCompressed bool

// abiSetCmd represents the address set command
var abiSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set the address of an ENS name",
	Long: `Set the address of a name registered with the Ethereum Name Service (ENS).  For example:

    ens address set --address=0x90f8bf6a479f320ead074411a4b0e7944ea8c9c1 --passphrase="my secret passphrase" enstest.eth

The keystore for the account that owns the name must be local (i.e. listed with 'get accounts list') and unlockable with the supplied passphrase.

In quiet mode this will return 0 if the transaction to set the address is sent successfully, otherwise 1.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Ensure that the name is in a suitable state
		cli.Assert(inState(args[0], "Owned"), quiet, "Domain not in a suitable state to set an address")

		// Fetch the owner of the name
		owner, err := registryContract.Owner(nil, ens.NameHash(args[0]))
		cli.ErrCheck(err, quiet, "Cannot obtain owner")
		cli.Assert(bytes.Compare(owner.Bytes(), ens.UnknownAddress.Bytes()) != 0, quiet, "Owner is not set")

		// Fetch the wallet and account for the owner
		wallet, account, err := obtainWalletAndAccount(owner, passphrase)
		cli.ErrCheck(err, quiet, "Failed to obtain account details for the owner of the name")

		gasPrice, err := etherutils.StringToWei(gasPriceStr)
		cli.ErrCheck(err, quiet, "Invalid gas price")

		// Obtain the resolver for this name
		resolverAddress, err := ens.Resolver(registryContract, args[0])
		cli.ErrCheck(err, quiet, "No resolver for that name")

		// Set the address to which we resolve
		resolverContract, err := ens.ResolverContractByAddress(client, resolverAddress)
		cli.ErrCheck(err, quiet, "Failed to obtain resolver contract")
		resolverSession := ens.CreateResolverSession(chainID, &wallet, account, passphrase, resolverContract, gasPrice)
		var contentType = big.NewInt(1)
		if abiSetCompressed {
			contentType = big.NewInt(2)
		}
		tx, err := ens.SetAbi(resolverSession, args[0], abiSetAbi, contentType)
		cli.ErrCheck(err, quiet, "Failed to set ABI for that name")
		if !quiet {
			fmt.Println("Transaction ID is", tx.Hash().Hex())
		}
		log.WithFields(log.Fields{"transactionid": tx.Hash().Hex(),
			"networkid": chainID,
			"name":      args[0],
			"abi":       abiSetAbi}).Info("ABI set")
	},
}

func init() {
	abiCmd.AddCommand(abiSetCmd)

	abiSetCmd.Flags().StringVarP(&passphrase, "passphrase", "p", "", "Passphrase for the account that owns the name")
	abiSetCmd.Flags().StringVarP(&abiSetAbi, "abi", "a", "", "ABI to associate with the name")
	abiSetCmd.Flags().StringVarP(&gasPriceStr, "gasprice", "g", "4 GWei", "Gas price for the transaction")
	abiSetCmd.Flags().BoolVarP(&abiSetCompressed, "compressed", "2", false, "Store the ABI in compressed form (content type 2)")
}
