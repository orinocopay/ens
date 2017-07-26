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

var addressPassphrase string
var addressAddressStr string
var addressGasPriceStr string

// addressSetCmd represents the address set command
var addressSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set the address of an ENS name",
	Long: `Set the address of a name registered with the Ethereum Name Service (ENS).  For example:

    ens address set --address=0x90f8bf6a479f320ead074411a4b0e7944ea8c9c1 --passphrase="my secret passphrase" enstest.eth

The keystore for the account that owns the name must be local (i.e. listed with 'get accounts list') and unlockable with the supplied passphrase.

In quiet mode this will return 0 if the transaction to set the address is sent successfully, otherwise 1.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Ensure that the name is in a suitable state
		registrarContract, err := ens.RegistrarContract(client, rpcclient)
		inState, err := ens.NameInState(registrarContract, args[0], "Owned")
		cli.ErrAssert(inState, err, quiet, "Name not in a suitable state to set an address")

		// Obtain the registry contract
		registryContract, err := ens.RegistryContract(client, rpcclient)
		cli.ErrCheck(err, quiet, "Failed to obtain registry contract")

		// Fetch the owner of the name
		nameHash, err := ens.NameHash(args[0])
		cli.ErrCheck(err, quiet, "Invalid name")
		owner, err := registryContract.Owner(nil, nameHash)
		cli.ErrCheck(err, quiet, "Cannot obtain owner")
		cli.Assert(bytes.Compare(owner.Bytes(), ens.UnknownAddress.Bytes()) != 0, quiet, "Owner is not set")

		// Fetch the wallet and account for the owner
		wallet, err := cli.ObtainWallet(chainID, owner)
		cli.ErrCheck(err, quiet, "Failed to obtain a wallet for the owner")
		account, err := cli.ObtainAccount(wallet, owner, addressPassphrase)
		cli.ErrCheck(err, quiet, "Failed to obtain an account for the owner")

		gasLimit := big.NewInt(50000)
		gasPrice, err := etherutils.StringToWei(addressGasPriceStr)
		cli.ErrCheck(err, quiet, "Invalid gas price")

		// Obtain the resolver for this name
		resolverAddress, err := ens.Resolver(registryContract, args[0])
		cli.ErrCheck(err, quiet, "No resolver for that name")

		// Obtain the address to which we resolve
		resolutionAddress, err := ens.Resolve(client, addressAddressStr, rpcclient)
		cli.ErrCheck(err, quiet, "Invalid address")

		// Set the address to which we resolve
		resolverContract, err := ens.ResolverContractByAddress(client, resolverAddress)
		cli.ErrCheck(err, quiet, "Failed to obtain resolver contract")
		resolverSession := ens.CreateResolverSession(chainID, &wallet, account, addressPassphrase, resolverContract, gasLimit, gasPrice)
		tx, err := ens.SetResolution(resolverSession, args[0], &resolutionAddress)
		cli.ErrCheck(err, quiet, "Failed to set resolution for that name")
		if !quiet {
			fmt.Println("Transaction ID is", tx.Hash().Hex())
		}
		log.WithFields(log.Fields{"transactionid": tx.Hash().Hex(),
			"networkid": chainID,
			"name":      args[0],
			"address":   resolutionAddress.Hex()}).Info("Address set")

	},
}

func init() {
	addressCmd.AddCommand(addressSetCmd)

	addressSetCmd.Flags().StringVarP(&addressPassphrase, "passphrase", "p", "", "Passphrase for the account that owns the name")
	addressSetCmd.Flags().StringVarP(&addressAddressStr, "address", "a", "", "Address of the resolver")
	addressSetCmd.Flags().StringVarP(&addressGasPriceStr, "gasprice", "g", "20 GWei", "Gas price for the transaction")
}
