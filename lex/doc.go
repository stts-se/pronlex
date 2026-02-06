/*
Package lex is used for general 'container' classes such as entry, transcription, lemma, etc.

The main unit here is the entry. The entry contains everything related to a lexicon entry: orthography, transcriptions, lemma, compound parts, sources/references, et cetera.
It is implemented as a go struct, and it can automatically be mapped into a JSON object.

The Entry struct is defined here: https://godoc.org/github.com/stts-se/pronlex/lex#Entry

A few JSON examples:

	// Minimal example (English)
	{
	   strn: "things",
	   transcriptions: [
	   {
	      strn: "' T I N z"
	   }
	   ]
	}



	// Entry "things" from the CMU (US English) lexicon
	{
	   id: 112326,
	   lexRef: {
	      DBRef: "en_am_cmu_lex",
	      LexName: "en-us.cmu"
	   },
	   strn: "things",
	   language: "en-us",
	   partOfSpeech: "",
	   morphology: "",
	   wordParts: "",
	   lemma: {
	      id: 0,
	      strn: "",
	      reading: "",
	      paradigm: ""
	   },
	   transcriptions: [
	   {
	      id: 120059,
	      entryId: 112326,
	      strn: "' T I N z",
	      language: "",
	      sources: [ ]
	   }
	   ],
	   status: {
	      id: 112326,
	      name: "imported",
	      source: "cmu",
	      timestamp: "2017-09-20T13:13:21Z",
	      current: true
	   },
	   entryValidations: [ ],
	   preferred: false,
	   tag: ""
	}



	// Entry "hästar" from the Swedish demo lexicon
	{
	id: 6,
	lexRef: {
	   DBRef: "wikispeech_lexserver_testdb",
	   LexName: "sv"
	},
	   strn: "hästar",
	   language: "sv",
	   partOfSpeech: "NN",
	   morphology: "NEU IND PLU",
	   wordParts: "hästar",
	   lemma: {
	      id: 4,
	      strn: "häst",
	      reading: "",
	      paradigm: ""
	   },
	   transcriptions: [
	   {
	      id: 9,
	      entryId: 6,
	      strn: "" h E . s t a r",
	      language: "sv",
	      sources: [ ]
	   }
	   ],
	   status: {
	      id: 6,
	      name: "demo",
	      source: "auto",
	      timestamp: "2017-09-22T08:43:32Z",
	      current: true
	   },
	   entryValidations: [ ],
	   preferred: false,
	   tag: ""
	}
*/
package lex
