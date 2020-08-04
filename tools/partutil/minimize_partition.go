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

package partutil

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
)

// MinimizePartition minimizes the input partition and
// returns the next sector of the end sector.
// The smallest partition from fdisk is 1 sector partition.
func MinimizePartition(disk string, partNumInt int) (uint64, error) {
	// Make sure the next partition can start at a 4K aligned sector.
	const minSize = 4096 // 2 MB
	if len(disk) == 0 || partNumInt <= 0 {
		return 0, fmt.Errorf("empty disk name or nonpositive part number, "+
			"input: disk=%q, partNumInt=%d", disk, partNumInt)
	}

	// get partition number string
	partNum, err := PartNumIntToString(disk, partNumInt)
	if err != nil {
		return 0, fmt.Errorf("error in converting partition number, "+
			"input: disk=%q, partNumInt=%d, "+
			"error msg: (%v)", disk, partNumInt, err)
	}

	partName := disk + partNum
	var tableBuffer bytes.Buffer

	// dump partition table.
	table, err := ReadPartitionTable(disk)
	if err != nil {
		return 0, fmt.Errorf("cannot read partition table of %q, "+
			"input: disk=%q, partNumInt=%d, "+
			"error msg: (%v)", disk, disk, partNumInt, err)
	}

	var startSector uint64

	// edit partition table.
	table, err = ParsePartitionTable(table, partName, true, func(p *PartContent) {
		startSector = p.Start
		p.Size = minSize
	})
	if err != nil {
		return 0, fmt.Errorf("error when editing partition table of %q, "+
			"input: disk=%q, partNumInt=%d, "+
			"error msg: (%v)", disk, disk, partNumInt, err)
	}

	tableBuffer.WriteString(table)

	// write partition table back.
	writeTableCmd := exec.Command("sudo", "sfdisk", "--no-reread", disk)
	writeTableCmd.Stdin = &tableBuffer
	writeTableCmd.Stdout = os.Stdout
	if err := writeTableCmd.Run(); err != nil {
		return 0, fmt.Errorf("error in writing partition table back to %q, "+
			"input: disk=%q, partNumInt=%d, "+
			"error msg: (%v)", disk, disk, partNumInt, err)
	}

	log.Printf("\nCompleted minimizing %q\n\n", partName)
	return startSector + minSize, nil
}
