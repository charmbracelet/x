package ansi

// C0 is a control character ranges from (0x00-0x1F) as defined in ISO 646 (ASCII).
// See: https://en.wikipedia.org/wiki/C0_and_C1_control_codes
type C0 = byte

// C0 control characters.
//
// These range from (0x00-0x1F) as defined in ISO 646 (ASCII).
// See: https://en.wikipedia.org/wiki/C0_and_C1_control_codes
const (
	// NUL is the null character (Caret: ^@, Char: \0).
	NUL C0 = 0x00
	// SOH is the start of heading character (Caret: ^A).
	SOH C0 = 0x01
	// STX is the start of text character (Caret: ^B).
	STX C0 = 0x02
	// ETX is the end of text character (Caret: ^C).
	ETX C0 = 0x03
	// EOT is the end of transmission character (Caret: ^D).
	EOT C0 = 0x04
	// ENQ is the enquiry character (Caret: ^E).
	ENQ C0 = 0x05
	// ACK is the acknowledge character (Caret: ^F).
	ACK C0 = 0x06
	// BEL is the bell character (Caret: ^G, Char: \a).
	BEL C0 = 0x07
	// BS is the backspace character (Caret: ^H, Char: \b).
	BS C0 = 0x08
	// HT is the horizontal tab character (Caret: ^I, Char: \t).
	HT C0 = 0x09
	// LF is the line feed character (Caret: ^J, Char: \n).
	LF C0 = 0x0A
	// VT is the vertical tab character (Caret: ^K, Char: \v).
	VT C0 = 0x0B
	// FF is the form feed character (Caret: ^L, Char: \f).
	FF C0 = 0x0C
	// CR is the carriage return character (Caret: ^M, Char: \r).
	CR C0 = 0x0D
	// SO is the shift out character (Caret: ^N).
	SO C0 = 0x0E
	// SI is the shift in character (Caret: ^O).
	SI C0 = 0x0F
	// DLE is the data link escape character (Caret: ^P).
	DLE C0 = 0x10
	// DC1 is the device control 1 character (Caret: ^Q).
	DC1 C0 = 0x11
	// DC2 is the device control 2 character (Caret: ^R).
	DC2 C0 = 0x12
	// DC3 is the device control 3 character (Caret: ^S).
	DC3 C0 = 0x13
	// DC4 is the device control 4 character (Caret: ^T).
	DC4 C0 = 0x14
	// NAK is the negative acknowledge character (Caret: ^U).
	NAK C0 = 0x15
	// SYN is the synchronous idle character (Caret: ^V).
	SYN C0 = 0x16
	// ETB is the end of transmission block character (Caret: ^W).
	ETB C0 = 0x17
	// CAN is the cancel character (Caret: ^X).
	CAN C0 = 0x18
	// EM is the end of medium character (Caret: ^Y).
	EM C0 = 0x19
	// SUB is the substitute character (Caret: ^Z).
	SUB C0 = 0x1A
	// ESC is the escape character (Caret: ^[, Char: \e).
	ESC C0 = 0x1B
	// FS is the file separator character (Caret: ^\).
	FS C0 = 0x1C
	// GS is the group separator character (Caret: ^]).
	GS C0 = 0x1D
	// RS is the record separator character (Caret: ^^).
	RS C0 = 0x1E
	// US is the unit separator character (Caret: ^_).
	US C0 = 0x1F
)
