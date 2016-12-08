package symbolset2

import "testing"

var fsExpTrans = "Expected: /%v/ got: /%v/"

func testSymbolSetConvertToIPA(t *testing.T, ss SymbolSet, input string, expect string) {
	result, err := ss.ConvertToIPA(input)
	if err != nil {
		t.Errorf("ConvertToIPA() error here; input=%s, expect=%s : %v", input, expect, err)
		return
	} else if result != expect {
		t.Errorf(fsExpTrans, expect, result)
	}
}

func testSymbolSetConvertFromIPA(t *testing.T, ss SymbolSet, input string, expect string) {
	result, err := ss.ConvertFromIPA(input)
	if err != nil {
		t.Errorf("ConvertFromIPA() error here; input=%s, expect=%s : %v", input, expect, err)
		return
	} else if result != expect {
		t.Errorf(fsExpTrans, expect, result)
	}
}

func Test_NewSymbolSet_WithoutPhonemeDelimiter(t *testing.T) {
	name := "ss"
	symbols := []Symbol{
		Symbol{"a", Syllabic, "", IPASymbol{"", ""}},
		Symbol{"t", NonSyllabic, "", IPASymbol{"", ""}},
	}
	_, err := NewSymbolSet(name, symbols)
	if err == nil {
		t.Errorf("NewSymbolSet() should fail if no phoneme delimiter is defined")
	}
}

func Test_NewSymbolSet_FailIfInputContainsDuplicates(t *testing.T) {
	name := "ss"
	symbols := []Symbol{
		Symbol{"a", Syllabic, "", IPASymbol{"", ""}},
		Symbol{"a", NonSyllabic, "", IPASymbol{"", ""}},
		Symbol{"t", NonSyllabic, "", IPASymbol{"", ""}},
		Symbol{" ", PhonemeDelimiter, "phn delim", IPASymbol{"", ""}},
	}
	_, err := NewSymbolSet(name, symbols)
	if err == nil {
		t.Errorf("NewSymbolSet() expected error here")
	}
}

func Test_NewSymbolSet_FailOnIncorrectIPAUnicode(t *testing.T) {
	name := "ss"
	symbols := []Symbol{
		Symbol{"a", Syllabic, "", IPASymbol{"a", "U+0061"}},
		Symbol{"t", NonSyllabic, "", IPASymbol{"t", "U+0074"}},
		Symbol{"A:", Syllabic, "", IPASymbol{"ɑː", "U+0251:"}},
		Symbol{" ", PhonemeDelimiter, "phn delim", IPASymbol{"", ""}},
	}
	_, err := NewSymbolSet(name, symbols)
	if err == nil {
		t.Errorf("NewSymbolSet() expected ipa/unicode error here")
	}
}

func Test_SplitTranscription_Normal1(t *testing.T) {
	name := "ss"
	symbols := []Symbol{
		Symbol{"a", Syllabic, "", IPASymbol{"", ""}},
		Symbol{"t", NonSyllabic, "", IPASymbol{"", ""}},
		Symbol{"s", NonSyllabic, "", IPASymbol{"", ""}},
		Symbol{"t_s", NonSyllabic, "", IPASymbol{"", ""}},
		Symbol{" ", PhonemeDelimiter, "phn delim", IPASymbol{"", ""}},
	}
	ss, err := NewSymbolSet(name, symbols)
	if err != nil {
		t.Errorf("SplitTranscription() didn't expect error here : %v", err)
		return
	}

	input := "a t s t_s s"
	expect := []string{"a", "t", "s", "t_s", "s"}
	result, err := ss.SplitTranscription(input)
	if err != nil {
		t.Errorf("SplitTranscription() didn't expect error here")
	}
	testEqStrings(t, expect, result)
}

