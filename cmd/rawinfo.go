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
	"encoding/hex"
	"fmt"
	"os"

	"github.com/orinocopay/go-etherutils/cli"
	"github.com/orinocopay/go-etherutils/ens"
	"github.com/spf13/cobra"
)

// rawInfoCmd represents the info command
var rawInfoCmd = &cobra.Command{
	Use:   "rawinfo",
	Short: "Obtain raw information about an ENS domain direct from the blockchain",
	Long: `Obtain raw information about a domain registered with the Ethereum Name Service (ENS).  For example:

    ens rawinfo enstest.eth

In quiet mode this will return 0 if the domain is owned, otherwise 1.`,

	Run: func(cmd *cobra.Command, args []string) {
		registrarContract, err := ens.RegistrarContract(client)
		cli.ErrCheck(err, quiet, "Failed to obtain registrar contract")
		state, deedAddress, registrationDate, value, highestBid, err := ens.Entry(registrarContract, client, args[0])
		cli.ErrCheck(err, quiet, "Cannot obtain raw info")
		if quiet {
			if state == "Owned" {
				os.Exit(0)
			} else {
				os.Exit(1)
			}
		} else {
			nameHash := ens.NameHash(args[0])
			fmt.Println("NameHash:", hex.EncodeToString(nameHash[:]))
			fmt.Println("State:", state)
			fmt.Println("Deed address:", deedAddress.Hex())
			fmt.Println("Registration date:", registrationDate)
			fmt.Println("Value:", value)
			fmt.Println("Highest bid:", highestBid)
		}
	},
}

func init() {
	RootCmd.AddCommand(rawInfoCmd)
}
