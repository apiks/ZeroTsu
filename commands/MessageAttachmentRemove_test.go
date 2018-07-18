package commands

import "testing"

func TestWhitelistFilenames(t *testing.T) {
	goodFiles := []string{"123.png", "aFwofEhFJ.jpeg"}
	badFiles := []string{"elf", "something.exe"}
	for _, filename := range goodFiles {
		if !isAllowed(filename) {
			t.Errorf("%s should be a whitelisted filename", filename)
		}
	}
	for _, filename := range badFiles {
		if isAllowed(filename) {
			t.Errorf("%s should not be a whitelisted filename", filename)
		}
	}
}