func Test_SplitTranscription_EmptyPhonemeDelmiter1(t *testing.T) {
	name := "ss"
	symbols := []Symbol{
		Symbol{"a", Syllabic, "", IPASymbol{"", ""}},
		Symbol{"t", NonSyllabic, "", IPASymbol{"", ""}},
		Symbol{"s", NonSyllabic, "", IPASymbol{"", ""}},
		Symbol{"t_s", NonSyllabic, "", IPASymbol{"", ""}},
		Symbol{"", PhonemeDelimiter, "", IPASymbol{"", ""}},
	}
	ss, err := NewSymbolSet(name, symbols)
	if err != nil {
		t.Errorf("SplitTranscription() didn't expect error here")
		return
	}

	input := "atst_ss"
	expect := []string{"a", "t", "s", "t_s", "s"}
	result, err := ss.SplitTranscription(input)
	if err != nil {
		t.Errorf("SplitTranscription() didn't expect error here")
	}
	testEqStrings(t, expect, result)
}

func Test_SplitTranscription_FailWithUnknownSymbols_EmptyDelim(t *testing.T) {
	name := "sampa"
	symbols := []Symbol{
		Symbol{"a", Syllabic, "", IPASymbol{"", ""}},
		Symbol{"b", NonSyllabic, "", IPASymbol{"", ""}},
		Symbol{"N", NonSyllabic, "", IPASymbol{"", ""}},
		Symbol{"", PhonemeDelimiter, "", IPASymbol{"", ""}},
		Symbol{".", SyllableDelimiter, "", IPASymbol{"", ""}},
		Symbol{"\"", Stress, "", IPASymbol{"", ""}},
		Symbol{"\"\"", Stress, "", IPASymbol{"", ""}},
	}
	ss, err := NewSymbolSet(name, symbols)
	if err != nil {
		t.Errorf("SplitTranscription() didn't expect error here : %v", err)
		return
	}
	input := "\"\"baN.ka"
	//expect := []string{"\"\"", "b", "a", "N", ".", "k", "a"}
	result, err := ss.SplitTranscription(input)
	if err == nil {
		t.Errorf("SplitTranscription() expected error here, but got %s", result)
	}
}

func Test_SplitTranscription_NoFailWithUnknownSymbols_NonEmptyDelim(t *testing.T) {
	name := "sampa"
	symbols := []Symbol{
		Symbol{"a", Syllabic, "", IPASymbol{"", ""}},
		Symbol{"b", NonSyllabic, "", IPASymbol{"", ""}},
		Symbol{"N", NonSyllabic, "", IPASymbol{"", ""}},
		Symbol{" ", PhonemeDelimiter, "", IPASymbol{"", ""}},
		Symbol{".", SyllableDelimiter, "", IPASymbol{"", ""}},
		Symbol{"\"", Stress, "", IPASymbol{"", ""}},
		Symbol{"\"\"", Stress, "", IPASymbol{"", ""}},
	}
	ss, err := NewSymbolSet(name, symbols)
	if err != nil {
		t.Errorf("SplitTranscription() didn't expect error here : %v", err)
		return
	}
	input := "\"\" b a N . k a"
	expect := []string{"\"\"", "b", "a", "N", ".", "k", "a"}
	result, err := ss.SplitTranscription(input)
	if err != nil {
		t.Errorf("SplitTranscription() didn't expect error here : %v", err)
	}
	testEqStrings(t, expect, result)
}

func Test_ValidSymbol1(t *testing.T) {
	name := "sampa"
	symbols := []Symbol{
		Symbol{"a", Syllabic, "", IPASymbol{"", ""}},
		Symbol{"b", NonSyllabic, "", IPASymbol{"", ""}},
		Symbol{"N", NonSyllabic, "", IPASymbol{"", ""}},
		Symbol{" ", PhonemeDelimiter, "", IPASymbol{"", ""}},
		Symbol{".", SyllableDelimiter, "", IPASymbol{"", ""}},
		Symbol{"\"", Stress, "", IPASymbol{"", ""}},
		Symbol{"\"\"", Stress, "", IPASymbol{"", ""}},
	}
	ss, err := NewSymbolSet(name, symbols)
	if err != nil {
		t.Errorf("didn't expect error here : %v", err)
		return
	}

	var phn = ""

	phn = "a"
	if !ss.ValidSymbol(phn) {
		t.Errorf("expected phoneme %v to be valid", phn)
	}

	phn = "."
	if !ss.ValidSymbol(phn) {
		t.Errorf("expected phoneme %v to be valid", phn)
	}

	phn = "x"
	if ss.ValidSymbol(phn) {
		t.Errorf("expected phoneme %v to be invalid", phn)
	}

}

