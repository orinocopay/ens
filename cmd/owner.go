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
	"os"

	"github.com/orinocopay/go-etherutils/cli"
	"github.com/orinocopay/go-etherutils/ens"
	"github.com/spf13/cobra"
)

// ownerCmd represents the owner command
var ownerCmd = &cobra.Command{
	Use:   "owner",
	Short: "Obtain owner of an ENS domain",
	Long: `Obtain owner of a domain registered with the Ethereum Name Service (ENS).  For example:

    ens owner enstest.eth

In quiet mode this will return 0 if the domain is owned, otherwise 1.`,

	Run: func(cmd *cobra.Command, args []string) {
		registrarContract, err := ens.RegistrarContract(client)
		cli.ErrCheck(err, quiet, "Failed to obtain registrar contract")
		state, deedAddress, _, _, _, err := ens.Entry(registrarContract, client, args[0])
		cli.ErrCheck(err, quiet, fmt.Sprintf("Cannot obtain raw info for %s", args[0]))
		if quiet {
			if state == "Owned" {
				os.Exit(0)
			} else {
				os.Exit(1)
			}
		} else {
			// Deed
			deedContract, err := ens.DeedContract(client, &deedAddress)
			cli.ErrCheck(err, quiet, "Failed to obtain deed contract")
			// Deed owner
			deedOwner, err := ens.Owner(deedContract)
			cli.ErrCheck(err, quiet, "Failed to obtain deed owner")
			fmt.Println(deedOwner.Hex())
		}
	},
}

func init() {
	RootCmd.AddCommand(ownerCmd)
}
