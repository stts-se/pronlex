package validators

import (
	"github.com/stts-se/pronlex/validation"
	"github.com/stts-se/pronlex/validation/rules"
	"github.com/stts-se/symbolset"
)

func newEnUsCmuValidator(symbolset symbolset.SymbolSet) (validation.Validator, error) {
	exactOnePrimStressRe, err := rules.ProcessTransRe(symbolset, "^[^']*'[^']*$")
	if err != nil {
		return validation.Validator{}, err
	}
	maxOneSecStressRe, err := rules.ProcessTransRe(symbolset, "%.*%")
	if err != nil {
		return validation.Validator{}, err
	}

	var vali = validation.Validator{
		Name: symbolset.Name,
		Rules: []validation.Rule{
			rules.MustHaveTrans{},
			rules.NoEmptyTrans{},
			rules.RequiredTransRe{
				NameStr:  "primary_stress",
				LevelStr: "Format",
				Message:  "Each trans should have one primary stress",
				Re:       exactOnePrimStressRe,
			},
			rules.IllegalTransRe{
				NameStr:  "secondary_stress",
				LevelStr: "Format",
				Message:  "Each trans can have max one secondary stress",
				Re:       maxOneSecStressRe,
			},
			rules.SymbolSetRule{
				SymbolSet: symbolset,
			},
		}}
	return vali, nil
}