func Test_ConvertToIPA(t *testing.T) {
	symbols := []Symbol{
		Symbol{"a", Syllabic, "", IPASymbol{"a", "U+0061"}},
		Symbol{"b", NonSyllabic, "", IPASymbol{"b", "U+0062"}},
		Symbol{"r", NonSyllabic, "", IPASymbol{"r", "U+0072"}},
		Symbol{"k", NonSyllabic, "", IPASymbol{"k", "U+006B"}},
		Symbol{"", PhonemeDelimiter, "", IPASymbol{"", ""}},
		Symbol{"A:", Syllabic, "", IPASymbol{"ɑː", "U+0251U+02D0"}},
		Symbol{"$", SyllableDelimiter, "", IPASymbol{".", "U+002E"}},
		Symbol{"\"", Stress, "", IPASymbol{"\u02C8", "U+02C8"}},
		Symbol{"\"\"", Stress, "", IPASymbol{"\u02C8\u0300", "U+02C8U+0300"}},
	}
	ss, err := NewSymbolSet("sampa", symbols)
	if err != nil {
		t.Errorf("NewSymbolSet() didn't expect error here : %v", err)
		return
	}

	// --
	input := "\"\"brA:$ka"
	expect := "\u02C8brɑ\u0300ː.ka"
	result, err := ss.ConvertToIPA(input)
	if err != nil {
		t.Errorf("ConvertToIPA() didn't expect error here : %v", err)
	}
	if result != expect {
		t.Errorf(fsExpTrans, expect, result)
	}

	// --
	input = "\"brA:$ka"
	expect = "\u02C8brɑː.ka"
	result, err = ss.ConvertToIPA(input)
	if err != nil {
		t.Errorf("ConvertToIPA() didn't expect error here : %v", err)
	}
	if result != expect {
		t.Errorf(fsExpTrans, expect, result)
	}
}

func Test_ConvertFromIPA(t *testing.T) {
	symbols := []Symbol{
		Symbol{"a", Syllabic, "", IPASymbol{"a", "U+0061"}},
		Symbol{"b", NonSyllabic, "", IPASymbol{"b", "U+0062"}},
		Symbol{"r", NonSyllabic, "", IPASymbol{"r", "U+0072"}},
		Symbol{"k", NonSyllabic, "", IPASymbol{"k", "U+006B"}},
		Symbol{"", PhonemeDelimiter, "", IPASymbol{"", ""}},
		Symbol{"A:", Syllabic, "", IPASymbol{"ɑː", "U+0251U+02D0"}},
		Symbol{"$", SyllableDelimiter, "", IPASymbol{".", "U+002E"}},
		Symbol{"\"", Stress, "", IPASymbol{"\u02C8", "U+02C8"}},
		Symbol{"\"\"", Stress, "", IPASymbol{"\u02C8\u0300", "U+02C8U+0300"}},
	}
	ss, err := NewSymbolSet("sampa", symbols)
	if err != nil {
		t.Errorf("NewSymbolSet() didn't expect error here : %v", err)
		return
	}

	// --
	input := "\u02C8brɑ\u0300ː.ka"
	expect := "\"\"brA:$ka"
	result, err := ss.ConvertFromIPA(input)
	if err != nil {
		t.Errorf("ConvertFromIPA() didn't expect error here : %v", err)
	}
	if result != expect {
		t.Errorf(fsExpTrans, expect, result)
	}

	// --
	input = "\u02C8brɑː.ka"
	expect = "\"brA:$ka"
	result, err = ss.ConvertFromIPA(input)
	if err != nil {
		t.Errorf("ConvertFromIPA() didn't expect error here : %v", err)
	}
	if result != expect {
		t.Errorf(fsExpTrans, expect, result)
	}
}

