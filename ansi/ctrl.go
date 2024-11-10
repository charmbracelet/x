package ansi

// ReportNameVersion (XTVERSION) is a control sequence that requests the
// terminal's name and version. It responds with a DSR sequence identifying the
// terminal.
//
//	CSI > 0 q
//	DCS > | text ST
//
// See https://invisible-island.net/xterm/ctlseqs/ctlseqs.html#h3-PC-Style-Function-Keys
const (
	ReportNameVersion = "\x1b[>0q"
	XTVERSION         = ReportNameVersion
)

// RequestXTVersion is a control sequence that requests the terminal's XTVERSION. It responds with a DSR sequence identifying the version.
//
//	CSI > Ps q
//	DCS > | text ST
//
// See https://invisible-island.net/xterm/ctlseqs/ctlseqs.html#h3-PC-Style-Function-Keys
// Deprecated: use [ReportNameVersion] instead.
const RequestXTVersion = "\x1b[>0q"

// PrimaryDeviceAttributes (DA1) is a control sequence that reports the
// terminal's primary device attributes.
//
//	CSI c
//	CSI 0 c
//
// See https://vt100.net/docs/vt510-rm/DA1.html
const (
	PrimaryDeviceAttributes = "\x1b[c"
	DA1                     = PrimaryDeviceAttributes
)

// RequestPrimaryDeviceAttributes is a control sequence that requests the
// terminal's primary device attributes (DA1).
//
//	CSI c
//
// See https://vt100.net/docs/vt510-rm/DA1.html
// Deprecated: use [PrimaryDeviceAttributes] instead.
const RequestPrimaryDeviceAttributes = "\x1b[c"
