// This package provides a parser for DEC ANSI escape sequences compatible with
// the VT500-series of terminals. It is based on the VT500-series state machine
// and the wonderful work of Williams, Paul Flo
// https://vt100.net/emu/dec_ansi_parser
//
// The implemented state machine include a few modifications to the original
// state machine to support UTF8 sequences and to recognize more parameters and
// states. Please refer to [GenerateTransitionTable](./table.go) for more
// details.
package parser
