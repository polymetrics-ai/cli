// Command windowsversion renders deterministic Windows VERSIONINFO resources for pm.exe.
//
// The generated .rc file is source text for review, while -goarch emits an architecture-specific
// .syso file that Go can link into Windows snapshots. Generated .syso files are build artifacts and
// must not be committed.
package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"flag"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

const (
	companyName      = "Polymetrics AI"
	fileDescription  = "Polymetrics CLI"
	internalName     = "pm"
	originalFilename = "pm.exe"
	productName      = "Polymetrics CLI"
	copyright        = "Copyright (c) 2026 Polymetrics AI"
)

const (
	coffMachineAMD64 = 0x8664
	coffMachineARM64 = 0xaa64

	coffRelAMD64Addr32NB = 0x0003
	coffRelARM64Addr32NB = 0x0002

	rtVersion       = 16
	defaultResource = 1
	langEnglishUS   = 0x0409
	codepageUnicode = 1200

	sectionInitializedData = 0x00000040
	sectionMemoryRead      = 0x40000000
)

var semverPattern = regexp.MustCompile(`^v?([0-9]+)\.([0-9]+)\.([0-9]+)(?:[-+].*)?$`)

// Version is the four-part numeric Windows file/product version.
type Version struct {
	Major int
	Minor int
	Patch int
	Build int
}

// String returns the dotted four-part Windows version string.
func (v Version) String() string {
	return fmt.Sprintf("%d.%d.%d.%d", v.Major, v.Minor, v.Patch, v.Build)
}

// CommaString returns the comma-separated numeric form required by VERSIONINFO FILEVERSION fields.
func (v Version) CommaString() string {
	return fmt.Sprintf("%d,%d,%d,%d", v.Major, v.Minor, v.Patch, v.Build)
}

// NormalizeVersion converts a PM release version into a four-part Windows VERSIONINFO version.
func NormalizeVersion(input string) (Version, error) {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return Version{}, errors.New("version is required")
	}

	matches := semverPattern.FindStringSubmatch(trimmed)
	if matches == nil {
		return Version{}, fmt.Errorf("version %q must look like vMAJOR.MINOR.PATCH", input)
	}

	parts := make([]int, 3)
	for i := range parts {
		value, err := strconv.Atoi(matches[i+1])
		if err != nil {
			return Version{}, fmt.Errorf("parse version component %q: %w", matches[i+1], err)
		}
		if value < 0 || value > 65535 {
			return Version{}, fmt.Errorf("version component %d out of Windows VERSIONINFO range 0..65535", value)
		}
		parts[i] = value
	}

	return Version{Major: parts[0], Minor: parts[1], Patch: parts[2], Build: 0}, nil
}

// RenderRC renders the Windows resource script consumed by rc.exe.
func RenderRC(version Version) string {
	commaVersion := version.CommaString()
	textVersion := version.String()

	return fmt.Sprintf(`1 VERSIONINFO
FILEVERSION %s
PRODUCTVERSION %s
FILEFLAGSMASK 0x3fL
FILEFLAGS 0x0L
FILEOS 0x40004L
FILETYPE 0x1L
FILESUBTYPE 0x0L
BEGIN
    BLOCK "StringFileInfo"
    BEGIN
        BLOCK "040904B0"
        BEGIN
            VALUE "CompanyName", "%s"
            VALUE "FileDescription", "%s"
            VALUE "FileVersion", "%s"
            VALUE "InternalName", "%s"
            VALUE "OriginalFilename", "%s"
            VALUE "ProductName", "%s"
            VALUE "ProductVersion", "%s"
            VALUE "LegalCopyright", "%s"
        END
    END
    BLOCK "VarFileInfo"
    BEGIN
        VALUE "Translation", 0x0409, 1200
    END
END
`, commaVersion, commaVersion, companyName, fileDescription, textVersion, internalName, originalFilename, productName, textVersion, copyright)
}

// RenderSyso renders a linkable PE/COFF .rsrc object containing the VERSIONINFO resource.
func RenderSyso(version Version, goarch string) ([]byte, error) {
	machine, relocType, err := coffTarget(goarch)
	if err != nil {
		return nil, err
	}

	resource, relocationOffset := renderResourceSection(version)
	return renderCOFFResourceObject(machine, relocType, resource, relocationOffset), nil
}

func coffTarget(goarch string) (machine uint16, relocType uint16, err error) {
	switch goarch {
	case "amd64":
		return coffMachineAMD64, coffRelAMD64Addr32NB, nil
	case "arm64":
		return coffMachineARM64, coffRelARM64Addr32NB, nil
	default:
		return 0, 0, fmt.Errorf("unsupported goarch %q; want amd64 or arm64", goarch)
	}
}

