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
	"context"
	"fmt"
	"time"

	"github.com/orinocopay/go-etherutils/cli"
	"github.com/orinocopay/go-etherutils/ens"
	"github.com/spf13/cobra"
)

// nonceCmd represents the nonce command
var nonceCmd = &cobra.Command{
	Use:   "nonce",
	Short: "Obtain the current nonce for an address",
	Long: `Start the auction for a name with the Ethereum Name Service (ENS).  For example:

    ens nonce 0x5FfC014343cd971B7eb70732021E26C35B744cc4

In quiet mode this will return 0 if the nonce can be obtained, otherwise 1.`,
	Run: func(cmd *cobra.Command, args []string) {

		nonceAddress, err := ens.Resolve(client, args[0])
		cli.ErrCheck(err, quiet, "Failed to obtain nonce address")

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		nonce, err := client.PendingNonceAt(ctx, nonceAddress)
		cli.ErrCheck(err, quiet, "Failed to obtain nonce")

		if !quiet {
			fmt.Println(nonce)
		}
	},
}

func init() {
	RootCmd.AddCommand(nonceCmd)
}
