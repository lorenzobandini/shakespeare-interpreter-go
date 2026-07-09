package parser

import "strings"

// Curated SPL word classification. ~80 entries. Unknown nouns default to neutral (+1).
// ponytail: minimal curated set; unknown words fall through to defaults (noun=+1, adj=count, comp=positive).

var negativeNouns = map[string]bool{
	"coward": true, "liar": true, "fool": true, "pig": true, "blister": true,
	"leech": true, "codpiece": true, "beggar": true, "thief": true, "villain": true,
	"traitor": true, "knave": true, "rat": true, "toad": true, "plague": true,
	"famine": true, "pestilence": true, "misery": true, "bastard": true,
	"idiot": true, "moron": true, "trash": true, "garbage": true, "dirt": true,
	"shadow": true, "venom": true, "poison": true, "scum": true, "hound": true,
	"mistletoe": true,
}

var adjectives = map[string]bool{
	"red": true, "hot": true, "big": true, "handsome": true, "rich": true,
	"brave": true, "beautiful": true, "fair": true, "warm": true, "peaceful": true,
	"sunny": true, "sweet": true, "good": true, "great": true, "fine": true,
	"lovely": true, "amazing": true, "bold": true, "cute": true, "fat": true,
	"little": true, "stuffed": true, "misused": true, "dusty": true, "old": true,
	"rotten": true, "green": true, "huge": true, "large": true, "rural": true,
	"bottomless": true, "embroidered": true, "bluest": true, "clearest": true,
	"sweetest": true, "reddest": true, "smooth": true, "small": true, "furry": true,
	"white": true, "black": true, "half-witted": true, "stupid": true,
	"lying": true, "fatherless": true, "smelly": true, "vile": true,
	"cowardly": true, "worried": true, "bad": true, "sick": true, "dead": true,
	"young": true, "remarkable": true, "infected": true, "oozing": true,
	"summer's": true, "proud": true, "mighty": true, "healthy": true,
	"hairy": true, "sorry": true, "disgusting": true, "loving": true,
	"likewise": true, "insulting": true, "flattering": true,
}

var negativeComparatives = map[string]bool{
	"worse": true, "smaller": true, "poorer": true, "uglier": true, "meaner": true,
	"baser": true, "lower": true, "weaker": true, "slower": true, "colder": true,
	"darker": true, "fouler": true, "viler": true, "sicker": true, "deader": true,
	"bad": true, "small": true, "cowardly": true, "disgusting": true,
}

var possessivePronouns = map[string]bool{
	"my": true, "thy": true, "your": true, "his": true, "her": true,
	"mine": true, "thine": true, "our": true,
}

var speakerPronouns = map[string]bool{
	"me": true, "myself": true,
}

var listenerPronouns = map[string]bool{
	"thyself": true, "yourself": true,
}

var articles = map[string]bool{
	"a": true, "an": true, "the": true,
}

var shakespeareCharacters = map[string]bool{
	"romeo": true, "juliet": true, "hamlet": true, "ophelia": true,
	"macbeth": true, "lady macbeth": true, "lear": true, "cordelia": true,
	"othello": true, "desdemona": true, "iago": true, "prospero": true,
	"miranda": true, "ariel": true, "caliban": true, "portia": true,
	"shylock": true, "antonio": true, "falstaff": true, "henry": true,
	"richard": true, "cleopatra": true, "caesar": true, "brutus": true,
	"cassius": true, "kate": true, "petruchio": true,
	"oberon": true, "titania": true, "puck": true, "bottom": true,
	"malvolio": true, "viola": true, "olivia": true, "orsino": true,
	"shakespeare": true, "andersen": true,
}

// Lookup functions.

func lower(w string) string { return strings.ToLower(w) }

func isAdjective(word string) bool {
	return adjectives[lower(word)]
}

func nounPolarity(word string) int {
	w := lower(word)
	if negativeNouns[w] {
		return -1
	}
	return 1
}

func comparativePolarity(word string) string {
	w := lower(word)
	if negativeComparatives[w] {
		return "negative"
	}
	return "positive"
}

func isPossessive(word string) bool      { return possessivePronouns[lower(word)] }
func isSpeakerPronoun(word string) bool  { return speakerPronouns[lower(word)] }
func isListenerPronoun(word string) bool { return listenerPronouns[lower(word)] }
func isArticle(word string) bool         { return articles[lower(word)] }
func isShakespeareCharacter(name string) bool {
	return shakespeareCharacters[lower(name)]
}

// isOpKeyword reports whether a word following "the" is an operator keyword.
func isOpKeyword(word string) bool {
	switch word {
	case "sum", "product", "difference", "quotient", "remainder",
		"square", "cube", "factorial":
		return true
	}
	return false
}

// parseRoman converts a Roman numeral string (I, II, III, IV, V, ..., XXXIX) to int.
// Returns (0, false) if invalid.
func parseRoman(s string) (int, bool) {
	if s == "" {
		return 0, false
	}
	values := map[byte]int{'I': 1, 'V': 5, 'X': 10, 'L': 50, 'C': 100, 'D': 500, 'M': 1000}
	prev := 0
	total := 0
	for i := len(s) - 1; i >= 0; i-- {
		v, ok := values[s[i]]
		if !ok {
			return 0, false
		}
		if v < prev {
			total -= v
		} else {
			total += v
			prev = v
		}
	}
	if total <= 0 || total > 3999 {
		return 0, false
	}
	return total, true
}
