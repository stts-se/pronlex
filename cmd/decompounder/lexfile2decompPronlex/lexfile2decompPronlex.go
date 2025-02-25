// NB: One-off hack that is not intended to be maintained. It's
// specific to a Swedish pronunciation dictionary (file).
// Takes a lexfile in WS format and tries to split the (first)
// transcription of each word into compound parts according to the
// word parts field, where the entry is decompounded.

package main

import (
	"bufio"
	"compress/gzip"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/stts-se/pronlex/lex"
	"github.com/stts-se/pronlex/line" // For formatting output file
)

var char2phon = map[rune]string{

	'a': "(?:a|A:)",
	'b': "b",
	'c': "[skxC]", // charlott c -> x, check c -> C
	'd': "[dj]",   // 'djur' d -> j
	'e': "(?:e:?|\\}:|E)",
	'é': "e:",
	'f': "f",
	'g': "[jgNx]", // -ing g -> N, gelé g -> x
	'h': "[hj]",   // hjort h -> j
	'i': "(?:I|i:)",
	'j': "[jx]", // Jeanette j -> x
	'k': "[kC]", // kärr k -> C
	'l': "[lj]", // ljud l -> j
	'm': "m",
	'n': "[nN]", // nypon+gränd
	'o': "(?:U|u:|O|o:)",
	'p': "p",
	'q': "k",
	'r': "r",
	's': "[sx]", // måls+skytt s -> x
	't': "[tC]", // tjänst t -> C
	'u': "(?:\\}:|u0)",
	'v': "v",
	'w': "v",
	'x': "s", // k
	'y': "(?:Y|y:)",
	'z': "s",
	'å': "(?:O|o:)",
	'ä': "(?:E:?|\\{:?|e)",
	'ö': "(?:2:?|9:?)",

	'7': "n", // 7-eleven
	'4': "a", // a4

	//TODO: Is this OK?
	'-': "",
}

type sm struct {
	suff    string
	matcher string
}

var suffixMatchers = []sm{
	{"ampere", "p {: r"},
	{"baby", "b I"},
	{"baisse", "b E: s"},
	{"bearnaise", "n E: s"},
	{"bayonne", "j O n"},
	{"bordeaux", "rd o:"},
	{"bourgogne", "g O n j"},
	{"braille", "b r a j"},
	{"boy", "b O j"},
	{"champagne", "p a n j"},
	{"college", "l I t C"},
	{"collage", "l A: rs"},
	{"colli", "k O [.] l I"},
	{"cockney", "k n I"},
	{"dessert", "s {: r"},
	{"nnie", "I"}, // Annie
	{"brie", "b r i:"},
	{"vue", "v (?:y:|Y)"}, //Bellevue
	{"vägen", "v E: [.] g e n"},
	{"bridge", "I (?:r[dt] rs|t C)"},
	{"frey", "E j"},
	//sm{"arbitrage", "A: rs"},
	//sm{"garage", "A: rs"},
	//sm{"plantage", "A: rs"},
	{"age", "A: rs"},
	{"allonge", "N rs"},
	{"geneve", "n E: v"},
	{"gustaf", "s t A: v"},
	{"hav", "h a f"},
	{"horney", "rn I"},
	{"ville", "v I l"},
	{"jacob", "k O p"},
	{"jakob", "k O p"},
	{"suu", "s u:"},
	{"kai", "k a j"},
	{"mai", "m a j"},
	{"may", "m a j"},
	{"thai", "t a j"},
	{"rose", "r o: s"}, // 'rose-marie'
	{"konsert", "s {: r"},
	{"träd", "t r E:?"},
	{"marie", "r (?:i:|I)"},
	{"ry", "r I"}, // 'mary'
	{"sch", "rs"}, // 'marsch'
	{"oecd", "d e:"},
	{"ph", "f"},
	{"renault", "n o:"},
	{"service", "I s"},
	{"skrid", "s k r I"},
	{"svend", "s v e n"},
	{"blind", "b l I n"},
	{"bond", "b U n"},
	{"stad", "s t a"}, // 'Vrigstad'
	{"zenith", "n I t"},
	{"ou", "u:"},
	{"ai", "a j"}, //Nicolai
}

