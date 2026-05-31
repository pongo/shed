package main

import "strings"

// repairPlatformArgs repairs cmd.exe argv after a quoted folder ending in a
// backslash swallows the closing quote and fuses the following flags into the
// same argument.
//
// Example:
//
//	shed "C:\Users\me\Downloads\Telegram Desktop2\" --age 0
//
// can arrive as:
//
//	[]string{`C:\Users\me\Downloads\Telegram Desktop2" --age 0`}
//
// pflag cannot parse --age until that fused argument is split back into a
// folder argument and flag arguments.
func repairPlatformArgs(args []string) []string {
	repaired := make([]string, 0, len(args))
	for _, arg := range args {
		folder, rest, ok := strings.Cut(arg, `" --`)
		if !ok {
			repaired = append(repaired, arg)
			continue
		}

		repaired = append(repaired, folder)
		repaired = append(repaired, strings.Fields("--"+rest)...)
	}
	return repaired
}
