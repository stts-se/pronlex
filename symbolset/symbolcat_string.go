// Code generated by "stringer -type=SymbolCat"; DO NOT EDIT

package symbolset

import "fmt"

const _SymbolCat_name = "SyllabicNonSyllabicStressPhonemeDelimiterSyllableDelimiterMorphemeDelimiterCompoundDelimiterWordDelimiter"

var _SymbolCat_index = [...]uint8{0, 8, 19, 25, 41, 58, 75, 92, 105}

func (i SymbolCat) String() string {
	if i < 0 || i >= SymbolCat(len(_SymbolCat_index)-1) {
		return fmt.Sprintf("SymbolCat(%d)", i)
	}
	return _SymbolCat_name[_SymbolCat_index[i]:_SymbolCat_index[i+1]]
}