func stringSliceContains(slice []string, s string) bool {
	for _, sl := range slice {
		if sl == s {
			return true
		}
	}
	return false
}

func Test_LoadSymbolSetsFromDir(t *testing.T) {
	symbolsets, err := LoadSymbolSetsFromDir("./static")
	if err != nil {
		t.Errorf("LoadSymbolSetsFromDir() didn't expect error here : %v", err)
		return
	}
	var ssNames []string
	for _, ss := range symbolsets {
		ssNames = append(ssNames, ss.Name)
	}
	if len(symbolsets) != 7 {
		t.Errorf("Expected 7 symbol sets in folder ./static, found %d", len(symbolsets))
	}
	if !stringSliceContains(ssNames, "sv-se_nst-xsampa") {
		t.Errorf("Expected %s in symbolsets. Found: %v", "sv-se_nst-xsampa", ssNames)
	}
}

func Test_NewSymbolSet_WithCorrectInput1(t *testing.T) {
	symbols := []Symbol{
		Symbol{"a", Syllabic, "", IPASymbol{"A", "U+0041"}},
		Symbol{"p", NonSyllabic, "", IPASymbol{"P", "U+0050"}},
		Symbol{" ", PhonemeDelimiter, "", IPASymbol{" ", "U+0020"}},
	}
	_, err := NewSymbolSet("test", symbols)
	if err != nil {
		t.Errorf("NewSymbolSet() didn't expect error here : %v", err)
	}
}
func Test_loadSymbolSet_NST2IPA_SV(t *testing.T) {
	name := "NST-XSAMPA"
	fName := "static/sv-se_nst-xsampa.tab"
	ss, err := loadSymbolSet0(name, fName)
	if err != nil {
		t.Errorf("loadSymbolSet0() didn't expect error here : %v", err)
		return
	}
	testSymbolSetConvertToIPA(t, ss, "\"bOt`", "\u02C8bɔʈ")
	testSymbolSetConvertToIPA(t, ss, "\"ku0rds", "\u02C8kɵrds")
	testSymbolSetConvertToIPA(t, ss, "\"\"ku0$d@", "\u02C8kɵ\u0300.də")
}

func Test_MapTranscription_Sampa2Ipa_Simple(t *testing.T) {
	symbols := []Symbol{
		Symbol{"a", Syllabic, "", IPASymbol{"a", "U+0061"}},
		Symbol{"p", NonSyllabic, "", IPASymbol{"p", "U+0070"}},
		Symbol{"", PhonemeDelimiter, "", IPASymbol{"", ""}},
		Symbol{"$", SyllableDelimiter, "", IPASymbol{".", "U+002E"}},
	}
	ss, err := NewSymbolSet("test", symbols)
	if err != nil {
		t.Errorf("NewSymbolSet() didn't expect error here : %v", err)
		return
	}
	input := "pa$pa"
	expect := "pa.pa"
	result, err := ss.ConvertToIPA(input)
	if err != nil {
		t.Errorf("ConvertToIPA() didn't expect error here : %v", err)
		return
	}
	if result != expect {
		t.Errorf(fsExpTrans, expect, result)
	}
}

func Test_loadSymbolSet_WS2IPA(t *testing.T) {
	name := "WS-SAMPA"
	fName := "static/sv-se_ws-sampa.tab"
	ss, err := loadSymbolSet0(name, fName)
	if err != nil {
		t.Errorf("LoadSymbolSet() error here : %v", err)
	}
	testSymbolSetConvertToIPA(t, ss, "\" b O rt", "\u02C8bɔʈ")
	testSymbolSetConvertToIPA(t, ss, "\" k u0 r d s", "\u02C8kɵrds")
}

