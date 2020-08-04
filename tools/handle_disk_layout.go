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

package tools

import (
	"cos-customizer/tools/partutil"
	"fmt"
	"log"
	"strconv"
)

// HandleDiskLayout changes the partitions on a COS disk.
// If the auto-update is disabled, it will shrink sda3 to reclaim the space.
// It also moves the OEM partition and the stateful partition by a distance relative
// to a start point. If sda3 is shrinked, the start point is the end of shrinked sda3.
// Otherwise, start point will be the original start of the stateful partition.
// The stateful partition will be moved to leave enough space for the OEM partition,
// and the OEM partition will be moved to the start point.
// Finally OEM partition will be resized to 1 sector before the new stateful partition.
// OEMSize can be the number of sectors (without unit) or size like "3G", "100M", "10000K" or "99999B"
// If no need to extend the OEM partition, oemSize=0.
func HandleDiskLayout(disk string, statePartNum, oemPartNum int, oemSize string, reclaimSDA3 bool) error {
	if len(disk) <= 0 || statePartNum <= 0 || oemPartNum <= 0 || len(oemSize) <= 0 {
		return fmt.Errorf("empty or non-positive input: disk=%q, statePartNum=%d, oemPartNum=%d, oemSize=%q",
			disk, statePartNum, oemPartNum, oemSize)
	}

	// print the old partition table.
	table, err := partutil.ReadPartitionTable(disk)
	if err != nil {
		return fmt.Errorf("cannot read old partition table of %q, "+
			"input: disk=%q, statePartNum=%d, oemPartNum=%d, oemSize=%q, reclaimSDA3=%t, "+
			"error msg: (%v)", disk, disk, statePartNum, oemPartNum, oemSize, reclaimSDA3, err)
	}
	log.Printf("\nOld partition table:\n%s\n", table)

	// read new size of OEM partition.
	newOEMSizeBytes, err := partutil.ConvertSizeToBytes(oemSize)
	if err != nil {
		return fmt.Errorf("error in reading new OEM size, "+
			"input: disk=%q, statePartNum=%d, oemPartNum=%d, oemSize=%q, reclaimSDA3=%t, "+
			"error msg: (%v)", disk, statePartNum, oemPartNum, oemSize, reclaimSDA3, err)
	}

	// read original size of OEM partition.
	oldOEMSize, err := partutil.ReadPartitionSize(disk, oemPartNum)
	if err != nil {
		return fmt.Errorf("error in reading old OEM size, "+
			"input: disk=%q, statePartNum=%d, oemPartNum=%d, oemSize=%q, reclaimSDA3=%t, "+
			"error msg: (%v)", disk, statePartNum, oemPartNum, oemSize, reclaimSDA3, err)
	}
	oldOEMSizeBytes := oldOEMSize << 8 // change unit to bytes.
	var startPointSector uint64

	if reclaimSDA3 {
		// start point is the end of shrinked sda3 +1
		startPointSector, err = partutil.MinimizePartition("/dev/sda3")
		if err != nil {
			return fmt.Errorf("error in reclaiming sda3, "+
				"input: disk=%q, statePartNum=%d, oemPartNum=%d, oemSize=%q, reclaimSDA3=%t, "+
				"error msg: (%v)", disk, statePartNum, oemPartNum, oemSize, reclaimSDA3, err)
		}
		log.Println("Shrinked /dev/sda3.")
		startPointSector++
	} else {
		// start point is the original start sector of the stateful partition.
		startPointSector, err = partutil.ReadPartitionStart(disk, statePartNum)
		if err != nil {
			return fmt.Errorf("cannot read old stateful partition start, "+
				"input: disk=%q, statePartNum=%d, oemPartNum=%d, oemSize=%q, reclaimSDA3=%t, "+
				"error msg: (%v)", disk, statePartNum, oemPartNum, oemSize, reclaimSDA3, err)
		}
	}

	// No need to resize the OEM partition.
	if newOEMSizeBytes <= oldOEMSizeBytes {
		if newOEMSizeBytes != 0 {
			log.Printf("\n!!!!!!!WARNING!!!!!!!\n"+
				"oemSize: %d bytes is not larger than the original OEM partition size: %d bytes, "+
				"nothing is done for the OEM partition.\n "+
				"input: disk=%q, statePartNum=%d, oemPartNum=%d, oemSize=%q, reclaimSDA3=%t",
				newOEMSizeBytes, oldOEMSizeBytes, disk, statePartNum, oemPartNum, oemSize, reclaimSDA3)
		}
		if !reclaimSDA3 {
			return nil
		}
		// move the stateful partition to the start point.
		if err := partutil.MovePartition(disk, statePartNum, strconv.FormatUint(startPointSector, 10)); err != nil {
			return fmt.Errorf("error in moving stateful partition, "+
				"input: disk=%q, statePartNum=%d, oemPartNum=%d, oemSize=%q, reclaimSDA3=%t, "+
				"error msg: (%v)", disk, statePartNum, oemPartNum, oemSize, reclaimSDA3, err)
		}
		log.Println("Reclaimed /dev/sda3.")
		// print the new partition table.
		table, err = partutil.ReadPartitionTable(disk)
		if err != nil {
			return fmt.Errorf("cannot read new partition table of %q, "+
				"input: disk=%q, statePartNum=%d, oemPartNum=%d, oemSize=%q, reclaimSDA3=%t, "+
				"error msg: (%v)", disk, disk, statePartNum, oemPartNum, oemSize, reclaimSDA3, err)
		}
		log.Printf("New partition table:\n%s\n", table)
		return nil
	}

	// leave enough space before the stateful partition for the OEM partition.
	newStateStartSector := startPointSector + (newOEMSizeBytes >> 8)

	// move the stateful partition.
	if err := partutil.MovePartition(disk, statePartNum, strconv.FormatUint(newStateStartSector, 10)); err != nil {
		return fmt.Errorf("error in moving stateful partition, "+
			"input: disk=%q, statePartNum=%d, oemPartNum=%d, oemSize=%q, reclaimSDA3=%t, "+
			"error msg: (%v)", disk, statePartNum, oemPartNum, oemSize, reclaimSDA3, err)
	}

	// move OEM partition to the start point.
	if err := partutil.MovePartition(disk, oemPartNum, strconv.FormatUint(startPointSector, 10)); err != nil {
		return fmt.Errorf("error in moving OEM partition, "+
			"input: disk=%q, statePartNum=%d, oemPartNum=%d, oemSize=%q, reclaimSDA3=%t, "+
			"error msg: (%v)", disk, statePartNum, oemPartNum, oemSize, reclaimSDA3, err)
	}
	log.Println("Reclaimed /dev/sda3.")

	// extend the OEM partition.
	if err = partutil.ExtendPartition(disk, oemPartNum, newStateStartSector-1); err != nil {
		return fmt.Errorf("error in extending OEM partition, "+
			"input: disk=%q, statePartNum=%d, oemPartNum=%d, oemSize=%q, reclaimSDA3=%t, "+
			"error msg: (%v)", disk, statePartNum, oemPartNum, oemSize, reclaimSDA3, err)
	}

	// print the new partition table.
	table, err = partutil.ReadPartitionTable(disk)
	if err != nil {
		return fmt.Errorf("cannot read new partition table of %q, "+
			"input: disk=%q, statePartNum=%d, oemPartNum=%d, oemSize=%q, reclaimSDA3=%t, "+
			"error msg: (%v)", disk, disk, statePartNum, oemPartNum, oemSize, reclaimSDA3, err)
	}
	log.Printf("\nCompleted extending OEM partition\n\n New partition table:\n%s\n", table)
	return nil
}
