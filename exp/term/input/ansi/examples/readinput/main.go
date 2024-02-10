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

	// for {
	// 	buf := [256]byte{}
	// 	n, err := os.Stdout.Read(buf[:])
	// 	if err != nil {
	// 		log.Fatalf("error reading input: %v\r\n", err)
	// 	}
	//
	// 	log.Printf("read %d bytes: %q\r\n", n, buf[:n])
	// }

	r := bufio.NewReader(os.Stdout)
	d := ansi.NewDriver(r, ansi.Stdflags)
	buf := [1]input.Event{}
	for {
		n, ev, err := d.ReadInput()
		if err != nil {
			if errors.Is(err, input.ErrUnknownEvent) {
				log.Printf("%v\r\n", err)
				r.Discard(n)
				continue
			}
			log.Fatalf("error reading input: %v\r\n", err)
		}

		// Gracefully exit on 'qq'
		if buf[0] != nil {
			k, ok1 := ev.(input.KeyEvent)
			p, ok2 := buf[0].(input.KeyEvent)
			if ok1 && ok2 && k.Rune == 'q' && p.Rune == 'q' {
				break
			}
		}

		log.Printf("event: %s (len: %d)\r\n", ev, n)
		buf[0] = ev
	}
}
