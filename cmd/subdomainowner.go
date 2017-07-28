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
	"strings"

	etherutils "github.com/orinocopay/go-etherutils"
	"github.com/orinocopay/go-etherutils/cli"
	"github.com/orinocopay/go-etherutils/ens"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var subdomainOwnerPassphrase string
var subdomainOwnerNameStr string
var subdomainOwnerGasPriceStr string
var subdomainOwnerBidPriceStr string
var subdomainOwnerMaskPriceStr string

// subdomainOwnerCmd represents the subdomainOwner set command
var subdomainOwnerCmd = &cobra.Command{
	Use:   "owner",
	Short: "Set the owner of an ENS subdomain",
	Long: `Set the owner of a subdomain for a name with the Ethereum Name Service (ENS).  For example:

    ens subdomain owner --owner=0x5FfC014343cd971B7eb70732021E26C35B744cc4 --passphrase="my secret passphrase" subdomain.enstest.eth

The keystore for the owner of the domain must be local (i.e. listed with 'get accounts list') and unlockable with the supplied passphrase.

In quiet mode this will return 0 if the transaction to set the owner of the subdomain is sent successfully, otherwise 1.`,
	Run: func(cmd *cobra.Command, args []string) {

		// Break the name in to domain and subdomain
		nameBits := strings.Split(args[0], ".")
		cli.Assert(len(nameBits) >= 3, quiet, "Invalid name")
		subdomain := nameBits[0]
		domain := args[0][len(subdomain)+1:]

		// Ensure that the name is in a suitable state
		registrarContract, err := ens.RegistrarContract(client, rpcclient)
		inState, err := ens.NameInState(registrarContract, client, domain, "Owned")
		cli.ErrAssert(inState, err, quiet, "Name not in a suitable state to set a subdomain owner")

		// Obtain the registry contract
		registryContract, err := ens.RegistryContract(client, rpcclient)
		cli.ErrCheck(err, quiet, "Failed to obtain registry contract")

		// Fetch the owner of the domain
		nameHash, err := ens.NameHash(domain)
		cli.ErrCheck(err, quiet, "Invalid name")
		owner, err := registryContract.Owner(nil, nameHash)
		cli.ErrCheck(err, quiet, "Cannot obtain owner")
		cli.Assert(bytes.Compare(owner.Bytes(), ens.UnknownAddress.Bytes()) != 0, quiet, "Owner is not set")

		// Fetch the wallet and account for the owner
		wallet, err := cli.ObtainWallet(chainID, owner)
		cli.ErrCheck(err, quiet, "Failed to obtain a wallet for the owner")
		account, err := cli.ObtainAccount(wallet, owner, subdomainOwnerPassphrase)
		cli.ErrCheck(err, quiet, "Failed to obtain an account for the owner")

		gasLimit := big.NewInt(500000)
		gasPrice, err := etherutils.StringToWei(subdomainOwnerGasPriceStr)
		cli.ErrCheck(err, quiet, "Invalid gas price")

		// Obtain the address who will own the subdomain
		subdomainOwnerAddress, err := ens.Resolve(client, subdomainOwnerNameStr, rpcclient)
		cli.ErrCheck(err, quiet, "Invalid owner")

		// Set up our session
		session := ens.CreateRegistrySession(chainID, &wallet, account, subdomainOwnerPassphrase, registryContract, gasLimit, gasPrice)

		// Set the subdomain owner
		tx, err := ens.SetSubdomainOwner(session, domain, subdomain, &subdomainOwnerAddress)
		cli.ErrCheck(err, quiet, "Failed to send transaction")
		if !quiet {
			fmt.Println("Transaction ID is", tx.Hash().Hex())
		}
		log.WithFields(log.Fields{"transactionid": tx.Hash().Hex(),
			"networkid": chainID,
			"name":      args[0],
			"owner":     subdomainOwnerAddress.Hex()}).Info("Subdomain owner")

	},
}

func init() {
	subdomainCmd.AddCommand(subdomainOwnerCmd)

	subdomainOwnerCmd.Flags().StringVarP(&subdomainOwnerPassphrase, "passphrase", "p", "", "Passphrase for the account that owns the name")
	subdomainOwnerCmd.Flags().StringVarP(&subdomainOwnerNameStr, "owner", "o", "", "Owner of the subdomain")
	subdomainOwnerCmd.Flags().StringVarP(&subdomainOwnerGasPriceStr, "gasprice", "g", "20 GWei", "Gas price for the transaction")
}
