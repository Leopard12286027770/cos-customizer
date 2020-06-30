// Copyright 2020 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package cmd contains cos-customizer subcommand implementations.
package cmd

import (
	"context"
	"flag"
	"fmt"
	"log"

	"github.com/google/subcommands"
)

// OEMConfig stores parameters needed for setting the OEM partition.
type OEMConfig struct {
	OEMSize string
}

// Name implements subcommands.Command.Name.
func (*OEMConfig) Name() string {
	return "set-oem"
}

// Synopsis implements subcommands.Command.Synopsis.
func (*OEMConfig) Synopsis() string {
	return "Set the OEM partition"
}

// Usage implements subcommands.Command.Usage.
func (*OEMConfig) Usage() string {
	return `set-oem [flags]
`
}

// SetFlags implements subcommands.Command.SetFlags.
func (o *OEMConfig) SetFlags(f *flag.FlagSet) {
	f.StringVar(&o.OEMSize, "size", ".", "Size of the new OEM partition, "+
		"can be a number with unit like 10G, 10M, 10K or 10B, "+
		"or without unit indicating the number of 512B sectors")
}

func (o *OEMConfig) validate() error {
	switch {
	case o.OEMSize == ".":
		return fmt.Errorf("OEM partition size must be set")
	default:
		return nil
	}
}

// Execute implements subcommands.Command.Execute.
// It saves the OEM partition settings.
func (o *OEMConfig) Execute(ctx context.Context, f *flag.FlagSet, args ...interface{}) subcommands.ExitStatus {
	// make sure no unused flags in command line.
	if f.NArg() != 0 {
		f.Usage()
		return subcommands.ExitUsageError
	}
	// files := args[0].(*fs.Files)
	// svc, _, err := args[1].(ServiceClients)(ctx, false)
	// if err != nil {
	// 	log.Println(err)
	// 	return subcommands.ExitFailure
	// }
	if err := o.validate(); err != nil {
		log.Println(err)
		return subcommands.ExitFailure
	}
	log.Println(o.OEMSize)
	// if err := fs.CreateBuildContextArchive(s.buildContext, files.UserBuildContextArchive); err != nil {
	// 	log.Println(err)
	// 	return subcommands.ExitFailure
	// }
	// if err := fs.CreateStateFile(files); err != nil {
	// 	log.Println(err)
	// 	return subcommands.ExitFailure
	// }
	// if err := fs.CreatePersistentBuiltinContext(files); err != nil {
	// 	log.Println(err)
	// 	return subcommands.ExitFailure
	// }

	return subcommands.ExitSuccess
}