func renderResourceSection(version Version) ([]byte, uint32) {
	versionInfo := renderVersionInfo(version)

	var section bytes.Buffer
	writeResourceDirectory(&section, 1)
	writeResourceDirectoryEntry(&section, rtVersion, 24, true)
	writeResourceDirectory(&section, 1)
	writeResourceDirectoryEntry(&section, defaultResource, 48, true)
	writeResourceDirectory(&section, 1)
	writeResourceDirectoryEntry(&section, langEnglishUS, 72, false)

	dataEntryOffset := uint32(section.Len())
	writeUint32(&section, 0) // OffsetToData, patched after the resource value offset is known.
	writeUint32(&section, uint32(len(versionInfo)))
	writeUint32(&section, codepageUnicode)
	writeUint32(&section, 0)
	pad4(&section)

	versionInfoOffset := uint32(section.Len())
	section.Write(versionInfo)
	resource := section.Bytes()
	binary.LittleEndian.PutUint32(resource[dataEntryOffset:], versionInfoOffset)

	return resource, dataEntryOffset
}

func writeResourceDirectory(buf *bytes.Buffer, idEntries uint16) {
	writeUint32(buf, 0) // Characteristics
	writeUint32(buf, 0) // TimeDateStamp
	writeUint16(buf, 0) // MajorVersion
	writeUint16(buf, 0) // MinorVersion
	writeUint16(buf, 0) // NumberOfNamedEntries
	writeUint16(buf, idEntries)
}

func writeResourceDirectoryEntry(buf *bytes.Buffer, id uint32, offset uint32, directory bool) {
	writeUint32(buf, id)
	if directory {
		offset |= 0x80000000
	}
	writeUint32(buf, offset)
}

func renderVersionInfo(version Version) []byte {
	var buf bytes.Buffer
	start := beginVersionBlock(&buf, "VS_VERSION_INFO", 52, 0)
	writeFixedFileInfo(&buf, version)
	pad4(&buf)
	writeStringFileInfo(&buf, version)
	writeVarFileInfo(&buf)
	endVersionBlock(&buf, start)
	return buf.Bytes()
}

func writeFixedFileInfo(buf *bytes.Buffer, version Version) {
	versionMS := uint32(version.Major)<<16 | uint32(version.Minor)
	versionLS := uint32(version.Patch)<<16 | uint32(version.Build)

	for _, value := range []uint32{
		0xFEEF04BD, // dwSignature
		0x00010000, // dwStrucVersion
		versionMS,
		versionLS,
		versionMS,
		versionLS,
		0x0000003F, // dwFileFlagsMask
		0x00000000, // dwFileFlags
		0x00040004, // dwFileOS: VOS_NT_WINDOWS32
		0x00000001, // dwFileType: VFT_APP
		0x00000000, // dwFileSubtype
		0x00000000, // dwFileDateMS
		0x00000000, // dwFileDateLS
	} {
		writeUint32(buf, value)
	}
}

func writeStringFileInfo(buf *bytes.Buffer, version Version) {
	start := beginVersionBlock(buf, "StringFileInfo", 0, 1)
	tableStart := beginVersionBlock(buf, "040904B0", 0, 1)
	for _, item := range []struct {
		key   string
		value string
	}{
		{key: "CompanyName", value: companyName},
		{key: "FileDescription", value: fileDescription},
		{key: "FileVersion", value: version.String()},
		{key: "InternalName", value: internalName},
		{key: "OriginalFilename", value: originalFilename},
		{key: "ProductName", value: productName},
		{key: "ProductVersion", value: version.String()},
		{key: "LegalCopyright", value: copyright},
	} {
		writeStringValue(buf, item.key, item.value)
	}
	endVersionBlock(buf, tableStart)
	endVersionBlock(buf, start)
}

func writeStringValue(buf *bytes.Buffer, key, value string) {
	encodedValue := utf16Bytes(value)
	start := beginVersionBlock(buf, key, uint16(len([]rune(value))+1), 1)
	buf.Write(encodedValue)
	pad4(buf)
	endVersionBlock(buf, start)
}

func writeVarFileInfo(buf *bytes.Buffer) {
	start := beginVersionBlock(buf, "VarFileInfo", 0, 1)
	translationStart := beginVersionBlock(buf, "Translation", 4, 0)
	writeUint16(buf, langEnglishUS)
	writeUint16(buf, codepageUnicode)
	pad4(buf)
	endVersionBlock(buf, translationStart)
	endVersionBlock(buf, start)
}

func beginVersionBlock(buf *bytes.Buffer, key string, valueLength uint16, valueType uint16) int {
	start := buf.Len()
	writeUint16(buf, 0)
	writeUint16(buf, valueLength)
	writeUint16(buf, valueType)
	buf.Write(utf16Bytes(key))
	pad4(buf)
	return start
}