func Test_loadSymbolSet_CMU2IPA(t *testing.T) {
	name := "CMU"
	fName := "static/en-us_cmu.tab"
	ss, err := loadSymbolSet0(name, fName)
	if err != nil {
		t.Errorf("LoadSymbolSet() error here : %v", err)
	}
	testSymbolSetConvertToIPA(t, ss, "AX $ B AW1 T", "ə.\u02C8ba⁀ʊt")
}

func Test_loadSymbolSet_MARY2IPA(t *testing.T) {
	name := "en-us_sampa_mary"
	fName := "static/en-us_sampa_mary.tab"
	ss, err := loadSymbolSet0(name, fName)
	if err != nil {
		t.Errorf("LoadSymbolSet() error here : %v", err)
	}
	testSymbolSetConvertToIPA(t, ss, "@ - ' b aU t", "ə.\u02C8ba⁀ʊt")
}

func Test_loadSymbolSet_NST2IPA_NB(t *testing.T) {
	name := "nb-no_nst-xsampa"
	fName := "static/nb-no_nst-xsampa.tab"
	ss, err := loadSymbolSet0(name, fName)
	if err != nil {
		t.Errorf("LoadSymbolSet() error here : %v", err)
	}
	if ss.Type != SAMPA {
		t.Errorf("Expected symbol set type %#v, got %#v", SAMPA.String(), ss.Type.String())
		return
	}
	testSymbolSetConvertToIPA(t, ss, "\"A:$bl@s", "\u02C8ɑː.bləs")
	testSymbolSetConvertToIPA(t, ss, "\"tSE$kIsk", "\u02C8tʃɛ.kɪsk")
	testSymbolSetConvertToIPA(t, ss, "\"\"b9$n@r", "\u02C8bœ\u0300.nər")
	testSymbolSetConvertToIPA(t, ss, "\"b9$n@r", "\u02C8bœ.nər")
}

func Test_loadSymbolSet_IPA2WS(t *testing.T) {
	name := "WS-SAMPA"
	fName := "static/sv-se_ws-sampa.tab"
	ss, err := loadSymbolSet0(name, fName)
	if err != nil {
		t.Errorf("LoadSymbolSet() didn't expect error here : %v", err)
	}
	if ss.Type != SAMPA {
		t.Errorf("Expected symbol set type %#v, got %#v", SAMPA.String(), ss.Type.String())
		return
	}
	testSymbolSetConvertFromIPA(t, ss, "\u02C8bɔʈ", "\" b O rt")
	testSymbolSetConvertFromIPA(t, ss, "\u02C8kɵrds", "\" k u0 r d s")
}

func Test_loadSymbolSet_IPA2MARY(t *testing.T) {
	name := "sampa-mary"
	fName := "static/en-us_sampa_mary.tab"
	ss, err := loadSymbolSet0(name, fName)
	if err != nil {
		t.Errorf("LoadSymbolSet() didn't expect error here : %v", err)
		return
	}
	if ss.Type != SAMPA {
		t.Errorf("Expected symbol set type %#v, got %#v", SAMPA.String(), ss.Type.String())
		return
	}

	testSymbolSetConvertFromIPA(t, ss, "ə.\u02C8ba⁀ʊt", "@ - ' b aU t")
}

func Test_loadSymbolSet_IPA2SAMPA(t *testing.T) {
	name := "ws-xsampa"
	fName := "static/sv-se_ws-sampa.tab"
	ss, err := loadSymbolSet0(name, fName)
	if err != nil {
		t.Errorf("LoadSymbolSet() didn't expect error here : %v", err)
		return
	}
	if ss.Type != SAMPA {
		t.Errorf("Expected symbol set type %#v, got %#v", SAMPA.String(), ss.Type.String())
		return
	}

	testSymbolSetConvertFromIPA(t, ss, "\u02C8kaj.rʊ", "\" k a j . r U")
	testSymbolSetConvertFromIPA(t, ss, "be.\u02C8liːn", "b e . \" l i: n")
}

