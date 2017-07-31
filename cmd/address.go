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

// addressCmd represents the address command
var addressCmd = &cobra.Command{
	Use:   "address",
	Short: "Obtain the address of an ENS name",
	Long: `Obtain the address of a name registered with the Ethereum Name Service (ENS).  For example:

	ens address enstest.eth

In quiet mode this will return 0 if the name resolves correctly, otherwise 1.`,
	Run: func(cmd *cobra.Command, args []string) {
		address, err := ens.Resolve(client, args[0])
		cli.ErrCheck(err, quiet, "Failed to obtain address")
		if !quiet {
			fmt.Println(address.Hex())
		}
	},
}

func init() {
	RootCmd.AddCommand(addressCmd)
}
