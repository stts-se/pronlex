package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/stts-se/pronlex/lex"
	"github.com/stts-se/pronlex/line"
	"github.com/stts-se/pronlex/validation/rules"
	"github.com/stts-se/symbolset/mapper"
)

var langCodes = map[string]string{
	"afr": "foreign",
	"ara": "ar",
	"asi": "foreign",
	"por": "pt",
	"dan": "da-dk",
	"dut": "nl",
	"eng": "en",
	"fin": "fi-fi",
	"for": "foreign",
	"fre": "fr-fr",
	"ger": "de-de",
	"gre": "el-gr",
	"ita": "it-it",
	"jap": "ja-jp",
	"lat": "la",
	"mix": "foreign",
	"nno": "nn-no",
	"nob": "nb-no",
	"nor": "nb-no",
	"pol": "pl-pl",
	"rus": "ru-ru",
	"sfi": "sv-fi",
	"sla": "foreign",
	"spa": "es-es",
	"swe": "sv-se",
	"trt": "tr-tr",
	"tur": "tr-tr",
	"unk": "foreign",
	"ind": "foreign",
}

// map to filter out duplicates, key is 'orth <tab> trans <tab> pos <tab> lang'
var printed = make(map[string]bool)

// var upperCase = regexp.MustCompile("^[A-ZÅÄÖ]+$")
var validSymbols = regexp.MustCompile("^[áýúíóèàùìòïćãêâôûçðþěžřūčāîêť+_&:.a-zåäöéßàèùçšóáíüëïøæ0-9' -]+$")

func removableLine(orth string, line string, e lex.Entry) (string, bool) {
	trans := ""
	if len(e.Transcriptions) > 0 {
		trans = e.Transcriptions[0].Strn
	}
	key := fmt.Sprintf("%s\t%s\t%s\t%s", orth, trans, e.PartOfSpeech, e.Language)
	if _, ok := printed[key]; ok {
		return "duplicate", true
	}
	if !validSymbols.MatchString(strings.ToLower(orth)) {
		return "symbolset", true
	}
	// if upperCase.MatchString(orth) && strings.HasPrefix(e.PartOfSpeech, "RG") { // roman numerals
	// 	return true
	// } else if garbLine.MatchString(line) {
	// 	return true
	// }
	printed[key] = true
	return "", false
}

func validPos(pos string) bool {
	return true
}

func mapLanguage(lang string) (string, error) {
	if lang == "" {
		return lang, nil
	}
	l, ok := langCodes[strings.ToLower(lang)]
	if ok {
		return l, nil
	}
	fmt.Fprintf(os.Stderr, "mapping language <%v> to default lang <foreign>\n", lang)
	return "foreign", nil
}

func mapTransLanguages(e *lex.Entry) error {
	var newTs []lex.Transcription
	for _, t := range e.Transcriptions {
		l, err := mapLanguage(t.Language)
		if err != nil {
			return err
		}
		t.Language = l
		newTs = append(newTs, t)
	}
	e.Transcriptions = newTs
	return nil
}

func mapTranscription(t0 string) string {
	t := t0
	// fromRE := regexp.MustCompile("([\"',])")
	// toString := "$1 "
	// t = fromRE.ReplaceAllString(t, toString)
	t = strings.Replace(t, "'", "' ", -1)
	t = strings.Replace(t, "\"", "\" ", -1)
	t = strings.Replace(t, ",", ", ", -1)
	// syllabic := (uw:|uuh|uu:|uex|oe:|iex|eex|ae:|aa:|y:|uw|uu|un|u:|ou|on|oi|oh|oe|ö:|o:|ih|i:|ex|eu|en|ei|eh|e:|au|an|ai|ae|ä:|a:|y|u|ö|o|i|e|ä|a)")
	fromRE := regexp.MustCompile("(^|[.|~-])([^.|~-]+)([\"',])")
	toString := "$1 $3 $2"
	t = fromRE.ReplaceAllString(t, toString)
	//t = strings.Replace(t, " ", " ", -1)
	t = regexp.MustCompile("  +").ReplaceAllString(t, " ")
	t = regexp.MustCompile("^ ").ReplaceAllString(t, "")
	//t = regexp.MustCompile(" $").ReplaceAllString(t, "")
	if len(t0) > len(t) {
		panic(fmt.Sprintf("mapTranscription | Conversion error %s => %s", t0, t))
	}
	return t
}

func testMapTranscription(input string, expect string) string {
	result := mapTranscription(input)
	if result != expect {
		return fmt.Sprintf("input: %s\nexpected : %s\ngot      : %s", input, expect, result)
	}
	return ""
}

