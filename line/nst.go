package line

import (
	"fmt"
	"strings"

	"github.com/stts-se/pronlex/lex"
)

// NST contains the line format used for NST lexicon data.
// Struct for package private usage.
// To create a new NST instance, use NewNST.
type NST struct {
	format Format
}

// Format is the line.Format instance used for line parsing inside of this parser
func (nst NST) Format() Format {
	return nst.format
}

// Parse is used for parsing input lines (calls underlying Format.Parse)
func (nst NST) Parse(line string) (map[Field]string, error) {
	return nst.format.Parse(line)
}

// ParseToEntry is used for parsing input lines (calls underlying Format.Parse).
// Orthography will be lower cased, but 2nd return argument is the input orthography with its original case
func (nst NST) ParseToEntry(line string) (lex.Entry, string, error) {
	fs, err := nst.format.Parse(line)
	if err != nil {
		return lex.Entry{}, "", err
	}

	splitted := strings.SplitN(fs[Pos], "|", 2)
	if len(splitted) == 2 {
		fs[Pos] = splitted[0]
		fs[Morph] = fs[Morph] + " " + splitted[1]
	} else if len(splitted) == 1 {
	} else {
		panic("???")
	}

	res := lex.Entry{
		Strn:           strings.ToLower(fs[Orth]),
		Language:       fs[Lang],
		PartOfSpeech:   fs[Pos],
		Morphology:     fs[Morph],
		WordParts:      fs[WordParts],
		Transcriptions: getTranses(fs),
	}

	lemmaReading := strings.SplitN(fs[Lemma], "|", 2)
	lemma := ""
	reading := ""
	if len(lemmaReading) == 2 {
		lemma = lemmaReading[0]
		reading = lemmaReading[1]
	} else if len(lemmaReading) == 1 {
		lemma = lemmaReading[0]
	}
	paradigm := fs[Paradigm]
	lemmaStruct := lex.Lemma{Strn: lemma, Reading: reading, Paradigm: paradigm}

	if lemmaStruct.Strn != "" {
		res.Lemma = lemmaStruct
	}

	return res, fs[Orth], nil
}

func appendTrans(ts []lex.Transcription, trans string, lang string, source string) []lex.Transcription {
	if strings.TrimSpace(trans) == "" {
		return ts
	}
	ts = append(ts, lex.Transcription{Strn: trans, Language: lang, Sources: []string{source}})
	return ts
}

func getTranses(fs map[Field]string) []lex.Transcription {
	var res []lex.Transcription
	statusSource := fs[StatusSource]
	res = appendTrans(res, fs[Trans1], fs[Translang1], statusSource)
	res = appendTrans(res, fs[Trans2], fs[Translang2], statusSource)
	res = appendTrans(res, fs[Trans3], fs[Translang3], statusSource)
	res = appendTrans(res, fs[Trans4], fs[Translang4], statusSource)
	return res
}

// String is used to generate an output line from a set of fields (calls underlying Format.String)
func (nst NST) String(fields map[Field]string) (string, error) {
	return nst.format.String(fields)
}

// Entry2String is used to generate an output line from a lex.Entry (calls underlying Format.String)
func (nst NST) Entry2String(e lex.Entry) (string, error) {
	fs, err := nst.fields(e)
	if err != nil {
		return "", err
	}
	s, err := nst.format.String(fs)
	if err != nil {
		return "", err
	}
	return s, nil
}

func (nst NST) Header() string {
	return nst.format.Header()
}

func (nst NST) fields(e lex.Entry) (map[Field]string, error) {

	// Fields ID and LexiconID are database internal  and not processed here

	var fs = make(map[Field]string)
	fs[Orth] = e.Strn
	fs[Lang] = e.Language
	fs[WordParts] = e.WordParts

	// PartOfSpeech => Pos + Morph
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
	fs[Morph] = e.Morphology

	// Lemma
	if e.Lemma.Reading != "" {
		fs[Lemma] = e.Lemma.Strn + "|" + e.Lemma.Reading
	} else {
		fs[Lemma] = e.Lemma.Strn
	}
	if e.Lemma.Reading != "" {
		fs[Lemma] = e.Lemma.Strn + "|" + e.Lemma.Reading
	} else {
		fs[Lemma] = e.Lemma.Strn
	}
	fs[Paradigm] = e.Lemma.Paradigm

	for i, t := range e.Transcriptions {
		switch i {
		case 0:
			fs[Trans1] = t.Strn
			fs[Translang1] = t.Language
		case 1:
			fs[Trans2] = t.Strn
			fs[Translang2] = t.Language
		case 2:
			fs[Trans3] = t.Strn
			fs[Translang3] = t.Language
		case 3:
			fs[Trans4] = t.Strn
			fs[Translang4] = t.Language
		default:
			return map[Field]string{}, fmt.Errorf("nst line format can contain max 4 transcriptions, but found %v in: %v", len(e.Transcriptions), e)
		}
	}
	return fs, nil
}

// NewNST is used to create an instance of the NST parser
func NewNST() (NST, error) {
	tests := []FormatTest{
		{"storstaden;NN;SIN|DEF|NOM|UTR;stor+staden;JJ+NN;LEX|INFL;SWE;;;;;\"\"stu:$%s`t`A:$den;1;STD;SWE;;;;;;;;;;;;;;18174;enter_se|inflector;;INFLECTED;storstad|95522;s111n, a->ä, stad;s111;;;;;;;;;;;;;storstaden;;;88748",
			map[Field]string{
				Orth:       "storstaden",
				Pos:        "NN",
				Morph:      "SIN|DEF|NOM|UTR",
				WordParts:  "stor+staden",
				Lang:       "SWE",
				Trans1:     "\"\"stu:$%s`t`A:$den",
				Translang1: "SWE",
				Trans2:     "",
				Translang2: "",
				Trans3:     "",
				Translang3: "",
				Trans4:     "",
				Translang4: "",
				Lemma:      "storstad|95522",
				Paradigm:   "s111n, a->ä, stad",
			},
			"storstaden;NN;SIN|DEF|NOM|UTR;stor+staden;;;SWE;;;;;\"\"stu:$%s`t`A:$den;;;SWE;;;;;;;;;;;;;;;;;;storstad|95522;s111n, a->ä, stad;;;;;;;;;;;;;;;;;",
		},
	}
	f, err := NewFormat(
		"NST",
		";",
		map[Field]int{
			Orth:       0,
			Pos:        1,
			Morph:      2,
			WordParts:  3,
			Lang:       6,
			Trans1:     11,
			Translang1: 14,
			Trans2:     15,
			Translang2: 18,
			Trans3:     19,
			Translang3: 22,
			Trans4:     23,
			Translang4: 26,
			Lemma:      32,
			Paradigm:   33,
		},
		51,
		tests,
	)
	if err != nil {
		return NST{}, err
	}
	return NST{f}, nil
}