func Test_loadSymbolSet_IPA2CMU(t *testing.T) {
	name := "en-us_cmu"
	fName := "static/en-us_cmu.tab"
	ss, err := loadSymbolSet0(name, fName)
	if err != nil {
		t.Errorf("LoadSymbolSet() didn't expect error here : %v", err)
		return
	}
	if ss.Type != CMU {
		t.Errorf("Expected symbol set type %#v, got %#v", CMU.String(), ss.Type.String())
		return
	}

	testSymbolSetConvertFromIPA(t, ss, "ə.\u02C8ba⁀ʊt", "AX $ B AW1 T")
	testSymbolSetConvertFromIPA(t, ss, "ʌ.\u02C8ba⁀ʊt", "AH $ B AW1 T")
}

func Test_loadSymbolSet_IPA2NST_NB(t *testing.T) {
	name := "NST-XSAMPA"
	fName := "static/nb-no_nst-xsampa.tab"
	ss, err := loadSymbolSet0(name, fName)
	if err != nil {
		t.Errorf("LoadSymbolSet() didn't expect error here : %v", err)
	}
	testSymbolSetConvertFromIPA(t, ss, "\u02C8ɑː.bləs", "\"A:$bl@s")
	testSymbolSetConvertFromIPA(t, ss, "\u02C8tʃɛ.kɪsk", "\"tSE$kIsk")
	testSymbolSetConvertFromIPA(t, ss, "\u02C8bœ\u0300.nər", "\"\"b9$n@r")
	testSymbolSetConvertFromIPA(t, ss, "\u02C8bœ.nər", "\"b9$n@r")
}

func Test_NewSymbolSet_IPADuplicates_ConvertToIPA(t *testing.T) {
	symbols := []Symbol{
		Symbol{"i", Syllabic, "", IPASymbol{"I", "U+0049"}},
		Symbol{"i3", Syllabic, "", IPASymbol{"I", "U+0049"}},
		Symbol{"p", NonSyllabic, "", IPASymbol{"P", "U+0050"}},
		Symbol{" ", PhonemeDelimiter, "", IPASymbol{" ", "U+0020"}},
	}
	ss, err := NewSymbolSet("test", symbols)
	if err != nil {
		t.Errorf("NewSymbolSet() didn't expect error here : %v", err)
		return
	}
	testSymbolSetConvertToIPA(t, ss, "i3 p", "I P")
	testSymbolSetConvertToIPA(t, ss, "i p", "I P")
}

func Test_NewSymbolSet_IPADuplicates_ConvertFromIPA(t *testing.T) {
	symbols := []Symbol{
		Symbol{"i", Syllabic, "", IPASymbol{"I", "U+0049"}},
		Symbol{"i3", Syllabic, "", IPASymbol{"I", "U+0049"}},
		Symbol{"p", NonSyllabic, "", IPASymbol{"P", "U+0050"}},
		Symbol{" ", PhonemeDelimiter, "", IPASymbol{"", ""}},
	}
	ss, err := NewSymbolSet("test", symbols)
	if err != nil {
		t.Errorf("NewSymbolSet() didn't expect error here : %v", err)
		return
	}
	testSymbolSetConvertFromIPA(t, ss, "IP", "i p")
	testSymbolSetConvertFromIPA(t, ss, "IP", "i p")
}

func Test_NewSymbolSet_FailIfLacksPhonemeDelimiter(t *testing.T) {
	symbols := []Symbol{
		Symbol{"a", Syllabic, "", IPASymbol{"A", "U+0041"}},
		Symbol{"p", NonSyllabic, "", IPASymbol{"P", "U+0050"}},
		Symbol{" ", NonSyllabic, "", IPASymbol{" ", "U+0020"}},
	}
	_, err := NewSymbolSet("test", symbols)
	if err == nil {
		t.Errorf("NewSymbolSet() expected error here")
	}
}

