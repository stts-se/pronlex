package validators

import (
	"github.com/dlclark/regexp2"
	"github.com/stts-se/pronlex/symbolset"
	"github.com/stts-se/pronlex/validation"
	"github.com/stts-se/pronlex/validation/rules"
)

func newSvSeNstValidator(symbolset symbolset.SymbolSet) (validation.Validator, error) {
	primaryStressRe, err := rules.ProcessTransRe(symbolset, "\"")
	if err != nil {
		return validation.Validator{}, err
	}
	syllabicRe, err := rules.ProcessTransRe(symbolset, "^(\"\"|\"|%)? *(nonsyllabic +)*syllabic( +nonsyllabic)*( (.|-) (\"\"|\"|%)? *(nonsyllabic +)*syllabic( +nonsyllabic)*)*$")
	if err != nil {
		return validation.Validator{}, err
	}

	maxOneSyllabic, err := rules.ProcessTransRe(symbolset, "syllabic[^.+%\"-]*( +syllabic)")
	if err != nil {
		return validation.Validator{}, err
	}

	reFrom, err := regexp2.Compile("(.)\\1[+]\\1", regexp2.None)
	if err != nil {
		return validation.Validator{}, err
	}
	decomp2Orth := rules.Decomp2Orth{CompDelim: "+",
		AcceptEmptyDecomp: true,
		PreFilterWordPartString: func(s string) (string, error) {
			res, err := reFrom.Replace(s, "$1+$1", 0, -1)
			if err != nil {
				return s, err
			}
			return res, nil
		}}

	repeatedPhnRe, err := rules.ProcessTransRe(symbolset, "symbol( +[.~])? +\\1")
	if err != nil {
		return validation.Validator{}, err
	}

	var vali = validation.Validator{
		Name: symbolset.Name,
		Rules: []validation.Rule{
			rules.MustHaveTrans{},
			rules.NoEmptyTrans{},
			rules.RequiredTransRe{
				Name:    "primary_stress",
				Level:   "Fatal",
				Message: "Primary stress required",
				Re:      primaryStressRe,
			},
			rules.RequiredTransRe{
				Name:    "syllabic",
				Level:   "Format",
				Message: "Each syllable needs a syllabic phoneme",
				Re:      syllabicRe,
			},
			rules.IllegalTransRe{
				Name:    "MaxOneSyllabic",
				Level:   "Fatal",
				Message: "A syllable cannot contain more than one syllabic phoneme",
				Re:      maxOneSyllabic,
			},
			rules.IllegalTransRe{
				Name:    "repeated_phonemes",
				Level:   "Fatal",
				Message: "Repeated phonemes cannot be used within the same morpheme",
				Re:      repeatedPhnRe,
			},
			decomp2Orth,
			rules.SymbolSetRule{
				SymbolSet: symbolset,
			},
		}}
	return vali, nil
}