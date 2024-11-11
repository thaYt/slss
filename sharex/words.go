package sharex

import (
	_ "embed"
	"math/rand/v2"
	"strings"
)

/*
https://gist.github.com/hugsy/8910dc78d208e40de42deb29e62df913
https://github.com/taikuukaits/SimpleWordlists/blob/master/Wordlist-Nouns-Common-Audited-Len-3-6.txt
https://github.com/taikuukaits/SimpleWordlists/blob/master/Wordlist-Adjectives-Common-Audited-Len-3-6.txt
*/

//go:embed adj-list
var adjectives string

//go:embed noun-list
var nouns string

var (
	adj = strings.Split(adjectives, "\n")
	noun = strings.Split(nouns, "\n")
)

func GenPhrase(currentFiles []string) string {
	var notExists bool
	var phrase string
	for !notExists {
		phrase = adj[rand.IntN(len(adj))] + "-" + noun[rand.IntN(len(noun))]
		notExists = true
		for _, f := range currentFiles {
			if f == phrase {
				notExists = false
				break
			}
		}
	}
	return phrase
}