func canMatchPrefix(p string) string {
	res := ""

	//fmt.Println(p)

	if p == "hundra" {
		return "h u0 n . d r a"
	}
	if p == "åttio" {
		return "O . t I . U"
	}

	if p == "g" { // 'g+dur'
		return "g e:"
	}
	if p == "hp" { // 'hp+färg'
		return "h o: . p e:"
	}
	if p == "pk" { // 'pk+banken'
		return "p e: . k o:"
	}
	if p == "tt" { // 'tt+afp'
		return "t e: . t e:"
	}
	if p == "tv" { // 'tv+shop'
		return "t e: . v e:"
	}

	if p == "och" { // 'berg+och+dal+bana'
		return "O(?: k)?"
	}

	if p == "anne" {
		return "a [nN]"
	}
	if p == "bernadotte" {
		return "d O t"
	}

	for _, sm := range suffixMatchers {
		if strings.HasSuffix(p, sm.suff) {
			return sm.matcher
		}
	}

	return res
}

func canMatchChar(c rune) (string, error) {
	var res = "XXXXZZZXXX"
	var err error

	if phon, ok := char2phon[c]; ok {
		res = phon
	} else {
		msg := fmt.Sprintf("svNSTHeuristicTransDecomp.CanMatchChar: unknown char: '%s'\n", string(c))
		//fmt.Fprintf(os.Stderr, msg)
		err = fmt.Errorf("%s", msg)
	}

	return res, err
}

func makeRes(matchStrings ...string) string {
	res := ""
	res0 := []string{}
	for _, m := range matchStrings {
		if m != "" {
			res0 = append(res0, m)
		}
	}

	if len(res0) == 1 {
		return res0[0]
	}
	if len(res0) > 0 {
		res1 := strings.Join(res0, "|")
		res = "(?:" + res1 + ")"
	}

	//fmt.Println(res)
	return res
}

func canMatchLHS(lhs string) (string, error) {
	var res string
	var err error
	if lhs == "" {
		return res, err
	}

	pMatcher := canMatchPrefix(lhs)

	runes := []rune(lhs)
	charMatcher, err := canMatchChar(runes[len(runes)-1])
	//if canMatch != "" {
	//	res = canMatch
	//}

	res = makeRes(pMatcher, charMatcher)

	return res, err
}

func canMatchRHS(rhs string) (string, error) {
	var res string
	var err error
	if rhs == "" {
		return res, err
	}

	runes := []rune(rhs)
	canMatch, err := canMatchChar(runes[0])
	if canMatch != "" {
		res = canMatch
	}

	return res, err
}

func endsInPotentialRetroflex(s string) bool {
	if strings.HasSuffix(s, "rs") {
		return true
	}
	if strings.HasSuffix(s, "rt") {
		return true
	}
	if strings.HasSuffix(s, "rts") {
		return true
	}
	if strings.HasSuffix(s, "rd") {
		return true
	}
	if strings.HasSuffix(s, "rds") {
		return true
	}

	if strings.HasSuffix(s, "rn") {
		return true
	}
	if strings.HasSuffix(s, "rns") {
		return true
	}

	if strings.HasSuffix(s, "rl") {
		return true
	}
	if strings.HasSuffix(s, "rls") {
		return true
	}

	if strings.HasSuffix(s, "r") {
		return true
	}

	return false
}

func startsWithPotentialRetroflex(s string) bool {
	if strings.HasPrefix(s, "s") {
		return true
	}
	if strings.HasPrefix(s, "c") { //motor+Cykel
		return true
	}
	if strings.HasPrefix(s, "z") { //buffert+zon
		return true
	}

	if strings.HasPrefix(s, "t") {
		return true
	}
	if strings.HasPrefix(s, "d") {
		return true
	}
	if strings.HasPrefix(s, "n") {
		return true
	}
	if strings.HasPrefix(s, "l") {
		return true
	}

	return false
}
func retroflexation(lhs, rhs string) bool {
	lhsRetro := endsInPotentialRetroflex(lhs)
	rhsRetro := startsWithPotentialRetroflex(rhs)

	return lhsRetro && rhsRetro
}

