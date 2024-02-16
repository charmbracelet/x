package ansi

import (
	"strings"

	"github.com/charmbracelet/x/exp/term/input"
	"github.com/xo/terminfo"
)

func (d *driver) registerTerminfoKeys() {
	if d.term == "" {
		return
	}

	ti, _ := terminfo.Load(d.term)
	if ti == nil {
		return
	}

	tiTable := defaultTerminfoKeys(d.flags)

	// Default keys
	for name, seq := range ti.StringCapsShort() {
		if !strings.HasPrefix(name, "k") || len(seq) == 0 {
			continue
		}

		if k, ok := tiTable[name]; ok {
			d.table[string(seq)] = k
		}
	}

	// Extended keys
	for name, seq := range ti.ExtStringCapsShort() {
		if !strings.HasPrefix(name, "k") || len(seq) == 0 {
			continue
		}

		if k, ok := tiTable[name]; ok {
			d.table[string(seq)] = k
		}
	}
}

// This returns a map of terminfo keys to key events. It's a mix of ncurses
// terminfo default and user-defined key capabilities.
// Upper-case caps that are defined in the default terminfo database are
//   - kNXT
//   - kPRV
//   - kHOM
//   - kEND
//   - kDC
//   - kIC
//   - kLFT
//   - kRIT
//
// See https://man7.org/linux/man-pages/man5/terminfo.5.html
// See https://github.com/mirror/ncurses/blob/master/include/Caps-ncurses
func defaultTerminfoKeys(flags int) map[string]input.KeyEvent {
	keys := map[string]input.KeyEvent{
		"kcuu1": {Sym: input.KeyUp},
		"kUP":   {Sym: input.KeyUp, Mod: input.Shift},
		"kUP3":  {Sym: input.KeyUp, Mod: input.Alt},
		"kUP4":  {Sym: input.KeyUp, Mod: input.Shift | input.Alt},
		"kUP5":  {Sym: input.KeyUp, Mod: input.Ctrl},
		"kUP6":  {Sym: input.KeyUp, Mod: input.Shift | input.Ctrl},
		"kUP7":  {Sym: input.KeyUp, Mod: input.Alt | input.Ctrl},
		"kUP8":  {Sym: input.KeyUp, Mod: input.Shift | input.Alt | input.Ctrl},
		"kcud1": {Sym: input.KeyDown},
		"kDN":   {Sym: input.KeyDown, Mod: input.Shift},
		"kDN3":  {Sym: input.KeyDown, Mod: input.Alt},
		"kDN4":  {Sym: input.KeyDown, Mod: input.Shift | input.Alt},
		"kDN5":  {Sym: input.KeyDown, Mod: input.Ctrl},
		"kDN7":  {Sym: input.KeyDown, Mod: input.Alt | input.Ctrl},
		"kDN6":  {Sym: input.KeyDown, Mod: input.Shift | input.Ctrl},
		"kDN8":  {Sym: input.KeyDown, Mod: input.Shift | input.Alt | input.Ctrl},
		"kcub1": {Sym: input.KeyLeft},
		"kLFT":  {Sym: input.KeyLeft, Mod: input.Shift},
		"kLFT3": {Sym: input.KeyLeft, Mod: input.Alt},
		"kLFT4": {Sym: input.KeyLeft, Mod: input.Shift | input.Alt},
		"kLFT5": {Sym: input.KeyLeft, Mod: input.Ctrl},
		"kLFT6": {Sym: input.KeyLeft, Mod: input.Shift | input.Ctrl},
		"kLFT7": {Sym: input.KeyLeft, Mod: input.Alt | input.Ctrl},
		"kLFT8": {Sym: input.KeyLeft, Mod: input.Shift | input.Alt | input.Ctrl},
		"kcuf1": {Sym: input.KeyRight},
		"kRIT":  {Sym: input.KeyRight, Mod: input.Shift},
		"kRIT3": {Sym: input.KeyRight, Mod: input.Alt},
		"kRIT4": {Sym: input.KeyRight, Mod: input.Shift | input.Alt},
		"kRIT5": {Sym: input.KeyRight, Mod: input.Ctrl},
		"kRIT6": {Sym: input.KeyRight, Mod: input.Shift | input.Ctrl},
		"kRIT7": {Sym: input.KeyRight, Mod: input.Alt | input.Ctrl},
		"kRIT8": {Sym: input.KeyRight, Mod: input.Shift | input.Alt | input.Ctrl},
		"kich1": {Sym: input.KeyInsert},
		"kIC":   {Sym: input.KeyInsert, Mod: input.Shift},
		"kIC3":  {Sym: input.KeyInsert, Mod: input.Alt},
		"kIC4":  {Sym: input.KeyInsert, Mod: input.Shift | input.Alt},
		"kIC5":  {Sym: input.KeyInsert, Mod: input.Ctrl},
		"kIC6":  {Sym: input.KeyInsert, Mod: input.Shift | input.Ctrl},
		"kIC7":  {Sym: input.KeyInsert, Mod: input.Alt | input.Ctrl},
		"kIC8":  {Sym: input.KeyInsert, Mod: input.Shift | input.Alt | input.Ctrl},
		"kdch1": {Sym: input.KeyDelete},
		"kDC":   {Sym: input.KeyDelete, Mod: input.Shift},
		"kDC3":  {Sym: input.KeyDelete, Mod: input.Alt},
		"kDC4":  {Sym: input.KeyDelete, Mod: input.Shift | input.Alt},
		"kDC5":  {Sym: input.KeyDelete, Mod: input.Ctrl},
		"kDC6":  {Sym: input.KeyDelete, Mod: input.Shift | input.Ctrl},
		"kDC7":  {Sym: input.KeyDelete, Mod: input.Alt | input.Ctrl},
		"kDC8":  {Sym: input.KeyDelete, Mod: input.Shift | input.Alt | input.Ctrl},
		"khome": {Sym: input.KeyHome},
		"kHOM":  {Sym: input.KeyHome, Mod: input.Shift},
		"kHOM3": {Sym: input.KeyHome, Mod: input.Alt},
		"kHOM4": {Sym: input.KeyHome, Mod: input.Shift | input.Alt},
		"kHOM5": {Sym: input.KeyHome, Mod: input.Ctrl},
		"kHOM6": {Sym: input.KeyHome, Mod: input.Shift | input.Ctrl},
		"kHOM7": {Sym: input.KeyHome, Mod: input.Alt | input.Ctrl},
		"kHOM8": {Sym: input.KeyHome, Mod: input.Shift | input.Alt | input.Ctrl},
		"kend":  {Sym: input.KeyEnd},
		"kEND":  {Sym: input.KeyEnd, Mod: input.Shift},
		"kEND3": {Sym: input.KeyEnd, Mod: input.Alt},
		"kEND4": {Sym: input.KeyEnd, Mod: input.Shift | input.Alt},
		"kEND5": {Sym: input.KeyEnd, Mod: input.Ctrl},
		"kEND6": {Sym: input.KeyEnd, Mod: input.Shift | input.Ctrl},
		"kEND7": {Sym: input.KeyEnd, Mod: input.Alt | input.Ctrl},
		"kEND8": {Sym: input.KeyEnd, Mod: input.Shift | input.Alt | input.Ctrl},
		"kpp":   {Sym: input.KeyPgUp},
		"kprv":  {Sym: input.KeyPgUp},
		"kPRV":  {Sym: input.KeyPgUp, Mod: input.Shift},
		"kPRV3": {Sym: input.KeyPgUp, Mod: input.Alt},
		"kPRV4": {Sym: input.KeyPgUp, Mod: input.Shift | input.Alt},
		"kPRV5": {Sym: input.KeyPgUp, Mod: input.Ctrl},
		"kPRV6": {Sym: input.KeyPgUp, Mod: input.Shift | input.Ctrl},
		"kPRV7": {Sym: input.KeyPgUp, Mod: input.Alt | input.Ctrl},
		"kPRV8": {Sym: input.KeyPgUp, Mod: input.Shift | input.Alt | input.Ctrl},
		"knp":   {Sym: input.KeyPgDown},
		"knxt":  {Sym: input.KeyPgDown},
		"kNXT":  {Sym: input.KeyPgDown, Mod: input.Shift},
		"kNXT3": {Sym: input.KeyPgDown, Mod: input.Alt},
		"kNXT4": {Sym: input.KeyPgDown, Mod: input.Shift | input.Alt},
		"kNXT5": {Sym: input.KeyPgDown, Mod: input.Ctrl},
		"kNXT6": {Sym: input.KeyPgDown, Mod: input.Shift | input.Ctrl},
		"kNXT7": {Sym: input.KeyPgDown, Mod: input.Alt | input.Ctrl},
		"kNXT8": {Sym: input.KeyPgDown, Mod: input.Shift | input.Alt | input.Ctrl},

		"kbs":  {Sym: input.KeyBackspace},
		"kcbt": {Sym: input.KeyTab, Mod: input.Shift},

		// Function keys
		// This only includes the first 12 function keys. The rest are treated
		// as modifiers of the first 12.
		// Take a look at XTerm modifyFunctionKeys
		//
		// XXX: To use unambiguous function keys, use fixterms or kitty clipboard.
		//
		// See https://invisible-island.net/xterm/manpage/xterm.html#VT100-Widget-Resources:modifyFunctionKeys
		// See https://invisible-island.net/xterm/terminfo.html

		"kf1":  {Sym: input.KeyF1},
		"kf2":  {Sym: input.KeyF2},
		"kf3":  {Sym: input.KeyF3},
		"kf4":  {Sym: input.KeyF4},
		"kf5":  {Sym: input.KeyF5},
		"kf6":  {Sym: input.KeyF6},
		"kf7":  {Sym: input.KeyF7},
		"kf8":  {Sym: input.KeyF8},
		"kf9":  {Sym: input.KeyF9},
		"kf10": {Sym: input.KeyF10},
		"kf11": {Sym: input.KeyF11},
		"kf12": {Sym: input.KeyF12},
		"kf13": {Sym: input.KeyF1, Mod: input.Shift},
		"kf14": {Sym: input.KeyF2, Mod: input.Shift},
		"kf15": {Sym: input.KeyF3, Mod: input.Shift},
		"kf16": {Sym: input.KeyF4, Mod: input.Shift},
		"kf17": {Sym: input.KeyF5, Mod: input.Shift},
		"kf18": {Sym: input.KeyF6, Mod: input.Shift},
		"kf19": {Sym: input.KeyF7, Mod: input.Shift},
		"kf20": {Sym: input.KeyF8, Mod: input.Shift},
		"kf21": {Sym: input.KeyF9, Mod: input.Shift},
		"kf22": {Sym: input.KeyF10, Mod: input.Shift},
		"kf23": {Sym: input.KeyF11, Mod: input.Shift},
		"kf24": {Sym: input.KeyF12, Mod: input.Shift},
		"kf25": {Sym: input.KeyF1, Mod: input.Ctrl},
		"kf26": {Sym: input.KeyF2, Mod: input.Ctrl},
		"kf27": {Sym: input.KeyF3, Mod: input.Ctrl},
		"kf28": {Sym: input.KeyF4, Mod: input.Ctrl},
		"kf29": {Sym: input.KeyF5, Mod: input.Ctrl},
		"kf30": {Sym: input.KeyF6, Mod: input.Ctrl},
		"kf31": {Sym: input.KeyF7, Mod: input.Ctrl},
		"kf32": {Sym: input.KeyF8, Mod: input.Ctrl},
		"kf33": {Sym: input.KeyF9, Mod: input.Ctrl},
		"kf34": {Sym: input.KeyF10, Mod: input.Ctrl},
		"kf35": {Sym: input.KeyF11, Mod: input.Ctrl},
		"kf36": {Sym: input.KeyF12, Mod: input.Ctrl},
		"kf37": {Sym: input.KeyF1, Mod: input.Shift | input.Ctrl},
		"kf38": {Sym: input.KeyF2, Mod: input.Shift | input.Ctrl},
		"kf39": {Sym: input.KeyF3, Mod: input.Shift | input.Ctrl},
		"kf40": {Sym: input.KeyF4, Mod: input.Shift | input.Ctrl},
		"kf41": {Sym: input.KeyF5, Mod: input.Shift | input.Ctrl},
		"kf42": {Sym: input.KeyF6, Mod: input.Shift | input.Ctrl},
		"kf43": {Sym: input.KeyF7, Mod: input.Shift | input.Ctrl},
		"kf44": {Sym: input.KeyF8, Mod: input.Shift | input.Ctrl},
		"kf45": {Sym: input.KeyF9, Mod: input.Shift | input.Ctrl},
		"kf46": {Sym: input.KeyF10, Mod: input.Shift | input.Ctrl},
		"kf47": {Sym: input.KeyF11, Mod: input.Shift | input.Ctrl},
		"kf48": {Sym: input.KeyF12, Mod: input.Shift | input.Ctrl},
		"kf49": {Sym: input.KeyF1, Mod: input.Alt},
		"kf50": {Sym: input.KeyF2, Mod: input.Alt},
		"kf51": {Sym: input.KeyF3, Mod: input.Alt},
		"kf52": {Sym: input.KeyF4, Mod: input.Alt},
		"kf53": {Sym: input.KeyF5, Mod: input.Alt},
		"kf54": {Sym: input.KeyF6, Mod: input.Alt},
		"kf55": {Sym: input.KeyF7, Mod: input.Alt},
		"kf56": {Sym: input.KeyF8, Mod: input.Alt},
		"kf57": {Sym: input.KeyF9, Mod: input.Alt},
		"kf58": {Sym: input.KeyF10, Mod: input.Alt},
		"kf59": {Sym: input.KeyF11, Mod: input.Alt},
		"kf60": {Sym: input.KeyF12, Mod: input.Alt},
		"kf61": {Sym: input.KeyF1, Mod: input.Shift | input.Alt},
		"kf62": {Sym: input.KeyF2, Mod: input.Shift | input.Alt},
		"kf63": {Sym: input.KeyF3, Mod: input.Shift | input.Alt},
	}

	// Preserve F keys from F13 to F63 instead of using them for F-keys
	// modifiers.
	if flags&FlagFKeys != 0 {
		keys["kf13"] = input.KeyEvent{Sym: input.KeyF13}
		keys["kf14"] = input.KeyEvent{Sym: input.KeyF14}
		keys["kf15"] = input.KeyEvent{Sym: input.KeyF15}
		keys["kf16"] = input.KeyEvent{Sym: input.KeyF16}
		keys["kf17"] = input.KeyEvent{Sym: input.KeyF17}
		keys["kf18"] = input.KeyEvent{Sym: input.KeyF18}
		keys["kf19"] = input.KeyEvent{Sym: input.KeyF19}
		keys["kf20"] = input.KeyEvent{Sym: input.KeyF20}
		keys["kf21"] = input.KeyEvent{Sym: input.KeyF21}
		keys["kf22"] = input.KeyEvent{Sym: input.KeyF22}
		keys["kf23"] = input.KeyEvent{Sym: input.KeyF23}
		keys["kf24"] = input.KeyEvent{Sym: input.KeyF24}
		keys["kf25"] = input.KeyEvent{Sym: input.KeyF25}
		keys["kf26"] = input.KeyEvent{Sym: input.KeyF26}
		keys["kf27"] = input.KeyEvent{Sym: input.KeyF27}
		keys["kf28"] = input.KeyEvent{Sym: input.KeyF28}
		keys["kf29"] = input.KeyEvent{Sym: input.KeyF29}
		keys["kf30"] = input.KeyEvent{Sym: input.KeyF30}
		keys["kf31"] = input.KeyEvent{Sym: input.KeyF31}
		keys["kf32"] = input.KeyEvent{Sym: input.KeyF32}
		keys["kf33"] = input.KeyEvent{Sym: input.KeyF33}
		keys["kf34"] = input.KeyEvent{Sym: input.KeyF34}
		keys["kf35"] = input.KeyEvent{Sym: input.KeyF35}
		keys["kf36"] = input.KeyEvent{Sym: input.KeyF36}
		keys["kf37"] = input.KeyEvent{Sym: input.KeyF37}
		keys["kf38"] = input.KeyEvent{Sym: input.KeyF38}
		keys["kf39"] = input.KeyEvent{Sym: input.KeyF39}
		keys["kf40"] = input.KeyEvent{Sym: input.KeyF40}
		keys["kf41"] = input.KeyEvent{Sym: input.KeyF41}
		keys["kf42"] = input.KeyEvent{Sym: input.KeyF42}
		keys["kf43"] = input.KeyEvent{Sym: input.KeyF43}
		keys["kf44"] = input.KeyEvent{Sym: input.KeyF44}
		keys["kf45"] = input.KeyEvent{Sym: input.KeyF45}
		keys["kf46"] = input.KeyEvent{Sym: input.KeyF46}
		keys["kf47"] = input.KeyEvent{Sym: input.KeyF47}
		keys["kf48"] = input.KeyEvent{Sym: input.KeyF48}
		keys["kf49"] = input.KeyEvent{Sym: input.KeyF49}
		keys["kf50"] = input.KeyEvent{Sym: input.KeyF50}
		keys["kf51"] = input.KeyEvent{Sym: input.KeyF51}
		keys["kf52"] = input.KeyEvent{Sym: input.KeyF52}
		keys["kf53"] = input.KeyEvent{Sym: input.KeyF53}
		keys["kf54"] = input.KeyEvent{Sym: input.KeyF54}
		keys["kf55"] = input.KeyEvent{Sym: input.KeyF55}
		keys["kf56"] = input.KeyEvent{Sym: input.KeyF56}
		keys["kf57"] = input.KeyEvent{Sym: input.KeyF57}
		keys["kf58"] = input.KeyEvent{Sym: input.KeyF58}
		keys["kf59"] = input.KeyEvent{Sym: input.KeyF59}
		keys["kf60"] = input.KeyEvent{Sym: input.KeyF60}
		keys["kf61"] = input.KeyEvent{Sym: input.KeyF61}
		keys["kf62"] = input.KeyEvent{Sym: input.KeyF62}
		keys["kf63"] = input.KeyEvent{Sym: input.KeyF63}
	}

	return keys
}
