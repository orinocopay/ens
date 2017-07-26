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
	"github.com/spf13/cobra"
)

// subdomainCmd represents the subdomain command
var subdomainCmd = &cobra.Command{
	Use:   "subdomain",
	Short: "Manage ENS subdomains",
	Long:  `Manage the subdomain of a domain in the Ethereum Name Service.`,
}

func init() {
	RootCmd.AddCommand(subdomainCmd)
}
