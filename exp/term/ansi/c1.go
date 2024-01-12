package ansi

// C1 is a control character ranges from (0x80-0x9F) as defined in ISO 6429 (ECMA-48).
// See: https://en.wikipedia.org/wiki/C0_and_C1_control_codes
type C1 = byte

// C1 control characters.
const (
	// PAD is the padding character.
	PAD C1 = 0x80
	// HOP is the high octet preset character.
	HOP C1 = 0x81
	// BPH is the break permitted here character.
	BPH C1 = 0x82
	// NBH is the no break here character.
	NBH C1 = 0x83
	// IND is the index character.
	IND C1 = 0x84
	// NEL is the next line character.
	NEL C1 = 0x85
	// SSA is the start of selected area character.
	SSA C1 = 0x86
	// ESA is the end of selected area character.
	ESA C1 = 0x87
	// HTS is the horizontal tab set character.
	HTS C1 = 0x88
	// HTJ is the horizontal tab with justification character.
	HTJ C1 = 0x89
	// VTS is the vertical tab set character.
	VTS C1 = 0x8A
	// PLD is the partial line forward character.
	PLD C1 = 0x8B
	// PLU is the partial line backward character.
	PLU C1 = 0x8C
	// RI is the reverse index character.
	RI C1 = 0x8D
	// SS2 is the single shift 2 character.
	SS2 C1 = 0x8E
	// SS3 is the single shift 3 character.
	SS3 C1 = 0x8F
	// DCS is the device control string character.
	DCS C1 = 0x90
	// PU1 is the private use 1 character.
	PU1 C1 = 0x91
	// PU2 is the private use 2 character.
	PU2 C1 = 0x92
	// STS is the set transmit state character.
	STS C1 = 0x93
	// CCH is the cancel character.
	CCH C1 = 0x94
	// MW is the message waiting character.
	MW C1 = 0x95
	// SPA is the start of guarded area character.
	SPA C1 = 0x96
	// EPA is the end of guarded area character.
	EPA C1 = 0x97
	// SOS is the start of string character.
	SOS C1 = 0x98
	// SGCI is the single graphic character introducer character.
	SGCI C1 = 0x99
	// SCI is the single character introducer character.
	SCI C1 = 0x9A
	// CSI is the control sequence introducer character.
	CSI C1 = 0x9B
	// ST is the string terminator character.
	ST C1 = 0x9C
	// OSC is the operating system command character.
	OSC C1 = 0x9D
	// PM is the privacy message character.
	PM C1 = 0x9E
	// APC is the application program command character.
	APC C1 = 0x9F
)
