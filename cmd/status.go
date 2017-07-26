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

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Obtain the status of an ENS name",
	Long: `Obtain the status of a name registered with the Ethereum Name Service (ENS).  For example:

    ens status enstest.eth

In quiet mode this will return 0 if the status is owned, otherwise 1.`,

	Run: func(cmd *cobra.Command, args []string) {
		registrarContract, err := ens.RegistrarContract(client, rpcclient)
		state, err := ens.State(registrarContract, args[0])
		cli.ErrCheck(err, quiet, "Cannot obtain status")
		if quiet {
			if state == "Owned" {
				os.Exit(0)
			} else {
				os.Exit(1)
			}
		} else {
			fmt.Println(state)
		}

	},
}

func init() {
	RootCmd.AddCommand(statusCmd)
}
