// Package fdisk implements the fdisk applet in list mode: read and print the MBR
// partition table of a device or image. Interactive editing is not provided.
package fdisk

import (
	"context"
	"encoding/binary"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the fdisk applet.
type Command struct{}

// New returns a fdisk command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "fdisk" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "List the MBR partition table" }

const (
	sectorSize   = 512
	tableOffset  = 446 // 0x1BE
	entrySize    = 16
	numEntries   = 4
	signatureLow = 510
	bootableFlag = 0x80
)

// partTypes names the common MBR partition type bytes.
var partTypes = map[byte]string{
	0x00: "Empty", 0x01: "FAT12", 0x04: "FAT16 <32M", 0x05: "Extended",
	0x06: "FAT16", 0x07: "HPFS/NTFS/exFAT", 0x0b: "W95 FAT32", 0x0c: "W95 FAT32 (LBA)",
	0x0e: "W95 FAT16 (LBA)", 0x82: "Linux swap", 0x83: "Linux", 0x8e: "Linux LVM",
	0xfd: "Linux raid autodetect", 0xef: "EFI (FAT-12/16/32)", 0xee: "GPT protective",
}

// Run executes fdisk.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "-l DEVICE", stdio.Err).WithHelp(command.Help{
		Description: "List the MBR (DOS) partition table of DEVICE (a block device or image file): each " +
			"partition's boot flag, start and end sector, sector count, and type. Only the list mode " +
			"(-l) is implemented; the interactive partition editor is not.",
		Examples: []command.Example{
			{Command: "fdisk -l disk.img", Explain: "List the partitions in the image."},
		},
		ExitStatus: "0  the table was listed.\n1  the device was unreadable or had no valid MBR.",
	})
	list := fs.BoolP("list", "l", false, "list the partition table and exit")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	rest := fs.Args()
	if len(rest) == 0 {
		return command.Failuref("a device or image is required")
	}
	if !*list {
		_, _ = fmt.Fprintln(stdio.Err, "fdisk: only the list mode (-l) is supported by this build")
		return command.SilentFailure()
	}
	device := rest[0]

	mbr, err := readSector(device)
	if err != nil {
		return command.Failuref("%s: %v", device, err)
	}
	if mbr[signatureLow] != 0x55 || mbr[signatureLow+1] != 0xAA {
		return command.Failuref("%s: no valid MBR signature", device)
	}

	tw := tabwriter.NewWriter(stdio.Out, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(tw, "Device\tBoot\tStart\tEnd\tSectors\tType")
	for i := 0; i < numEntries; i++ {
		e := mbr[tableOffset+i*entrySize:]
		typeByte := e[4]
		if typeByte == 0 {
			continue // unused entry
		}
		start := binary.LittleEndian.Uint32(e[8:])
		sectors := binary.LittleEndian.Uint32(e[12:])
		boot := ""
		if e[0] == bootableFlag {
			boot = "*"
		}
		end := start + sectors - 1
		_, _ = fmt.Fprintf(tw, "%s%d\t%s\t%d\t%d\t%d\t%s\n",
			device, i+1, boot, start, end, sectors, typeName(typeByte))
	}
	_ = tw.Flush()
	return nil
}

func typeName(b byte) string {
	if name, ok := partTypes[b]; ok {
		return name
	}
	return fmt.Sprintf("unknown (0x%02x)", b)
}

// readSector reads the first 512-byte sector of path.
func readSector(path string) ([]byte, error) {
	f, err := os.Open(path) //nolint:gosec // user-named device or image
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	buf := make([]byte, sectorSize)
	if _, err := f.ReadAt(buf, 0); err != nil {
		return nil, fmt.Errorf("cannot read the MBR: %w", err)
	}
	return buf, nil
}