var nonPhoneme = map[string]bool{

	".":  true,
	"%":  true,
	`"`:  true,
	`""`: true,
}

var isRetroflex = map[string]bool{
	"rs": true,
	"rt": true,
	"rn": true,
	"rd": true,
	"rl": true,
}

// remove initial retroflexation
func deRetroflex(s string) string {
	var res []string
	syms := strings.Split(strings.TrimSpace(s), " ")
	done := false
	for _, sym := range syms {
		if nonPhoneme[sym] {
			res = append(res, sym)
			continue
		}
		if !done && isRetroflex[sym] {
			sym = strings.Replace(sym, "r", "", 1)
			res = append(res, sym)
		} else {
			done = true
			res = append(res, sym)
		}
	}

	return strings.Join(res, " ")
}

var nonPhonemes = `[ .%"]*`

// doubleChar returns true iff lhs ends with the same rune as rhs
// starts with
/*func doubleChar(lhs, rhs string) (rune, bool) {
	fmt.Printf("LHS: %s\tRHS: %s\n", lhs, rhs)

	var res rune
	var isSame bool

	if lhs == "" || rhs == "" {
		return res, isSame
	}

	lhsRunes := []rune(lhs)
	lastLHS := lhsRunes[len(lhsRunes)-1]

	rhsRunes := []rune(rhs)
	firstRHS := rhsRunes[0]

	if lastLHS == firstRHS {
		return lastLHS, true
	}

	return res, isSame
}
*/
func splitTrans(lhs, rhs, trans string) (string, string, error) {

	lhs = strings.ToLower(lhs)
	rhs = strings.ToLower(rhs)
	lhs0 := lhs

	//TODO: Does this work? Should replace strings with []rune, to be shure that things work?

	//fmt.Println("lhs: " + lhs)
	//fmt.Println("rhs: " + rhs)

	// If retroflexation spanns the compound boundary, matching is
	// a bit more tricky
	retro := retroflexation(lhs0, rhs)
	//fmt.Printf("retro: %v\n", retro)

	// Final -r "jumps" to rhs transcription, thus the character
	// before final 'r' should be used for matching lhs
	//TODO: Does this work? Should replace strings with []rune, to be shure that things work?
	if retro && strings.HasSuffix(lhs0, "r") && !strings.HasSuffix(lhs0, "rr") && len(lhs0) > 1 {
		lhs0 = lhs0[0 : len(lhs0)-1]
	}

	//TODO: Handling double chars over compound boundaries more complicated than I first though.
	// One reason is that the last char of lhs is not always silent.

	// Chech whether lhs ends with the same rune that rhs starts
	// with. If so, the last char of lhs may be silent in the
	// transcription of lhs
	// _, isDouble := doubleChar(lhs, rhs)
	// //fmt.Printf(">>>> %v %v\n", char, isDouble)
	// if !retro && isDouble {
	// 	// Knock of the last char, and hope that is enough
	// 	lhs0 = lhs0[0 : len(lhs0)-1]
	// }

	lhsM, lErr := canMatchLHS(lhs0)
	//fmt.Printf("lhsM: %s\n", lhsM)
	//fmt.Printf("lErr: %v\n", lErr)

	// if final -r has been removed from lhs, we still need
	// optional matching of ' r'
	if retro && len(lhs) > len(lhs0) {
		lhsM = lhsM + "(?: r)?"
	}
	//	if !retro && isDouble && len(lhs) > len(lhs0) {
	//		lhsM = lhsM + "(?: " + +")?"
	//	}

	rhsM, rErr := canMatchRHS(rhs)

	// TODO This is not how you handle disjunctive errors
	var errr error
	if lErr != nil {
		if errr != nil {
			errr = fmt.Errorf("%v : %v", errr, lErr)
		} else {
			errr = fmt.Errorf("%v", lErr)
		}
	}
	if rErr != nil {
		if errr != nil {
			errr = fmt.Errorf("%v : %v", errr, rErr)
		} else {
			errr = fmt.Errorf("%v", rErr)
		}
	}

	if retro {
		rhsM = "r?" + rhsM
	}

	reStrn := lhsM + "([ .]+)" + nonPhonemes + rhsM

	//fmt.Println(reStrn)

	// TODO Panics on incorrect RE
	re := regexp.MustCompile(reStrn)

	indxs := re.FindStringSubmatchIndex(trans)
	if len(indxs) != 4 {
		return "", trans, errr
	}

	t1 := strings.TrimSpace(trans[0:indxs[2]])
	t2 := strings.TrimSpace(trans[indxs[3]:])

	if retro {
		// We might have to add final ' r' to lhs if it has
		// hopped compound boundary:
		if strings.HasSuffix(lhs, "r") && !strings.HasSuffix(t1, " r") {
			t1 = t1 + " r"
		}

		t2 = deRetroflex(t2)
	}
	//t2 = deRetroflex(t2)
	return t1, t2, errr
}

