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
	"math/big"
	"os"
	"time"

	etherutils "github.com/orinocopay/go-etherutils"
	"github.com/orinocopay/go-etherutils/cli"
	"github.com/orinocopay/go-etherutils/ens"
	"github.com/orinocopay/go-etherutils/ens/registrarcontract"
	"github.com/spf13/cobra"
)

var zero = big.NewInt(0)

// infoCmd represents the info command
var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Obtain information about an ENS domain",
	Long: `Obtain information about a domain registered with the Ethereum Name Service (ENS).  For example:

    ens info enstest.eth

In quiet mode this will return 0 if the domain is owned, otherwise 1.`,

	Run: func(cmd *cobra.Command, args []string) {
		registrarContract, err := ens.RegistrarContract(client)
		cli.ErrCheck(err, quiet, "Failed to obtain registrar contract")
		if ens.DomainLevel(args[0]) == 1 {
			state, err := ens.State(registrarContract, client, args[0])
			cli.ErrCheck(err, quiet, "Cannot obtain info")
			if quiet {
				if state == "Owned" {
					os.Exit(0)
				} else {
					os.Exit(1)
				}
			} else {
				switch state {
				case "Available":
					availableInfo(args[0])
				case "Bidding":
					biddingInfo(registrarContract, args[0])
				case "Revealing":
					revealingInfo(registrarContract, args[0])
				case "Won":
					wonInfo(registrarContract, args[0])
				case "Owned":
					ownedInfo(registrarContract, args[0])
				default:
					fmt.Println(state)
				}
			}
		} else {
			subdomainInfo(registrarContract, args[0])
		}

	},
}

func init() {
	RootCmd.AddCommand(infoCmd)
}

func availableInfo(name string) {
	if len(name) < 11 { // 7 + 4 for '.eth'
		fmt.Println("Unavailable due to name length restrictions")
	} else {
		fmt.Println("Available")
	}
}

func biddingInfo(registrar *registrarcontract.RegistrarContract, name string) {
	_, _, registrationDate, _, _, err := ens.Entry(registrar, client, name)
	cli.ErrCheck(err, quiet, "Cannot obtain auction status")
	twoDaysAgo := time.Duration(-48) * time.Hour
	fmt.Println("Bidding until", registrationDate.Add(twoDaysAgo))
}

func revealingInfo(registrar *registrarcontract.RegistrarContract, name string) {
	_, _, registrationDate, value, highestBid, err := ens.Entry(registrar, client, name)
	cli.ErrCheck(err, quiet, "Cannot obtain information for that name")
	fmt.Println("Revealing until", registrationDate)
	// If the value is 0 then it is is minvalue instead
	if value.Cmp(zero) == 0 {
		value, _ = etherutils.StringToWei("0.01 ether")
	}
	fmt.Println("Locked value is", etherutils.WeiToString(value, true))
	fmt.Println("Highest bid is", etherutils.WeiToString(highestBid, true))
	// TODO number of bids revealed?
}

func wonInfo(registrar *registrarcontract.RegistrarContract, name string) {
	_, deedAddress, registrationDate, value, highestBid, err := ens.Entry(registrar, client, name)
	cli.ErrCheck(err, quiet, "Cannot obtain information for that name")
	fmt.Println("Won since", registrationDate)
	if value.Cmp(zero) == 0 {
		value, _ = etherutils.StringToWei("0.01 ether")
	}
	fmt.Println("Locked value is", etherutils.WeiToString(value, true))
	fmt.Println("Highest bid was", etherutils.WeiToString(highestBid, true))

	// Deed
	deedContract, err := ens.DeedContract(client, &deedAddress)
	cli.ErrCheck(err, quiet, "Failed to obtain deed contract")
	// Deed owner
	deedOwner, err := deedContract.Owner(nil)
	cli.ErrCheck(err, quiet, "Failed to obtain deed owner")
	deedOwnerName, _ := ens.ReverseResolve(client, &deedOwner)
	if deedOwnerName == "" {
		fmt.Println("Deed owner is", deedOwner.Hex())
	} else {
		fmt.Printf("Deed owner is %s (%s)\n", deedOwnerName, deedOwner.Hex())
	}
}

