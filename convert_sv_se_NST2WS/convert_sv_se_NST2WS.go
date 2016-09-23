package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/stts-se/pronlex/lex"
	"github.com/stts-se/pronlex/line"
	"github.com/stts-se/pronlex/symbolset"
	"github.com/stts-se/pronlex/vrules"
)

var sucTags = map[string]bool{
	"AB": true,
	"DT": true,
	"HA": true,
	"HD": true,
	"HP": true,
	"HS": true,
	"IE": true,
	"IN": true,
	"JJ": true,
	"KN": true,
	"NN": true,
	"PC": true,
	"PF": true, // ???
	"PL": true,
	"PM": true,
	"PN": true,
	"PP": true,
	"PS": true,
	"RG": true,
	"RO": true,
	"SN": true,
	"UO": true,
	"VB": true,
}

var langCodes = map[string]string{
	"swe": "sv-se",
	"sfi": "sv-fi",
	"nor": "nb-no",
	"nno": "nn-no",
	"eng": "en",
	"fin": "fi-fi",
	"ger": "de-de",
	"fre": "fr-fr",
	"rus": "ru-ru",
	"lat": "la",
	"ita": "it-it",
	"for": "foreign",
	"dan": "da-dk",
	"spa": "es-es",
}

func validPos(pos string) bool {
	if pos == "" {
		return true
	}
	_, ok := sucTags[pos]
	if ok {
		return true
	}
	return false
}

func mapLanguage(lang string) (string, error) {
	if lang == "" {
		return lang, nil
	}
	l, ok := langCodes[strings.ToLower(lang)]
	if ok {
		return l, nil
	}
	return lang, fmt.Errorf("couldn't map language <%v>", lang)
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

func main() {

	if len(os.Args) != 4 {
		fmt.Fprintln(os.Stderr, "<INPUT NST LEX FILE> <LEX2IPA MAPPER> <IPA2SAMPA MAPPER>")
		fmt.Fprintln(os.Stderr, "\tsample invokation:  go run convertNST2WS.go swe030224NST.pron.utf8 sv-se_nst-xsampa.tab sv-se_ws-sampa.tab ")
		return
	}

	nstFileName := os.Args[1]
	ssFileName1 := os.Args[2]
	ssFileName2 := os.Args[3]

	mapper, err := symbolset.LoadMapperFromFile("SAMPA", "SYMBOL", ssFileName1, ssFileName2)
	if err != nil {
		fmt.Fprintf(os.Stderr, "couldn't load mappers: %v\n", err)
		return
	}
	ssRuleTo := vrules.SymbolSetRule{mapper.SymbolSet2.To}

	nstFile, err := os.Open(nstFileName)
	defer nstFile.Close()
	if err != nil {
		fmt.Fprintf(os.Stderr, "couldn't open lexicon file: %v\n", err)
		return
	}

	nstFmt, err := line.NewNST()
	if err != nil {
		log.Fatal(err)
	}
	wsFmt, err := line.NewWS()
	if err != nil {
		log.Fatal(err)
	}

	nst := bufio.NewScanner(nstFile)
	n := 0
	for nst.Scan() {
		n++
		hasError := false
		if err := nst.Err(); err != nil {
			fmt.Fprintf(os.Stderr, "general error	failed reading line %v : %v\n", n, err)
			hasError = true
		}
		line := nst.Text()

		e, err := nstFmt.ParseToEntry(line)
		if err != nil {
			fmt.Fprintf(os.Stderr, "general error	failed to convert line %v to entry : %v\n", n, err)
			fmt.Fprintf(os.Stderr, "general error	failing line: %v\n", line)
			hasError = true
		}

		e.EntryStatus.Name = "imported"
		e.EntryStatus.Source = "nst"
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

		err = mapper.MapTranscriptions(&e)
		if err != nil {
			fmt.Fprintf(os.Stderr, "transcription error	failed to map transcription symbols : %v\n", err)
			hasError = true
		}

		if !hasError {
			for _, r := range ssRuleTo.Validate(e) {
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
