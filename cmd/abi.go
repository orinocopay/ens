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

	"github.com/orinocopay/go-etherutils/cli"
	"github.com/orinocopay/go-etherutils/ens"
	"github.com/spf13/cobra"
)

// abiCmd represents the abi command
var abiCmd = &cobra.Command{
	Use:   "abi",
	Short: "Obtain the ABI associated with an ENS name",
	Long: `Obtain the ABI associated with a name registered with the Ethereum Name Service (ENS).  For example:

	ens abi enstest.eth

In quiet mode this will return 0 if the name has an ABI, otherwise 1.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Obtain the resolver for this name
		resolverAddress, err := ens.Resolver(registryContract, args[0])
		cli.ErrCheck(err, quiet, "No resolver for that name")

		// Set the address to which we resolve
		resolverContract, err := ens.ResolverContractByAddress(client, resolverAddress)
		cli.ErrCheck(err, quiet, "Failed to obtain resolver contract")

		// Fetch the ABI
		abi, err := ens.Abi(resolverContract, args[0])
		cli.ErrCheck(err, quiet, "Failed to obtain ABI")
		if !quiet {
			fmt.Println(string(abi))
		}
	},
}

func init() {
	RootCmd.AddCommand(abiCmd)
}
