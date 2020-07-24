package tools

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// SealOEMPartition sets the hashtree of the OEM partition
// with "veritysetup" and modifies the kernel command line to
// verify the OEM partition at boot time.
func SealOEMPartition(oemFSSize4K uint64) error {
	const veritysetupImgPath = "./veritysetup.img"
	const devName = "oemroot"
	if err := loadVeritysetupImage(veritysetupImgPath); err != nil {
		return fmt.Errorf("cannot load veritysetup image at %q, error msg:(%v)", veritysetupImgPath, err)
	}
	log.Println("docker image for veritysetup loaded.")
	hash, salt, err := veritysetup(oemFSSize4K)
	if err != nil {
		return fmt.Errorf("cannot run veritysetup, input:oemFSSize4K=%d, "+
			"error msg:(%v)", oemFSSize4K, err)
	}
	grupPath, err := mountEFIPartition()
	log.Println("EFI parititon mounted.")
	if err != nil {
		return fmt.Errorf("cannot mount EFI partition (/dev/sda12), error msg:(%v)", err)
	}
	partUUID, err := getPartUUID("/dev/sda8")
	if err != nil {
		return fmt.Errorf("cannot read partUUID of /dev/sda8")
	}
	if err := appendDMEntryToGRUB(grupPath, devName, partUUID, hash, salt, oemFSSize4K); err != nil {
		return fmt.Errorf("error in appending entry to grub.cfg, input:oemFSSize4K=%d, "+
			"error msg:(%v)", oemFSSize4K, err)
	}
	log.Println("kernel command line modified.")
	return nil
}

// loadVeritysetupImage loads the docker image of veritysetup
func loadVeritysetupImage(imgPath string) error {
	cmd := exec.Command("sudo", "docker", "load", "-i", imgPath)
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error in loading docker image, "+
			"input: imgPath=%q, error msg: (%v)", imgPath, err)
	}
	return nil
}

// mountEFIPartition mounts the EFI partition (/dev/sda12)
// and returns the path where grub.cfg is at.
func mountEFIPartition() (string, error) {
	var tmpDirBuf bytes.Buffer
	cmd := exec.Command("mktemp", "-d")
	cmd.Stdout = &tmpDirBuf
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("error in creating tmp directory, "+
			"error msg: (%v)", err)
	}
	dir := tmpDirBuf.String()
	dir = dir[:len(dir)-1]
	cmd = exec.Command("sudo", "mount", "/dev/sda12", dir)
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("error in mounting /dev/sda12 at %q, "+
			"error msg: (%v)", dir, err)
	}
	return dir + "/efi/boot", nil
}

// veritysetup runs the docker container command veritysetup to build hash tree of OEM partition
// and generate hash root value and salt value.
func veritysetup(oemFSSize4K uint64) (string, string, error) {
	dataBlocks := "--data-blocks=" + strconv.FormatUint(oemFSSize4K, 10)
	// --hash-offset is in Bytes
	hashOffset := "--hash-offset=" + strconv.FormatUint(oemFSSize4K<<12, 10)
	cmd := exec.Command("sudo", "docker", "run", "--privileged", "-v", "/dev:/dev", "veritysetup", "veritysetup",
		"format", "/dev/sda8", "/dev/sda8", "--data-block-size=4096", "--hash-block-size=4096", dataBlocks,
		hashOffset, "--no-superblock", "--format=0")
	var verityBuf bytes.Buffer
	cmd.Stdout = &verityBuf
	if err := cmd.Run(); err != nil {
		return "", "", fmt.Errorf("error in running docker veritysetup, "+
			"input: oemFSSize4K=%d, error msg: (%v)", oemFSSize4K, err)
	}
	// Output of veritysetup is like:
	// VERITY header information for /dev/sdb1
	// UUID:
	// Hash type:              0
	// Data blocks:            2048
	// Data block size:        4096
	// Hash block size:        4096
	// Hash algorithm:         sha256
	// Salt:                   9cd7ba29a1771b2097a7d72be8c13b29766d7617c3b924eb0cf23ff5071fee47
	// Root hash:              d6b862d01e01e6417a1b5e7eb0eed2a2189594b74325dd0749cd83bbf78f5dc8
	lines := strings.Split(verityBuf.String(), "\n")
	if !strings.HasPrefix(lines[len(lines)-2], "Root hash:") || !strings.HasPrefix(lines[len(lines)-3], "Salt:") {
		return "", "", fmt.Errorf("error in veritsetup output format, the last two lines are not \"Salt:\" and \"Root hash:\", "+
			"input: oemFSSize4K=%d, veritysetup output: %s", oemFSSize4K, verityBuf.String())
	}
	hash := strings.TrimSpace(strings.Split(lines[len(lines)-2], ":")[1])
	salt := strings.TrimSpace(strings.Split(lines[len(lines)-3], ":")[1])
	return hash, salt, nil
}