// HerusticSvNSTTransDecomp is a heuristic trans decompounder for Swedish
func HerusticSvNSTTransDecomp(orthDecomp, trans string) ([]string, error) {
	var res []string
	var err error
	decomps := strings.Split(orthDecomp, "+")
	// Loop over pairs of decomps

	transSoFar := trans
	for i := 0; i < len(decomps)-1; i++ {
		lhs := strings.TrimSpace(decomps[i])
		rhs := strings.TrimSpace(decomps[i+1])
		if lhs == "-" || lhs == "" {
			continue
		}

		tLHS, tRHS, err0 := splitTrans(lhs, rhs, transSoFar)
		if err0 != nil {
			if err == nil {
				err = err0
			} else {
				err = fmt.Errorf("%v : %v", err, err0)
			}
		}
		if tLHS != "" {
			res = append(res, tLHS)
		}
		transSoFar = tRHS
	}
	res = append(res, transSoFar)
	//fmt.Println()

	if len(res) != len(decomps) {
		msg := "failed to match decomps and transparts"
		if err == nil {
			err = errors.New(msg)
		} else {
			err = fmt.Errorf("%v : %s", err, msg)
		}
	}

	return res, err
}

func cleanUpDecomp(d string) string {
	res := strings.ToLower(d)
	res = strings.Replace(res, "!", "", -1)
	res = strings.Replace(res, "+s+", "s+", -1)
	res = strings.Replace(res, "+i+", "i+", -1) // Nikola+i+
	res = strings.Replace(res, "+o+", "o+", -1)
	res = strings.Replace(res, "+a+", "a+", -1)
	res = strings.Replace(res, "+u+", "u+", -1)
	res = strings.Replace(res, "+g+", "g+", -1) // armerin+g+
	res = strings.Replace(res, "+ar+", "ar+", -1)
	res = strings.Replace(res, "+r+", "r+", -1)
	res = strings.Replace(res, "+ra+", "ra+", -1)
	res = strings.Replace(res, "+es+", "es+", -1)
	res = strings.Replace(res, "+na+", "na+", -1)
	res = strings.Replace(res, "+on+", "on+", -1)
	res = strings.Replace(res, "+e+", "e+", -1)
	res = strings.Replace(res, "+-+", "+", -1)
	if strings.HasPrefix(res, "+-+") {
		res = strings.Replace(res, "+-+", "", 1)
	}
	if strings.HasPrefix(res, "+") {
		res = strings.Replace(res, "+", "", 1)
	}

	if strings.HasSuffix(res, "+s") {
		res = strings.Replace(res, "+s", "s", 1)
	}

	return res
}

// WP defines a word part
type WP struct {
	trans string
	pos   string
	morph string
}

var prefixLex = make(map[string]map[WP]int)
var infixLex = make(map[string]map[WP]int)
var suffixLex = make(map[string]map[WP]int)

func add(strn, trans, pos, morph string, lex map[string]map[WP]int) {
	if v, ok := lex[strn]; ok {
		v[WP{trans: trans, pos: pos, morph: morph}]++
	} else {
		m := make(map[WP]int)
		wp := WP{trans: trans, pos: pos, morph: morph}
		m[wp]++
		lex[strn] = m
	}
}

