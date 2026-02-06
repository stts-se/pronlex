/*
Package line is used to define lexicon line formats for parsing input and printing output.

Interfaces:

* Format - simple line format definition (field names and indices)

* Parser - a more complex parser, containing a Format definition, but also adds the possibility to write specific code for parsing that cannot be handeled by the Format specs alone (multi-value fields, etc).

# THE WIKISPEECH FILE FORMAT

The Wikispeech lexicon file format is defined in ws.go. Lexicon files are tab separated text files (UTF-8 encoded), and should contain the fields listed below. Empty fields are allowed in most positions.

Any lexicon files you want to import into the lexicon database must be in this file format.

	Orth           The word's orthography
	Pos            The part of speech tag
	Morph          Morphological features (gender, number, etc)
	WordParts      Compound parts, if any, separated by a plus sign (+)
	Lemma          The word's lemma form
	Paradigm       The name of the paradigm used for inflections
	Lang           The word's language label
	Trans1         The first transcription (default for TTS)
	Translang1     The language of the Trans1
	Trans2         Alternative transcription
	Translang2     The language of the Trans2
	Trans3         Alternative transcription
	Translang3     The language of the Trans3
	Trans4         Alternative transcription
	Translang4     The language of the Trans4
	StatusName     Status of the lexicon entry
	StatusSource   Source of the status
	Preferred      Takes values 1/0, and is used to defined which reading for a specific
	               orthography should be the standard one (in case of homographs)
	Tag            A tag (string) that can be used to disambiguate between homographs if needed (default: empty)
	Comments       On or more comments containing a label (category), a comment (text), and a source (user or other source).
	               Comments are defined in the following format (separated by §§§):
	                 [label: comment text] (source) §§§ [anotherlabel: another comment] (anothersource_or_user)

Sample line:

	finalspelet	NN	SIN|DEF|NOM|NEU	final+spelet	finalspel	s7n-övriga ex träd	sv-se	f I . "" n A: l . % s p e: . l e t	sv-se							imported	nst	false	dummytag	[assign_to: john] (jane) §§§
	[nolabel: typo] (hanna)
*/
package line
