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
	"strings"

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
		registryContract, err := ens.RegistryContract(client)
		cli.ErrCheck(err, quiet, "Failed to obtain registry contract")
		if !quiet {
			registryContractAddress, err := ens.RegistryContractAddress(client)
			if err == nil {
				fmt.Println("Registry contract at", registryContractAddress.Hex())
			}
		}
		registrarContract, err := ens.RegistrarContract(client)
		cli.ErrCheck(err, quiet, "Failed to obtain registrar contract")
		if !quiet {
			registrarContractAddress, err := ens.RegistrarContractAddress(client)
			if err == nil {
				fmt.Println("Registrar contract at", registrarContractAddress.Hex())
			}
		}
		state, deedAddress, registrationDate, value, highestBid, err := ens.Entry(registrarContract, client, args[0])
		cli.ErrCheck(err, quiet, fmt.Sprintf("Cannot obtain raw info for %s", args[0]))
		if quiet {
			if state == "Owned" {
				os.Exit(0)
			} else {
				os.Exit(1)
			}
		} else {
			nameHash := ens.NameHash(args[0])
			fmt.Println("\nRegistry")
			fmt.Println("~~~~~~~~")
			registryOwner, err := registryContract.Owner(nil, nameHash)
			if err == nil {
				fmt.Println("Owner:", registryOwner.Hex())
			}
			resolver, err := registryContract.Resolver(nil, nameHash)
			if err == nil {
				fmt.Println("Resolver:", resolver.Hex())
			}
			nameParts := strings.Split(args[0], ".")
			for i := len(nameParts) - 1; i >= 0; i-- {
				namePart := strings.Join(nameParts[i:], ".")
				fmt.Printf("\n%s hashes\n", namePart)
				fmt.Printf("%s~~~~~~~\n", strings.Repeat("~", len(namePart)))
				labelHash := ens.LabelHash(nameParts[i])
				fmt.Printf("LabelHash: 0x%s\n", hex.EncodeToString(labelHash[:]))
				nameHash := ens.NameHash(namePart)
				fmt.Printf("NameHash: 0x%s\n", hex.EncodeToString(nameHash[:]))
			}
			fmt.Println("\nEntry")
			fmt.Println("~~~~~")
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