func Test_ConvertToIPA_Sampa2Ipa_Simple(t *testing.T) {
	symbols := []Symbol{
		Symbol{"a", Syllabic, "", IPASymbol{"a", "U+0061"}},
		Symbol{"p", NonSyllabic, "", IPASymbol{"p", "U+0070"}},
		Symbol{"", PhonemeDelimiter, "", IPASymbol{"", ""}},
		Symbol{"$", SyllableDelimiter, "", IPASymbol{".", "U+002E"}},
	}
	ss, err := NewSymbolSet("test", symbols)
	if err != nil {
		t.Errorf("ConvertToIPA() didn't expect error here : %v", err)
		return
	}
	input := "pa$pa"
	expect := "pa.pa"
	result, err := ss.ConvertToIPA(input)
	if err != nil {
		t.Errorf("ConvertToIPA() didn't expect error here : %v", err)
	}
	if result != expect {
		t.Errorf(fsExpTrans, expect, result)
	}
}

func Test_ConvertToIPA_Sampa2Ipa_WithSwedishStress_1(t *testing.T) {
	symbols := []Symbol{
		Symbol{"a", Syllabic, "", IPASymbol{"a", "U+0061"}},
		Symbol{"p", NonSyllabic, "", IPASymbol{"p", "U+0070"}},
		Symbol{"", PhonemeDelimiter, "", IPASymbol{"", ""}},
		Symbol{"$", SyllableDelimiter, "", IPASymbol{".", "U+002E"}},
		Symbol{"\"", Stress, "", IPASymbol{"\u02C8", "U+02C8"}},
		Symbol{"\"\"", Stress, "", IPASymbol{"\u02C8\u0300", "U+02C8U+0300"}},
	}
	ss, err := NewSymbolSet("test", symbols)
	if err != nil {
		t.Errorf("ConvertToIPA() didn't expect error here : %v", err)
		return
	}
	input := "\"\"pa$pa"
	expect := "\u02C8pa\u0300.pa"
	result, err := ss.ConvertToIPA(input)
	if err != nil {
		t.Errorf("ConvertToIPA() didn't expect error here : %v", err)
	}
	if result != expect {
		t.Errorf(fsExpTrans, expect, result)
	}
}

func Test_ConvertToIPA_Sampa2Ipa_WithSwedishStress_2(t *testing.T) {
	symbols := []Symbol{
		Symbol{"a", Syllabic, "", IPASymbol{"a", "U+0061"}},
		Symbol{"b", NonSyllabic, "", IPASymbol{"b", "U+0062"}},
		Symbol{"r", NonSyllabic, "", IPASymbol{"r", "U+0072"}},
		Symbol{"k", NonSyllabic, "", IPASymbol{"k", "U+006B"}},
		Symbol{"", PhonemeDelimiter, "", IPASymbol{"", ""}},
		Symbol{"A:", Syllabic, "", IPASymbol{"ɑː", "U+0251U+02D0"}},
		Symbol{"$", SyllableDelimiter, "", IPASymbol{".", "U+002E"}},
		Symbol{"\"", Stress, "", IPASymbol{"\u02C8", "U+02C8"}},
		Symbol{"\"\"", Stress, "", IPASymbol{"\u02C8\u0300", "U+02C8U+0300"}},
	}
	ss, err := NewSymbolSet("test", symbols)
	if err != nil {
		t.Errorf("ConvertToIPA() didn't expect error here : %v", err)
		return
	}
	input := "\"\"brA:$ka"
	expect := "\u02C8brɑ\u0300ː.ka"
	result, err := ss.ConvertToIPA(input)
	if err != nil {
		t.Errorf("ConvertToIPA() didn't expect error here : %v", err)
	}
	if result != expect {
		t.Errorf(fsExpTrans, expect, result)
	}
}

