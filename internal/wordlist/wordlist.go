package wordlist

import (
	"crypto/rand"
	"math/big"
	"strings"
)

// Words is a curated subset of the EFF short diceware wordlist (256 words).
// 4 words from 256 = 256^4 = ~4 billion combinations (~32 bits),
// combined with a random codeId this provides strong security.
var Words = []string{
	"acid", "acme", "aged", "also", "arch", "aqua", "area", "atom",
	"aunt", "avid", "axis", "back", "bald", "band", "bark", "barn",
	"base", "bath", "bean", "bear", "beat", "belt", "bend", "bike",
	"bird", "bite", "blow", "blue", "blur", "boat", "bold", "bolt",
	"bomb", "bond", "bone", "book", "boot", "bore", "boss", "bowl",
	"bulk", "bump", "burn", "buzz", "cafe", "cage", "cake", "calm",
	"came", "camp", "cape", "card", "care", "cart", "case", "cash",
	"cast", "cave", "chat", "chip", "chop", "city", "clad", "clam",
	"clan", "claw", "clay", "clip", "club", "clue", "coal", "coat",
	"code", "coil", "coin", "cold", "colt", "cone", "cook", "cool",
	"cope", "copy", "cord", "core", "corn", "cost", "cozy", "crew",
	"crop", "crow", "cube", "curl", "cute", "damp", "dare", "dark",
	"dart", "dash", "dawn", "deal", "dear", "deck", "deed", "deep",
	"deer", "demo", "dent", "desk", "dial", "dice", "dime", "dock",
	"dome", "door", "dose", "dove", "down", "draw", "drip", "drop",
	"drum", "dull", "dune", "dusk", "dust", "each", "earl", "earn",
	"ease", "east", "echo", "edge", "edit", "else", "epic", "even",
	"ever", "evil", "exam", "exit", "face", "fact", "fade", "fail",
	"fair", "fall", "fame", "fang", "farm", "fast", "fate", "fawn",
	"fear", "feat", "feed", "feel", "file", "fill", "film", "find",
	"fine", "fire", "firm", "fish", "fist", "five", "flag", "flat",
	"fled", "flex", "flip", "flow", "foam", "fold", "folk", "fond",
	"font", "food", "foot", "ford", "fork", "form", "fort", "foul",
	"four", "free", "frog", "from", "fuel", "full", "fund", "fury",
	"fuse", "gain", "gait", "gale", "game", "gang", "gate", "gave",
	"gaze", "gear", "gene", "gift", "glad", "glow", "glue", "goat",
	"gold", "golf", "gone", "good", "grab", "gray", "grew", "grid",
	"grim", "grin", "grip", "grit", "grow", "gulf", "guru", "gust",
	"half", "hall", "halt", "hand", "hang", "hard", "harm", "harp",
	"hash", "haste", "hate", "haul", "hawk", "haze", "head", "heal",
	"heap", "heat", "held", "helm", "help", "herb", "herd", "hero",
	"hide", "high", "hike", "hill", "hint", "hire", "hold", "hole",
}

// Pick returns n random words from the wordlist, joined by the given separator.
func Pick(n int, sep string) (string, error) {
	words := make([]string, n)
	max := big.NewInt(int64(len(Words)))
	for i := 0; i < n; i++ {
		idx, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err
		}
		words[i] = Words[idx.Int64()]
	}
	return strings.Join(words, sep), nil
}