func ownedInfo(registrar *registrarcontract.RegistrarContract, name string) {
	_, deedAddress, registrationDate, value, highestBid, err := ens.Entry(registrar, client, name)
	cli.ErrCheck(err, quiet, "Cannot obtain information for that name")
	fmt.Println("Owned since", registrationDate)
	fmt.Println("Locked value is", etherutils.WeiToString(value, true))
	fmt.Println("Highest bid was", etherutils.WeiToString(highestBid, true))

	// Deed
	deedContract, err := ens.DeedContract(client, &deedAddress)
	cli.ErrCheck(err, quiet, "Failed to obtain deed contract")
	// Deed owner
	deedOwner, err := deedContract.Owner(nil)
	cli.ErrCheck(err, quiet, "Failed to obtain deed owner")
	deedOwnerName, _ := ens.ReverseResolve(client, &deedOwner)
	if deedOwnerName == "" {
		fmt.Println("Deed owner is", deedOwner.Hex())
	} else {
		fmt.Printf("Deed owner is %s (%s)\n", deedOwnerName, deedOwner.Hex())
	}

	// Address owner
	registry, err := ens.RegistryContract(client)
	cli.ErrCheck(err, quiet, "Failed to obtain registry contract")
	domainOwnerAddress, err := registry.Owner(nil, ens.NameHash(name))
	cli.ErrCheck(err, quiet, "Failed to obtain domain owner")
	if domainOwnerAddress == ens.UnknownAddress {
		fmt.Println("Address owner not set")
		return
	}
	domainOwnerName, _ := ens.ReverseResolve(client, &domainOwnerAddress)
	if domainOwnerName == "" {
		fmt.Println("Address owner is", domainOwnerAddress.Hex())
	} else {
		fmt.Printf("Address owner is %s (%s)\n", domainOwnerName, domainOwnerAddress.Hex())
	}

	// Resolver
	resolverAddress, err := ens.Resolver(registry, name)
	if err != nil {
		fmt.Println("Resolver not configured")
		return
	}
	resolverName, _ := ens.ReverseResolve(client, &resolverAddress)
	if resolverName == "" {
		fmt.Println("Resolver is", resolverAddress.Hex())
	} else {
		fmt.Printf("Resolver is %s (%s)\n", resolverName, resolverAddress.Hex())
	}

	// Address
	address, err := ens.Resolve(client, name)
	if err != nil || address == ens.UnknownAddress {
		fmt.Println("Name does not resolve to an address")
		return
	}
	fmt.Println("Domain resolves to", address.Hex())

	// Reverse resolution
	reverseDomain, err := ens.ReverseResolve(client, &address)
	if err != nil || reverseDomain == "" {
		fmt.Println("Address does not resolve to a domain")
		return
	}
	fmt.Println("Address resolves to", reverseDomain)

	// TODO Other common fields (addr, abi, etc.) (if configured)
}

func subdomainInfo(registrar *registrarcontract.RegistrarContract, name string) {
	// Address owner
	registry, err := ens.RegistryContract(client)
	cli.ErrCheck(err, quiet, "Failed to obtain registry contract")
	domainOwnerAddress, err := registry.Owner(nil, ens.NameHash(name))
	cli.ErrCheck(err, quiet, "Failed to obtain domain owner")
	if domainOwnerAddress == ens.UnknownAddress {
		fmt.Println("Address owner not set")
		return
	}
	domainOwnerName, _ := ens.ReverseResolve(client, &domainOwnerAddress)
	if domainOwnerName == "" {
		fmt.Println("Address owner is", domainOwnerAddress.Hex())
	} else {
		fmt.Printf("Address owner is %s (%s)\n", domainOwnerName, domainOwnerAddress.Hex())
	}

	// Resolver
	resolverAddress, err := ens.Resolver(registry, name)
	if err != nil {
		fmt.Println("Resolver not configured")
		return
	}
	resolverName, _ := ens.ReverseResolve(client, &resolverAddress)
	if resolverName == "" {
		fmt.Println("Resolver is", resolverAddress.Hex())
	} else {
		fmt.Printf("Resolver is %s (%s)\n", resolverName, resolverAddress.Hex())
	}

	// Address
	address, err := ens.Resolve(client, name)
	if err != nil || address == ens.UnknownAddress {
		fmt.Println("Name does not resolve to an address")
		return
	}
	fmt.Println("Domain resolves to", address.Hex())

	// Reverse resolution
	reverseDomain, err := ens.ReverseResolve(client, &address)
	if err != nil || reverseDomain == "" {
		fmt.Println("Address does not resolve to a domain")
		return
	}
	fmt.Println("Address resolves to", reverseDomain)

	// TODO Other common fields (addr, abi, etc.) (if configured)
}
