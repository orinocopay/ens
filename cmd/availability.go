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

// availabilityCmd represents the availability command
var availabilityCmd = &cobra.Command{
	Use:   "availability",
	Short: "Check availability of an ENS domain.",
	Long: `State if a domain is availabile with the Ethereum Name Service (ENS).  For example:

    ens availability enstest.eth

In quiet mode this will return 0 if the domain is availabile, otherwise 1.`,

	Run: func(cmd *cobra.Command, args []string) {
		registrarContract, err := ens.RegistrarContract(client)
		cli.ErrCheck(err, quiet, "Failed to obtain registrar contract")
		if ens.DomainLevel(args[0]) == 1 {
			// Top-level domain
			state, err := ens.State(registrarContract, client, args[0])
			cli.ErrCheck(err, quiet, "Cannot obtain info")
			if quiet {
				if state == "Available" {
					os.Exit(0)
				} else {
					os.Exit(1)
				}
			} else {
				fmt.Println(state)
			}
		} else {
			// Subdomain
			registry, err := ens.RegistryContract(client)
			cli.ErrCheck(err, quiet, "Failed to obtain registry contract")
			subdomainOwnerAddress, err := registry.Owner(nil, ens.NameHash(args[0]))
			cli.ErrCheck(err, quiet, "Failed to obtain subdomain owner")
			if quiet {
				if subdomainOwnerAddress == ens.UnknownAddress {
					os.Exit(0)
				} else {
					os.Exit(1)
				}
			} else {
				if subdomainOwnerAddress == ens.UnknownAddress {
					fmt.Println("Available")
				} else {
					fmt.Println("Owned")
				}
			}
		}
	},
}

func init() {
	RootCmd.AddCommand(availabilityCmd)
}