func Test_ConvertToIPA_FailWithUnknownSymbols_NonEmptyDelim(t *testing.T) {
	symbols := []Symbol{
		Symbol{"a", Syllabic, "", IPASymbol{"a", "U+0061"}},
		Symbol{"b", NonSyllabic, "", IPASymbol{"b", "U+0062"}},
		Symbol{"ŋ", NonSyllabic, "", IPASymbol{"N", "U+004E"}},
		Symbol{" ", PhonemeDelimiter, "", IPASymbol{"", ""}},
		Symbol{".", SyllableDelimiter, "", IPASymbol{"$", "U+0024"}},
		Symbol{"\"", Stress, "", IPASymbol{"\u02C8", "U+02C8"}},
		Symbol{"\"\"", Stress, "", IPASymbol{"\u02C8\u0300", "U+02C8U+0300"}},
	}
	ss, err := NewSymbolSet("test", symbols)
	if err != nil {
		t.Errorf("ConvertToIPA() didn't expect error here : %v", err)
		return
	}
	input := "\"\" b a ŋ . k a"
	result, err := ss.ConvertToIPA(input)
	if err == nil {
		t.Errorf("NewSymbolSet() expected error here, but got %s", result)
	}
}

func Test_Get(t *testing.T) {
	symbols := []Symbol{
		Symbol{"a", Syllabic, "", IPASymbol{"A", "U+0041"}},
		Symbol{"P", NonSyllabic, "", IPASymbol{"p", "U+0070"}},
		Symbol{" ", PhonemeDelimiter, "", IPASymbol{" ", "U+0020"}},
	}
	ss, err := NewSymbolSet("test", symbols)
	if err != nil {
		t.Errorf("test didn't expect error here : %v", err)
		return
	}

	// --
	{
		res, err := ss.Get("P")
		if err != nil {
			t.Errorf("didn't expect error here : %v", err)
			return
		}
		if res.IPA.String != "p" {
			t.Errorf(fsExpTrans, "p", res.IPA.String)
		}
		if res.String != "P" {
			t.Errorf(fsExpTrans, "P", res.String)
		}
	}

	// --
	{
		_, err := ss.Get("A")
		if err == nil {
			t.Errorf("expected error here for unknown input symbol : %v", "A")
			return
		}
	}

}

func Test_getFromIPA(t *testing.T) {
	symbols := []Symbol{
		Symbol{"a", Syllabic, "", IPASymbol{"A", "U+0041"}},
		Symbol{"P", NonSyllabic, "", IPASymbol{"p", "U+0070"}},
		Symbol{" ", PhonemeDelimiter, "", IPASymbol{" ", "U+0020"}},
	}
	ss, err := NewSymbolSet("test", symbols)
	if err != nil {
		t.Errorf("test didn't expect error here : %v", err)
		return
	}

	// --
	{
		res, err := ss.getFromIPA("p")
		if err != nil {
			t.Errorf("didn't expect error here : %v", err)
			return
		}
		if res.IPA.String != "p" {
			t.Errorf(fsExpTrans, "p", res.IPA.String)
		}
		if res.String != "P" {
			t.Errorf(fsExpTrans, "P", res.String)
		}
	}

	// --
	{
		_, err := ss.getFromIPA("a")
		if err == nil {
			t.Errorf("expected error here for unknown input symbol : %v", "A")
			return
		}
	}

}

func Test_NewSymbolSet_DontFailIfIPAContainsDuplicates(t *testing.T) {
	symbols := []Symbol{
		Symbol{"a", NonSyllabic, "", IPASymbol{"A", "U+0041"}},
		Symbol{"A", Syllabic, "", IPASymbol{"A", "U+0041"}},
		Symbol{"p", NonSyllabic, "", IPASymbol{"P", "U+0050"}},
		Symbol{" ", PhonemeDelimiter, "", IPASymbol{" ", "U+0020"}},
	}
	_, err := NewSymbolSet("test", symbols)
	if err != nil {
		t.Errorf("NewSymbolSet() didn't expect error when output phoneme set contains duplicates : %v", err)
	}
}