func testMapTranscriptions() {
	var res []string
	res = append(res, testMapTranscription(`n 'o l`, `' n o l`))
	var errs []string
	for _, s := range res {
		if s != "" {
			errs = append(errs, s)
		}
	}
	if len(errs) > 0 {
		panic(fmt.Sprintf("%v", errs))
	}
}

func mapTranscriptions(e *lex.Entry, mapper mapper.Mapper) error {

	for i, t := range e.Transcriptions {
		e.Transcriptions[i].Strn = mapTranscription(t.Strn)
		//fmt.Println(t.Strn)
	}
	//fmt.Println(e.Transcriptions)
	err := line.MapTranscriptions(mapper, e)
	if err != nil {
		return err
	}
	var newTs []lex.Transcription
	for _, t := range e.Transcriptions {
		//t.Strn = mapTranscription(t.Strn)
		newTs = append(newTs, t)
	}
	e.Transcriptions = newTs
	return nil
}

func main() {

	if len(os.Args) != 4 {
		fmt.Fprintln(os.Stderr, "<INPUT BRAXEN LEX FILE> <BRAXEN-SAMPA SYMBOLSET> <WS-SAMPA SYMBOLSET>")
		fmt.Fprintln(os.Stderr, "\tsample invokation: svSeBraxen2WS braxen-sv.tsv sv-braxen-sampa.sym sv-se_ws-sampa.sym")
		return
	}

	braxenFileName := os.Args[1]
	ssFileName1 := os.Args[2]
	ssFileName2 := os.Args[3]

	mapper, err := mapper.LoadMapperFromFile("SAMPA", "SYMBOL", ssFileName1, ssFileName2)
	if err != nil {
		fmt.Fprintf(os.Stderr, "couldn't load mappers: %v\n", err)
		return
	}

	testMapTranscriptions()

	ssRuleTo := rules.SymbolSetRule{SymbolSet: mapper.SymbolSet2}

	braxenFile, err := os.Open(filepath.Clean(braxenFileName))
	if err != nil {
		fmt.Fprintf(os.Stderr, "couldn't open lexicon file: %v\n", err)
		return
	}
	/* #nosec G307 */
	defer braxenFile.Close()

	braxenFmt, err := line.NewBraxen()
	if err != nil {
		log.Fatal(err)
	}
	wsFmt, err := line.NewWS()
	if err != nil {
		log.Fatal(err)
	}

	braxen := bufio.NewScanner(braxenFile)
	n := 0
	for braxen.Scan() {
		n++
		hasError := false
		if err := braxen.Err(); err != nil {
			fmt.Fprintf(os.Stderr, "general error	failed reading line %v : %v\n", n, err)
			hasError = true
		}
		line := braxen.Text()
		if line == "#" { // first line in braxen
			continue
		}

		e, origOrth, err := braxenFmt.ParseToEntry(line)
		if err != nil {
			fmt.Fprintf(os.Stderr, "general error	failed to convert line %v to entry : %v\n", n, err)
			fmt.Fprintf(os.Stderr, "general error	failing line: %v\n", line)
			hasError = true
		}

		e.EntryStatus.Name = "imported"
		e.EntryStatus.Source = "braxen"
		e.Language, err = mapLanguage(e.Language)
		if err != nil {
			fmt.Fprintf(os.Stderr, "entry language error	%v\n", err)
			hasError = true
		}
		err = mapTransLanguages(&e)
		if err != nil {
			fmt.Fprintf(os.Stderr, "trans language error	%v\n", err)
			hasError = true
		}
		if !validPos(e.PartOfSpeech) {
			fmt.Fprintf(os.Stderr, "pos error	invalid pos tag <%v>\n", e.PartOfSpeech)
			hasError = true
		}

		err = mapTranscriptions(&e, mapper)
		if err != nil {
			fmt.Fprintf(os.Stderr, "transcription error	failed to map transcription symbols for %s : %v\n", e.Strn, err)
			hasError = true
		}

		if msg, remove := removableLine(origOrth, line, e); remove {
			fmt.Fprintf(os.Stderr, "skipping line\t%s\t%v\n", msg, line)
			continue
		}

		if !hasError {
			valres, err := ssRuleTo.Validate(e)
			if err != nil {
				panic(err) // shouldn't happen
			}
			for _, r := range valres.Messages {
				panic(r) // shouldn't happen
			}

			res, err := wsFmt.Entry2String(e)
			if err != nil {
				fmt.Fprintf(os.Stderr, "failed to convert entry to string : %v\n", err)
			} else {
				fmt.Printf("%v\n", res)
			}
		}
	}

	_ = lex.Entry{}
}