// getPartUUID finds the PartUUID of a partition using blkid
func getPartUUID(partName string) (string, error) {
	var idBuf bytes.Buffer
	cmd := exec.Command("sudo", "blkid")
	cmd.Stdout = &idBuf
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("error in running blkid, "+
			"error msg: (%v)", err)
	}
	// blkid has output like:
	// /dev/sda1: LABEL="STATE" UUID="120991ff-4f12-43bf-b962-17325185121d" TYPE="ext4"
	// /dev/sda3: LABEL="ROOT-A" SEC_TYPE="ext2" TYPE="ext4" PARTLABEL="ROOT-A" PARTUUID="00ce255b-db42-1e47-a62b-735c7a9a7397"
	// /dev/sda8: LABEL="OEM" UUID="1401457b-449d-4755-9a1e-57054b287489" TYPE="ext4" PARTLABEL="OEM" PARTUUID="9db2ae75-98dc-5b4f-a38b-b3cb0b80b17f"
	// /dev/sda12: SEC_TYPE="msdos" LABEL="EFI-SYSTEM" UUID="F6E7-003C" TYPE="vfat" PARTLABEL="EFI-SYSTEM" PARTUUID="aaea6e5e-bc5f-2542-b19a-66c2daa4d5a8"
	// /dev/dm-0: LABEL="ROOT-A" SEC_TYPE="ext2" TYPE="ext4"
	// /dev/sda2: PARTLABEL="KERN-A" PARTUUID="de4778dd-c187-8343-b86c-e122f9d234c0"
	// /dev/sda4: PARTLABEL="KERN-B" PARTUUID="7b8374db-78b2-2748-bab9-a52d0867455b"
	// /dev/sda5: PARTLABEL="ROOT-B" PARTUUID="8ac60384-1187-9e49-91ce-3abd8da295a7"
	// /dev/sda11: PARTLABEL="RWFW" PARTUUID="682ef1a5-f7f6-7d42-a407-5d8ad0430fc1"
	lines := strings.Split(idBuf.String(), "\n")
	for _, line := range lines {
		if !strings.HasPrefix(line, partName) {
			continue
		}
		for _, content := range strings.Split(line, " ") {
			if !strings.HasPrefix(content, "PARTUUID") {
				continue
			}
			return strings.Trim(strings.Split(content, "=")[1], "\""), nil
		}
	}
	return "", fmt.Errorf("partition UUID not found, input: partName=%q ,"+
		"output of \"blkid\": %s", partName, idBuf.String())
}

// appendDMEntryToGRUB appends an dm-verity table entry to kernel command line in grub.cfg
// A target line in grub.cfg looks like
// ...... root=/dev/dm-0 dm="1 vroot none ro 1,0 4077568 verity payload=PARTUUID=8AC60384-1187-9E49-91CE-3ABD8DA295A7 hashtree=PARTUUID=8AC60384-1187-9E49-91CE-3ABD8DA295A7 hashstart=4077568 alg=sha256 root_hexdigest=xxxxxxxx salt=xxxxxxxx"
func appendDMEntryToGRUB(grubPath, name, partUUID, hash, salt string, oemFSSize4K uint64) error {
	grubPath = grubPath + "/grub.cfg"
	// from 4K blocks to 512B sectors
	oemFSSizeSector := oemFSSize4K << 3
	entryString := fmt.Sprintf("%s none ro 1, 0 %d verity payload=PARTUUID=%s hashtree=PARTUUID=%s hashstart=%d alg=sha256 "+
		"root_hexdigest=%s salt=%s\"", name, oemFSSizeSector, partUUID, partUUID, oemFSSizeSector, hash, salt)
	grubContent, err := ioutil.ReadFile(grubPath)
	if err != nil {
		return fmt.Errorf("cannot read grub.cfg at %q, "+
			"input: grubPath=%q, name=%q, partUUID=%q, oemFSSize4K=%d, hash=%q, salt=%q, "+
			"error msg:(%v)", grubPath, grubPath, name, partUUID, oemFSSize4K, hash, salt, err)
	}
	lines := strings.Split(string(grubContent), "\n")
	// add the entry to all kernel command lines containing "dm="
	for idx, line := range lines {
		if !strings.Contains(line, "dm=") {
			continue
		}
		startPos := strings.Index(line, "dm=")
		lineBuf := []rune(line[:len(line)-1])
		// add number of entries.
		lineBuf[startPos+4] = '2'
		lines[idx] = strings.Join(append(strings.Split(string(lineBuf), ","), entryString), ",")
	}
	// new content of grub.cfg
	grubContent = []byte(strings.Join(lines, "\n"))
	err = ioutil.WriteFile(grubPath, grubContent, 0755)
	if err != nil {
		return fmt.Errorf("cannot write to grub.cfg at %q, "+
			"input: grubPath=%q, name=%q, partUUID=%q, oemFSSize4K=%d, hash=%q, salt=%q, "+
			"error msg:(%v)", grubPath, grubPath, name, partUUID, oemFSSize4K, hash, salt, err)
	}
	return nil

}