func addWordParts(wps string, trans []string, pos, morph string) {
	var wps0 = strings.Split(wps, "+")

	if len(wps0) != len(trans) {
		fmt.Fprintf(os.Stderr, "skipping input: different len: %v vs %v\n", wps0, trans)
	}

	if len(wps0) < 2 {
		return
	}

	if len(wps0) > 1 {
		add(wps0[0], trans[0], "", "", prefixLex)
	}

	if len(wps0) > 2 {
		//add(wps0[0], trans[0], prefixLex)
		for i := 1; i < len(wps0)-1; i++ {
			add(wps0[i], trans[i], "", "", infixLex)
		}
	}

	last := len(wps0) - 1
	add(wps0[last], trans[last], pos, morph, suffixLex)

	//fmt.Println(wps)
}

// Freq defines a word part frequency
type Freq struct {
	Word WP
	Freq int
}

func freqSort(m map[WP]int) []Freq {
	var res []Freq
	for k, v := range m {
		res = append(res, Freq{k, v})
	}

	sort.Slice(res, func(i, j int) bool { return res[i].Freq > res[j].Freq })

	return res
}

func totFreq(fs []Freq) int {
	var res int
	for _, f := range fs {
		res += f.Freq
	}
	return res
}

// Filter that returns unique transcriptions over frequency min, and discards the rest.
// Starting with the highest frequency transcription.
func minFreqAndDifferentTrans(fs []Freq, min int) []Freq {
	var res []Freq
	seenTrans := make(map[string]bool)

	sort.Slice(fs, func(i, j int) bool { return fs[i].Freq > fs[j].Freq })

	for _, f := range fs {
		trans := f.Word.trans
		if f.Freq >= min && !seenTrans[trans] {
			seenTrans[trans] = true
			res = append(res, f)
		}
	}

	return res
}

// Filter that collapses entries of the same POS and MORPH values. The
// most frequent transcription is placed first in a slice.
func collapseEntriesWithSamePOS(w string, f []Freq) []lex.Entry {
	var res []lex.Entry

	// Keeps track of already seen POS+MORPH
	seen := make(map[string]bool)

	for _, s := range f {

		posMorph := s.Word.pos + s.Word.morph

		t := lex.Transcription{Strn: s.Word.trans}
		e := lex.Entry{
			Strn:           w,
			WordParts:      w,
			PartOfSpeech:   s.Word.pos,
			Morphology:     s.Word.morph,
			Transcriptions: []lex.Transcription{t},
		}

		if ok := seen[posMorph]; !ok {
			res = append(res, e)
			seen[posMorph] = true
		} else { // Add transcription variant to already existing Entry
			for i, ex := range res {
				if ex.PartOfSpeech == s.Word.pos && ex.Morphology == s.Word.morph {
					ts := res[i].Transcriptions
					ts = append(ts, t)
					res[i].Transcriptions = ts
				}
			}
		}
	}
	return res
}

/*
	func dump(m map[string]map[WP]int) {
		for k, v := range m {
			srt := freqSort(v)
			tot := totFreq(srt)
			min := 4
			if tot > 50 {
				min = 20
			}
			fltr := minFreqAndDifferentTrans(srt, min)
			if len(fltr) == 0 && len(srt) > 0 {
				fltr = srt[0:1]
			}
			if tot > 5 {
				for _, s := range fltr {
					fmt.Printf("%d\t%s\t%v\n", tot, k, s)
				}
			}
		}
	}
*/
func toFile(m map[string]map[WP]int, fileName string) error {

	fh, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {

		return fmt.Errorf("lexfile2decompPronlex.toFile failed : %v", err)
	}
	/* #nosec G307 */
	defer fh.Close()

	fmt.Fprintf(os.Stderr, "printing compound parts to file '%s'\n", fileName)

	wsFmt, err := line.NewWS()
	if err != nil {
		return fmt.Errorf("lexfile2decompPronlen.toFile failed initializing : %v", err)
	}

	for k, v := range m {
		srt := freqSort(v) // []Freq
		tot := totFreq(srt)
		min := 4
		if tot > 50 {
			min = 20
		}
		fltr := minFreqAndDifferentTrans(srt, min)
		//fltr := collapseEntriesWithSamePOS(fltr0)
		if len(fltr) == 0 && len(srt) > 0 { // Nothing survived the filtering. Put back first element if any.
			fltr = srt[0:1]
		}
		if tot > 5 {
			// TODO: If same POS+MORPH but different trans, collapse into one lex.Entry
			entries := collapseEntriesWithSamePOS(k, fltr)
			for _, e := range entries {

				es, err := wsFmt.Entry2String(e)
				if err != nil {
					return fmt.Errorf("lexfile2decompPronlex.toFile failed formatting : %v", err)
				}

				//fmt.Printf("%d\t%s\t%s\n", tot, k, es)
				fmt.Fprintf(fh, "%s\n", es)
			}
		}
	}

	//fh.Flush()

	return nil
}

