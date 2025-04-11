package line

import (
	"fmt"

	"github.com/stts-se/pronlex/lex"
)

// Braxen contains the line format used for Braxen lexicon data.
// Struct for package private usage.
// To create a new Braxen instance, use NewBraxen.
type Braxen struct {
	format Format
}

// Format is the line.Format instance used for line parsing inside of this parser
func (brax Braxen) Format() Format {
	return brax.format
}

// Parse is used for parsing input lines (calls underlying Format.Parse)
func (brax Braxen) Parse(line string) (map[Field]string, error) {
	return brax.format.Parse(line)
}

// ParseToEntry is used for parsing input lines (calls underlying Format.Parse).
// Orthography will be lower cased, but 2nd return argument is the input orthography with its original case
func (brax Braxen) ParseToEntry(line string) (lex.Entry, string, error) {
	fs, err := brax.format.Parse(line)
	if err != nil {
		return lex.Entry{}, "", err
	}

	// splitted := strings.SplitN(fs[Pos], "|", 2)
	// if len(splitted) == 2 {
	// 	fs[Pos] = splitted[0]
	// 	fs[Morph] = fs[Morph] + " " + splitted[1]
	// } else if len(splitted) == 1 {
	// } else {
	// 	panic("???")
	// }

	res := lex.Entry{
		Strn:         fs[Orth], //strings.ToLower(fs[Orth]),
		Language:     fs[Lang],
		PartOfSpeech: fs[Pos],
		// Morphology:     fs[Morph],
		// WordParts:      fs[WordParts],
		Transcriptions: getTranses(fs), // <-- func getTranses declared in nst.go
	}
	// if strings.HasPrefix(res.PartOfSpeech, "PM") {
	// 	res.Lemma = lex.Lemma{Strn: fs[Orth]}
	// }
	return res, fs[Orth], nil
}

// String is used to generate an output line from a set of fields (calls underlying Format.String)
func (brax Braxen) String(fields map[Field]string) (string, error) {
	return brax.format.String(fields)
}

// Entry2String is used to generate an output line from a lex.Entry (calls underlying Format.String)
func (brax Braxen) Entry2String(e lex.Entry) (string, error) {
	fs, err := brax.fields(e)
	if err != nil {
		return "", err
	}
	s, err := brax.format.String(fs)
	if err != nil {
		return "", err
	}
	return s, nil
}

func (brax Braxen) Header() string {
	return brax.format.Header()
}

func (brax Braxen) fields(e lex.Entry) (map[Field]string, error) {

	// Fields ID and LexiconID are database internal, and not processed here

	var fs = make(map[Field]string)
	fs[Orth] = e.Strn
	fs[Lang] = e.Language
	fs[WordParts] = e.WordParts

	// // PartOfSpeech => Pos + Morph
	// posMorph := strings.SplitN(e.PartOfSpeech, " ", 2)
	// switch len(posMorph) {
	// case 2:
	// 	fs[Pos] = posMorph[0]
	// 	fs[Morph] = posMorph[1]
	// case 1:
	// 	fs[Pos] = posMorph[0]
	// default:
	// 	return map[Field]string{}, fmt.Errorf("couldn't split db partofspeech into pos+morph: %s", e.PartOfSpeech)
	// }

	fs[Pos] = e.PartOfSpeech
	//fs[Morph] = e.Morphology

	for i, t := range e.Transcriptions {
		switch i {
		case 0:
			fs[Trans1] = t.Strn
			fs[Translang1] = t.Language
		// case 1:
		// 	fs[Trans2] = t.Strn
		// 	fs[Translang2] = t.Language
		// case 2:
		// 	fs[Trans3] = t.Strn
		// 	fs[Translang3] = t.Language
		// case 3:
		// 	fs[Trans4] = t.Strn
		// 	fs[Translang4] = t.Language
		default:
			return map[Field]string{}, fmt.Errorf("braxen line format can contain max one transcription, but found %v in: %v", len(e.Transcriptions), e)
		}
	}
	return fs, nil
}

// NewBraxen is used to create an instance of the Braxen parser
func NewBraxen() (Braxen, error) {
	tests := []FormatTest{
		{"storstaden	s t \"u: r - s t ,a: . d ex n	NN UTR SIN DEF NOM	swe	-	-	-	-	-	-	-	-	-	-	-	-	0	-	-	-	-	-	-	-	-	-	480719",
			map[Field]string{
				Orth:   "storstaden",
				Trans1: "s t \"u: r - s t ,a: . d ex n",
				Pos:    "NN UTR SIN DEF NOM",
				Lang:   "swe",
			},
			"storstaden	s t \"u: r - s t ,a: . d ex n	NN UTR SIN DEF NOM	swe																							",
		},
	}
	f, err := NewFormat(
		"Braxen",
		"\t",
		map[Field]int{
			Orth:   0,
			Trans1: 1,
			Pos:    2,
			Lang:   3,
		},
		27,
		tests,
	)
	if err != nil {
		return Braxen{}, err
	}
	return Braxen{f}, nil
}
