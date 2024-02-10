package main

import (
	"bufio"
	"errors"
	"log"
	"os"

	"github.com/charmbracelet/x/exp/term"
	"github.com/charmbracelet/x/exp/term/input"
	"github.com/charmbracelet/x/exp/term/input/ansi"
)

func main() {
	state, err := term.MakeRaw(os.Stdin.Fd())
	if err != nil {
		log.Fatalf("error making raw: %v", err)
	}

	defer term.Restore(os.Stdin.Fd(), state)

	// r := bufio.NewReader(strings.NewReader("\x00\x1ba\x1b[Z\x1b\x01\x1b[A"))
	rd := ansi.NewDriver(bufio.NewReaderSize(os.Stdin, 256), os.Getenv("TERM"), ansi.Stdflags)

	// p, err := d.PeekInput(2)
	// if err != nil {
	// 	log.Fatalf("error peeking input: %v\r\n", err)
	// }
	//
	// for _, e := range p {
	// 	log.Printf("event: %s (len: %d)\r\n\r\n", e, len(p))
	// }

	// go func() {
	// 	time.Sleep(2 * time.Second)
	// 	io.WriteString(os.Stdout, "\x1b[?u\x1b[c\x1b]11;?\x07")
	// }()

	lastEv := input.Event(nil)
	for {
		n, err := rd.ReadInput()
		if err != nil {
			if errors.Is(err, input.ErrUnknownEvent) {
				log.Printf("%v\r\n", err)
				continue
			}
			log.Fatalf("error reading input: %v\r\n", err)
		}

		// Gracefully exit on 'qq'
		if lastEv != nil {
			k, ok1 := n[len(n)-1].(input.KeyEvent)
			p, ok2 := lastEv.(input.KeyEvent)
			if ok1 && ok2 && k.Rune == 'q' && p.Rune == 'q' {
				break
			}
		}

		for _, e := range n {
			log.Printf("event: %s (len: %d)\r\n\r\n", e, len(n))
		}
		if len(n) > 0 {
			lastEv = n[len(n)-1]
		}
	}
}
