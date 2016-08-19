package symbolset

// SvNSTHardWired is a temporary function that should not be used in production
func SvNSTHardWired() (SymbolSet, error) {

	syms := []Symbol{

		Symbol{Desc: "sil", String: "i:", Cat: Syllabic},
		Symbol{Desc: "sill", String: "I", Cat: Syllabic},
		Symbol{Desc: "full", String: "u0", Cat: Syllabic},
		Symbol{Desc: "ful", String: "}:", Cat: Syllabic},
		Symbol{Desc: "matt", String: "a", Cat: Syllabic},
		Symbol{Desc: "mat", String: "A:", Cat: Syllabic},
		Symbol{Desc: "bot", String: "u:", Cat: Syllabic},
		Symbol{Desc: "bott", String: "U", Cat: Syllabic},
		Symbol{Desc: "häl", String: "E:", Cat: Syllabic},
		Symbol{Desc: "häll", String: "E", Cat: Syllabic},
		Symbol{Desc: "aula", String: "a*U", Cat: Syllabic},
		Symbol{Desc: "syl", String: "y:", Cat: Syllabic},
		Symbol{Desc: "syll", String: "Y", Cat: Syllabic},
		Symbol{Desc: "hel", String: "e:", Cat: Syllabic},
		Symbol{Desc: "herr,hett", String: "e", Cat: Syllabic},
		Symbol{Desc: "nöt", String: "2:", Cat: Syllabic},
		Symbol{Desc: "mött,förra", String: "9", Cat: Syllabic},
		Symbol{Desc: "mål", String: "o:", Cat: Syllabic},
		Symbol{Desc: "moll,håll", String: "O", Cat: Syllabic},
		Symbol{Desc: "bättre", String: "@", Cat: Syllabic},
		Symbol{Desc: "europa", String: "E*U", Cat: Syllabic},
		Symbol{Desc: "pol", String: "p", Cat: NonSyllabic},
		Symbol{Desc: "bok", String: "b", Cat: NonSyllabic},
		Symbol{Desc: "tok", String: "t", Cat: NonSyllabic},
		Symbol{Desc: "bort", String: "t`", Cat: NonSyllabic},
		Symbol{Desc: "mod", String: "m", Cat: NonSyllabic},
		Symbol{Desc: "nod", String: "n", Cat: NonSyllabic},
		Symbol{Desc: "dop", String: "d", Cat: NonSyllabic},
		Symbol{Desc: "bord", String: "d`", Cat: NonSyllabic},
		Symbol{Desc: "fot", String: "k", Cat: NonSyllabic},
		Symbol{Desc: "våt", String: "g", Cat: NonSyllabic},
		Symbol{Desc: "lång", String: "N", Cat: NonSyllabic},
		Symbol{Desc: "forna", String: "n`", Cat: NonSyllabic},
		Symbol{Desc: "fot", String: "f", Cat: NonSyllabic},
		Symbol{Desc: "våt", String: "v", Cat: NonSyllabic},
		Symbol{Desc: "kjol (in NST specs)", String: "s’", Cat: NonSyllabic},
		Symbol{Desc: "kjol (in NST actual transcriptions)", String: "s'", Cat: NonSyllabic},
		Symbol{Desc: "fors", String: "s`", Cat: NonSyllabic},
		Symbol{Desc: "rov", String: "r", Cat: NonSyllabic},
		Symbol{Desc: "lov", String: "l", Cat: NonSyllabic},
		Symbol{Desc: "sot", String: "s", Cat: NonSyllabic},
		Symbol{Desc: "sjok", String: "x\\", Cat: NonSyllabic},
		Symbol{Desc: "hot", String: "h", Cat: NonSyllabic},
		Symbol{Desc: "porla", String: "l`", Cat: NonSyllabic},
		Symbol{Desc: "jord", String: "j", Cat: NonSyllabic},
		Symbol{Desc: "syllable delimiter", String: "$", Cat: SyllableDelimiter},
		Symbol{Desc: "accent I", String: `"`, Cat: Stress},
		Symbol{Desc: "accent II", String: `""`, Cat: Stress},
		Symbol{Desc: "secondary stress", String: "%", Cat: Stress},
		Symbol{Desc: "phoneme delimiter", String: " ", Cat: PhonemeDelimiter},
	}

	return SymbolSet{Name: "sv.se.nst-SAMPA", Symbols: syms}, nil
	//return NewSymbolSet("sv.se.nst-SAMPA", syms)

}
