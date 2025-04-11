package line

import (
	"fmt"
	"testing"

	"github.com/stts-se/pronlex/lex"
)

//var fsExpField = "For field %v, expected: '%v' got: '%v'"
//var fsExp = "Expected: '%v' got: '%v'"

func checkBraxenResultField(t *testing.T, field string, x string, r string) {
	if x != r {
		t.Errorf(fsExpField, field, x, r)
	}
}

func checkBraxenResult(t *testing.T, x lex.Entry, r lex.Entry) {
	checkBraxenResultField(t, Orth.String(), x.Strn, r.Strn)
	checkBraxenResultField(t, Pos.String(), x.PartOfSpeech, r.PartOfSpeech)
	//checkBraxenResultField(t, Morph.String(), x.Morphology, r.Morphology)
	checkBraxenResultField(t, Lang.String(), x.Language, r.Language)

	if len(x.Transcriptions) != len(r.Transcriptions) {
		t.Errorf("Expected %v transcriptions, got %v", len(x.Transcriptions), len(r.Transcriptions))
	} else {
		for i, trans := range x.Transcriptions {
			transID := fmt.Sprintf("Trans%d", (i + 1))
			translangID := fmt.Sprintf("Translang%d", (i + 1))
			checkBraxenResultField(t, transID, trans.Strn, r.Transcriptions[i].Strn)
			checkBraxenResultField(t, translangID, trans.Language, r.Transcriptions[i].Language)
		}
	}
}

func Test_NewBraxen(t *testing.T) {
	_, err := NewBraxen()
	if err != nil {
		t.Errorf("didn't expect error here: %s", err)
	}
}

func Test_BraxenParse_01(t *testing.T) {
	nst, err := NewBraxen()
	if err != nil {
		t.Errorf("didn't expect error here: %s", err)
		return
	}

	input := "storstaden	s t \"u: r - s t ,a: . d ex n	NN UTR SIN DEF NOM	swe	-	-	-	-	-	-	-	-	-	-	-	-	0	-	-	-	-	-	-	-	-	-	480719"
	expect := lex.Entry{
		Strn:         "storstaden",
		PartOfSpeech: "NN UTR SIN DEF NOM",
		// PartOfSpeech: "NN",
		// Morphology:   "UTR SIN DEF NOM",
		Language: "swe",
		Transcriptions: []lex.Transcription{
			{
				Strn: "s t \"u: r - s t ,a: . d ex n",
				// Language: "swe",
			},
		},
	}
	result, _, err := nst.ParseToEntry(input)
	if err != nil {
		t.Errorf("didn't expect error here : %v", err)
	} else {
		checkBraxenResult(t, expect, result)
	}

}
