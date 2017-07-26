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

	"github.com/ethereum/go-ethereum/common"
	"github.com/orinocopay/go-etherutils/cli"
	"github.com/orinocopay/go-etherutils/ens"
	"github.com/spf13/cobra"
)

// nameCmd represents the name command
var nameCmd = &cobra.Command{
	Use:   "name",
	Short: "Obtain the ENS name of an address",
	Long: `Obtain the name registered with the Ethereum Name Service (ENS) for an address.  For example:

	ens name 0xe40626310e0726e45041ac34094037f30d2a9cc3

In quiet mode this will return 0 if the address resolves correctly, otherwise 1.`,
	Run: func(cmd *cobra.Command, args []string) {
		address := common.HexToAddress(args[0])
		name, err := ens.Reverse(client, &address, rpcclient)
		cli.ErrCheck(err, quiet, "Failed to obtain name")
		if !quiet {
			fmt.Println(name)
		}
	},
}

func init() {
	RootCmd.AddCommand(nameCmd)
}
