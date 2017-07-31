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

	"github.com/orinocopay/go-etherutils/ens"
	"github.com/spf13/cobra"
)

// hashCmd represents the hash command
var hashCmd = &cobra.Command{
	Use:   "hash",
	Short: "Obtain the ENS namehash of a name",
	Long: `Obtain the ENS namehash of a name.  For example:

	ens hash foo.eth

In quiet mode this will return 0 if the name can be hashed, otherwise 1.`,
	Run: func(cmd *cobra.Command, args []string) {
		name, err := ens.NameHash(args[0])
		if quiet {
			if err != nil {
				os.Exit(1)
			} else {
				os.Exit(0)
			}
		}

		if !quiet {
			fmt.Println(hex.EncodeToString(name[:]))
		}
	},
}

func init() {
	RootCmd.AddCommand(hashCmd)
}
