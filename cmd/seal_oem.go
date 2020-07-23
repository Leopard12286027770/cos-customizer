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

package cmd

import (
	"context"
	"cos-customizer/config"
	"cos-customizer/fs"
	"flag"

	"github.com/google/subcommands"
)

// SealOEM implements subcommands.Command for the "seal-oem" command.
// It builds a hash tree of the OEM partition and modifies the kernel
// command line to verify the OEM partition at boot time.
type SealOEM struct {
}

// Name implements subcommands.Command.Name.
func (s *SealOEM) Name() string {
	return "seal-oem"
}

// Synopsis implements subcommands.Command.Synopsis.
func (s *SealOEM) Synopsis() string {
	return "Seal the OEM partition."
}

// Usage implements subcommands.Command.Usage.
func (s *SealOEM) Usage() string {
	return `seal-oem
`
}

// SetFlags implements subcommands.Command.SetFlags.
func (s *SealOEM) SetFlags(f *flag.FlagSet) {}

// Execute implements subcommands.Command.Execute. It configures the current image build process to
// customize the result image with a shell script.
func (s *SealOEM) Execute(_ context.Context, f *flag.FlagSet, args ...interface{}) subcommands.ExitStatus {
	if f.NArg() != 0 {
		f.Usage()
		return subcommands.ExitUsageError
	}
	files := args[0].(*fs.Files)
	buildConfig := &config.Build{}
	if err := config.LoadFromFile(files.BuildConfig, buildConfig); err != nil {
		return subcommands.ExitFailure
	}
	buildConfig.SealOEM = true
	return subcommands.ExitSuccess
}