func main() {

	if len(os.Args) != 2 && len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, filepath.Base(os.Args[0]), "<lexicon file> <N errors before exit>?")
		os.Exit(1)
	}

	fn := os.Args[1]
	fh, err := os.Open(filepath.Clean(fn))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Mabel Tainter Memorial Building! : %v\n", err)
	}

	var s *bufio.Scanner
	if strings.HasSuffix(fn, ".gz") {
		gz, err := gzip.NewReader(fh)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Streptomyces tsukubaensis! : %v\n", err)
		}
		s = bufio.NewScanner(gz)
	} else {
		s = bufio.NewScanner(fh)
	}

	exitAfter := 0
	if len(os.Args) == 3 {
		i, err := strconv.Atoi(os.Args[2])
		if err != nil {
			msg := fmt.Sprintf("Second optional argument should be an integer, got '%s'", os.Args[2])
			fmt.Fprintf(os.Stderr, "%s\n", msg)
			os.Exit(1)
		} //else {
		exitAfter = i
		//}
	}

	fails := 0

	for s.Scan() {
		l := s.Text()

		wsFmt, err := line.NewWS()
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to initialize line formatter : %v\n", err)
			os.Exit(0)
		}

		e, err := wsFmt.ParseToEntry(l)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to parse line : %v\n", err)
			os.Exit(0)
		}

		// disregard items withs spurious morphology:
		if strings.Contains(e.Morphology, "|||") || strings.HasPrefix(e.Morphology, " |") {
			continue
		}

		pos := e.PartOfSpeech
		morph := e.Morphology
		decomp0 := e.WordParts
		decomp := cleanUpDecomp(decomp0)
		firstTrans := e.Transcriptions[0].Strn // fs[8]
		if strings.Contains(decomp, "+") {
			//fmt.Printf("%s %s %s\n", orth, decomp, firstTrans)

			rez, err := HerusticSvNSTTransDecomp(decomp, firstTrans)
			if err != nil {
				fmt.Fprintf(os.Stderr, "FAIL: %v : %s\t%s\t%#v\n", err, decomp, firstTrans, rez)
				fails++
				if fails >= exitAfter {
					os.Exit(1)
				}
				continue
			}

			addWordParts(decomp, rez, pos, morph)

		}
	}

	// TODO: File base name as command line arg instead
	outFileBaseName := strings.Replace(filepath.Base(fn), ".gz", "", -1)
	err = toFile(suffixLex, outFileBaseName+"_sufflex.txt")
	if err != nil {
		fmt.Fprintf(os.Stderr, "FAILURE: %v\n", err)
		os.Exit(0)
	}
	err = toFile(prefixLex, outFileBaseName+"_preflex.txt")
	if err != nil {
		fmt.Fprintf(os.Stderr, "FAILURE: %v\n", err)
		os.Exit(0)
	}
	err = toFile(infixLex, outFileBaseName+"_infixlex.txt")
	if err != nil {
		fmt.Fprintf(os.Stderr, "FAILURE: %v\n", err)
		os.Exit(0)
	}

	fmt.Fprintf(os.Stderr, "Lines failed to split: %d\n", fails)
}