func endVersionBlock(buf *bytes.Buffer, start int) {
	binary.LittleEndian.PutUint16(buf.Bytes()[start:], uint16(buf.Len()-start))
}

func utf16Bytes(value string) []byte {
	encoded := make([]byte, 0, (len(value)+1)*2)
	for _, r := range value {
		if r > 0xFFFF {
			// PM's VERSIONINFO strings are ASCII. Keep replacement deterministic if that changes.
			r = '?'
		}
		encoded = binary.LittleEndian.AppendUint16(encoded, uint16(r))
	}
	encoded = binary.LittleEndian.AppendUint16(encoded, 0)
	return encoded
}

func renderCOFFResourceObject(machine uint16, relocType uint16, resource []byte, relocationOffset uint32) []byte {
	const (
		coffHeaderSize    = 20
		sectionHeaderSize = 40
		relocationSize    = 10
		symbolSize        = 18
	)

	rawDataOffset := uint32(coffHeaderSize + sectionHeaderSize)
	relocationTableOffset := rawDataOffset + uint32(len(resource))
	symbolTableOffset := relocationTableOffset + relocationSize

	var buf bytes.Buffer
	writeUint16(&buf, machine)
	writeUint16(&buf, 1) // NumberOfSections
	writeUint32(&buf, 0) // TimeDateStamp
	writeUint32(&buf, symbolTableOffset)
	writeUint32(&buf, 1) // NumberOfSymbols
	writeUint16(&buf, 0) // SizeOfOptionalHeader
	writeUint16(&buf, 0) // Characteristics

	writeName(&buf, ".rsrc")
	writeUint32(&buf, 0) // VirtualSize
	writeUint32(&buf, 0) // VirtualAddress
	writeUint32(&buf, uint32(len(resource)))
	writeUint32(&buf, rawDataOffset)
	writeUint32(&buf, relocationTableOffset)
	writeUint32(&buf, 0) // PointerToLinenumbers
	writeUint16(&buf, 1) // NumberOfRelocations
	writeUint16(&buf, 0) // NumberOfLinenumbers
	writeUint32(&buf, sectionInitializedData|sectionMemoryRead)

	buf.Write(resource)

	writeUint32(&buf, relocationOffset)
	writeUint32(&buf, 0) // SymbolTableIndex: .rsrc section symbol
	writeUint16(&buf, relocType)

	writeName(&buf, ".rsrc")
	writeUint32(&buf, 0) // Value
	writeUint16(&buf, 1) // SectionNumber
	writeUint16(&buf, 0) // Type
	buf.WriteByte(3)     // StorageClass: IMAGE_SYM_CLASS_STATIC
	buf.WriteByte(0)     // NumberOfAuxSymbols
	writeUint32(&buf, 4) // Empty COFF string table length.

	if buf.Len() != int(symbolTableOffset)+symbolSize+4 {
		panic("internal COFF size mismatch")
	}
	return buf.Bytes()
}

func writeName(buf *bytes.Buffer, name string) {
	var raw [8]byte
	copy(raw[:], name)
	buf.Write(raw[:])
}

func pad4(buf *bytes.Buffer) {
	for buf.Len()%4 != 0 {
		buf.WriteByte(0)
	}
}

func writeUint16(buf *bytes.Buffer, value uint16) {
	var raw [2]byte
	binary.LittleEndian.PutUint16(raw[:], value)
	buf.Write(raw[:])
}

func writeUint32(buf *bytes.Buffer, value uint32) {
	var raw [4]byte
	binary.LittleEndian.PutUint32(raw[:], value)
	buf.Write(raw[:])
}

func main() {
	versionFlag := flag.String("version", "", "PM release version, for example v1.2.3")
	goarchFlag := flag.String("goarch", "", "optional Windows GOARCH for .syso output; supported values are amd64 and arm64")
	outFlag := flag.String("out", "", "output path; writes to stdout when omitted")
	flag.Parse()

	version, err := NormalizeVersion(*versionFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "windowsversion: %v\n", err)
		os.Exit(2)
	}

	var content []byte
	if *goarchFlag == "" {
		content = []byte(RenderRC(version))
	} else {
		content, err = RenderSyso(version, *goarchFlag)
		if err != nil {
			fmt.Fprintf(os.Stderr, "windowsversion: %v\n", err)
			os.Exit(2)
		}
	}
	if *outFlag == "" {
		if _, err := os.Stdout.Write(content); err != nil {
			fmt.Fprintf(os.Stderr, "windowsversion: write stdout: %v\n", err)
			os.Exit(1)
		}
		return
	}
	if err := os.WriteFile(*outFlag, content, 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "windowsversion: write %s: %v\n", *outFlag, err)
		os.Exit(1)
	}
}
