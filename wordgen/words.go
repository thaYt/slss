package wordgen

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

func GenPhrase() string {
	adj := strings.Split(adjectives, "\n")
	noun := strings.Split(nouns, "\n")

	arand := rand.IntN(len(adj))
	nrand := rand.IntN(len(noun))

	return adj[arand] + "-" + noun[nrand]
}
