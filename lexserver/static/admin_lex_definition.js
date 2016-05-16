var ADMLD = {};


ADMLD.baseURL = window.location.origin;



ADMLD.AdminLexDefModel = function () {
    var self = this; 
    
    
    
    // TODO hard wired names. Fetch from somewhere?
    self.symbolCategories = 
	["syllabic", "non syllabic", "stress", "phoneme delimiter", "explicit phoneme delimiter", "syllable delimiter", "morpheme delimiter", "word delimiter"];
    
    

    // TODO remove this (see below)
    self.nRead = ko.observable(0);    
    
    // TODO remove this. Too slow for large files. Use file upload instead.
    self.readLexiconFile = function(fileObject) {
	var i = 0;
	new LineReader(fileObject).readLines(function(line){
	    i = i + 1;
	    if (i % 1000 === 0 ) {
		//console.log(i);
		self.nRead(i);
	    };
	});
    };
    
    
    
    self.lexicons = ko.observableArray();
    
    // selectedLexicon is a trigger for different things
    // Sample lexicon object: {"id":0,"name":"nisse2","symbolSetName":"kvack2"}
    self.selectedLexicon = ko.observable({'id': 0, 'name': '', 'symbolSetName': ''});
    
    self.addLexiconName = ko.observable("");
    self.addSymbolSetName = ko.observable("");

    self.loadLexiconNames = function () {
	
	$.getJSON(ADMLD.baseURL +"/listlexicons")
	    .done(function (data) {
		self.lexicons(data);
	    })
    	    .fail(function (xhr, textStatus, errorThrown) {
		alert("loadLexiconNames says: "+ xhr.responseText);
	    });
    };
    
    self.updateLexicon = function () {
	
    	if ( self.selectedLexicon().name.trim() === "" || self.selectedLexicon().symbolSetName.trim() === "" ) {
    	    alert("Name and Symbol set name field must not be empty")
    	    return;
    	}
	
    	var params = {'id' : self.selectedLexicon().id, 'name' : self.selectedLexicon().name, 'symbolsetname' : self.selectedLexicon().symbolSetName}
	
	console.log("updateLexicon params: "+ JSON.stringify(params));
    	
	$.get(ADMLD.baseURL + "/admin/insertorupdatelexicon", params)
    	    .done(function(data){
		console.log("updateLexicon $get return data: "+ JSON.stringify(data));
    		self.loadLexiconNames();
    		self.selectedLexicon(data); // ?
    	    })
    	    .fail(function (xhr, textStatus, errorThrown) {
		console.log("updateLexicon fail xhr: "+ JSON.stringify(xhr));
		console.log("updateLexicon xhr.responseText: "+ xhr.responseText);
    		console.log("updateLexicon fail textStatus: "+ textStatus);
		console.log("updateLexicon fail errorThrown: "+ errorThrown);
		alert("updateLexicon says: "+ xhr.responseText);
    	    });	
    };
    
    self.deleteLexicon = function (lexicon) {
	
	var params = {'id' : lexicon.id}
    	$.get(ADMLD.baseURL + "/admin/deletelexicon", params)
    	    .done(function(data){
    		// dumdelidum
    		self.loadLexiconNames();
    	    })
    	    .fail(function (xhr, textStatus, errorThrown) {
    		alert(xhr.responseText);	    
    	    });
    };
    

    
    // An object/hash with symbol set name as key and a list of symbol objects as value
    self.symbolSets = ko.observable({});

    self.deleteSymbol = function(zymbl) {
	var currSyms = self.symbolSets()[self.selectedLexicon().symbolSetName];
	
	// This appears to be JavaScript's way of removing an entry from an array: 
	var i = currSyms.indexOf(zymbl);
	if(i != -1) {
	    currSyms.splice(i, 1);
	}
	
	// update to trigger event
	// TODO why is this needed?
	self.selectedLexicon(self.selectedLexicon());
	
    };
    

    // List of Symbol objects
    self.selectedSymbolSet = ko.computed(function() {
	
	if (self.symbolSets().hasOwnProperty(self.selectedLexicon().symbolSetName)) {
	    return self.symbolSets()[self.selectedLexicon().symbolSetName];
	} else {
	    return [];
	};
    }, this);
    self.saveSymbolSetToDB = function () {
	
	var ssName = self.selectedLexicon().symbolSetName;
	var ss = self.symbolSets()[ssName];
	if (typeof ss === 'undefined' ) return;
	for(var i = 0; i < ss.length; i++) {
	    console.log(JSON.stringify(ss[i]));
	};
    };


    // A sample symbol: {"symbol":"O","category":"Phoneme","description":"h(å)ll","ipa":"ɔ"}
    self.selectedSymbol = ko.observable({});

    self.loadSymbolSet = function () {
	if(self.selectedLexicon() !== undefined) { // TODO Gör man så?
	    //$.getJSON(DMCRLX.baseURL +"/admin/listphonemesymbols", {lexiconId: DMCRLX.selectedLexicon().id}, function (data) {
	    $.getJSON(ADMLD.baseURL +"/admin/listsymbolset", {lexiconId: self.selectedLexicon.id}, function (data) {
		console.log("FFFFF> "+ JSON.stringify(data));
		var syms = _.map(data, function (s) {
		    return {'lexiconId': s.lexiconId, 'symbol': s.symbol, 'category': s.category, 'description': s.description, 'ipa': s.ipa};
		}); 
		DMCRLX.symbolSet(syms);
	    })
		.fail(function (xhr, textStatus, errorThrown) {
		    alert(xhr.responseText);
		});
	}
    };

    
    self.saveSymbolSet = function () {
	$.post(ADMLD.baseURL +"/admin/savesymbolset", JSON.stringify(self.selectedSymbolSet))
	    .fail(function (xhr, textStatus, errorThrown) {
		alert(xhr.responseText);
	    });
    };
    

    self.setSelectedSymbol= function (symbol) {
	self.selectedSymbol(symbol);
    };
    
    // TODO hard wired list of symbol set file header field names
    // Fetch from somewhere?
    self.headerFields = {'DESCRIPTION' : true, 'SYMBOL': true, 'IPA': true, 'CATEGORY': true};
    self.readSymbolSetFile = function (symbolSetfile) {
	
	// returns hash of header field name -> field index
	function headerIndexMap(header) {
	    var rez = {};
	    
	    var fields = header.trim().split(/\t/); 
	    // TODO hard wired number of fields
	    // TODO proper error handling
	    if (fields.length !== 4) { 
		alert("Wrong number of fields in header: "+ header);
		return;
	    };
	    for(var i = 0; i < fields.length; i++) {
		if (! self.headerFields.hasOwnProperty(fields[i])) {
		    // TODO proper error handling
		    alert("Unknown header field: "+ fields[i]);
		}  
		rez[fields[i]] = i;
	    };
	    return rez;
	};

	var reader = new FileReader();
	reader.onloadend = function(evt) {      
	    // Currently expecting hard wired tab separated format: 
	    // DESC/EXAMPLE	NST-XSAMPA	WS-SAMPA	IPA	TYPE
	    // Lines starting with # are descarded
	    
            lines = evt.target.result.split(/\r?\n/);
            if( lines.length > 0 ) {
		var header = lines.shift();
		var headerIndexes = headerIndexMap(header);
		
	    } else {  // TODO How do you do error handling when asynchronously reading a file?
		alert("Empty input file: "+ symbolSetfile.name)
		return; // ?
	    }
	    //var newSyms = [];
	    lines.forEach(function (line) {
		if (line.trim() === "") return; // "continue"
		if (line.trim().startsWith("#")) return; // "continue"
		
		var fs = line.split(/\t/);
		// TODO hard wired
		if (fs.length !== 4 ) alert("Wrong number of fields in line: "+ line);
		var symbol = {'symbol': fs[headerIndexes['SYMBOL']],
			      'category': fs[headerIndexes['CATEGORY']],
			      'description': fs[headerIndexes['DESCRIPTION']],
			      'ipa': fs[headerIndexes['IPA']]
			     };
		
		if(! self.symbolSets().hasOwnProperty(self.selectedLexicon().symbolSetName)) {
		    self.symbolSets()[self.selectedLexicon().symbolSetName] = [];
		};
		
		self.addSymbolToSet(self.selectedLexicon().symbolSetName, symbol);
		
            });
	    
	    // update to trigger event
	    // TODO why is this needed?
	    self.selectedLexicon(self.selectedLexicon());
	};
	
	reader.readAsText(symbolSetfile,"UTF-8");
    };
    

    
    self.addLexicon = function() {
		
	var newLex = {'id': 0, 'name' :  self.addLexiconName(), 'symbolSetName': self.addSymbolSetName()};
	self.selectedLexicon(newLex);
	
	self.updateLexicon();
	
	self.addLexiconName("");
	self.addSymbolSetName("");
    };
    
    // These fields maka up the definition of a symbol
    self.symbolToAdd = ko.observable();
    self.categoryToAdd = ko.observable();
    self.descriptionToAdd = ko.observable();
    self.ipaToAdd = ko.observable();
    
    self.addSymbol = function() {
	
	var newSymbol = {'symbol': self.symbolToAdd(), 
			 'category': self.categoryToAdd(), 
			 'description': self.descriptionToAdd(), 
			 'ipa': self.ipaToAdd()};
	
	self.addSymbolToSet(self.selectedLexicon().symbolSetName, newSymbol);
	
	// update to trigger event
	// TODO why is this needed?
	self.selectedLexicon(self.selectedLexicon());
    };
    
    self.addSymbolToSet = function(symbolSetName, symbol) {	
	if ( ! self.symbolSets().hasOwnProperty(symbolSetName) ) {
	    var ss = self.symbolSets();		
	    ss[symbolSetName] = [];
	};
	self.symbolSets()[symbolSetName].push(symbol);
    };
    
    self.setSelectedIPA = function(symbol) {
	self.ipaToAdd(symbol.symbol);
	self.descriptionToAdd(symbol.description);
    };
    
    
    self.nColumns = ko.observable(15);
    
    self.createIPATableRows = function (nColumns, ipaList ) {
	var res = [];
	var row = [];
	var j = 0;
	for(var i = 0; i < self.ipaTable.length; i++) {
	    
	    var ipaChar = {'symbol': self.ipaTable[i].symbol, 'description': self.ipaTable[i].description};
	    row.push(ipaChar);
	    j++;
	    if ( j === nColumns) {
		res.push(row);
		row = [];
		j = 0;
	    };
	}; // <- for
	// "flush":
	if ( j !== nColumns) {
	    res.push(row);
	};
	return res;
    }; 
    

    
    // TODO remove hard wired IPA table
    // This should be downloaded from lexserver: ipa_table.txt
    self.ipaTable = [{"symbol": "ɐ", "description":  "Near-open central vowel"},
		     {"symbol": "ɑ", "description":  "Open back unrounded vowel"},
		     {"symbol": "ɒ", "description":  "Open back rounded vowel"},
		     {"symbol": "ɓ", "description":  "Voiced bilabial implosive"},
		     {"symbol": "ɔ", "description":  "Open-mid back rounded vowel"},
		     {"symbol": "ɕ", "description":  "Voiceless alveolo-palatal fricative"},
		     {"symbol": "ɖ", "description":  "Voiced retroflex plosive"},
		     {"symbol": "ɗ", "description":  "Voiced alveolar implosive"},
		     {"symbol": "ɘ", "description":  "Close-mid central unrounded vowel"},
		     {"symbol": "ə", "description":  "Mid central vowel"},
		     {"symbol": "ɚ", "description":  "Rhotacized Mid central vowel"},
		     {"symbol": "ɛ", "description":  "Open-mid front unrounded vowel"},
		     {"symbol": "ɜ", "description":  "Open-mid central unrounded vowel"},
		     {"symbol": "ɝ", "description":  "Rhotacized Open-mid central unrounded vowel"},
		     {"symbol": "ɞ", "description":  "Open-mid central rounded vowel"},
		     {"symbol": "ɟ", "description":  "Voiced palatal plosive"},
		     {"symbol": "ɠ", "description":  "Voiced velar implosive"},
		     {"symbol": "ɡ", "description":  "Voiced velar plosive"},
		     {"symbol": "ɢ", "description":  "Voiced uvular plosive"},
		     {"symbol": "ɣ", "description":  "Voiced velar fricative"},
		     {"symbol": "ɤ", "description":  "Close-mid back unrounded vowel"},
		     {"symbol": "ɥ", "description":  "Labial-palatal approximant"},
		     {"symbol": "ɦ", "description":  "Voiced glottal fricative"},
		     {"symbol": "ɧ", "description":  "Swedish sj-sound. Similar to: Voiceless postalveolar fricative, Voiceless velar fricative"},
		     {"symbol": "ɨ", "description":  "Close central unrounded vowel"},
		     {"symbol": "ɩ", "description":  "pre-1989 form of 'ɪ' (obsolete)"},
		     {"symbol": "ɪ", "description":  "Near-close near-front unrounded vowel"},
		     {"symbol": "ɫ", "description":  "Velar/pharyngeal Alveolar lateral approximant"},
		     {"symbol": "ɬ", "description":  "Voiceless alveolar lateral fricative"},
		     {"symbol": "ɭ", "description":  "Retroflex lateral approximant"},
		     {"symbol": "ɮ", "description":  "Voiced alveolar lateral fricative"},
		     {"symbol": "ɯ", "description":  "Close back unrounded vowel"},
		     {"symbol": "ɰ", "description":  "Velar approximant"},
		     {"symbol": "ɱ", "description":  "Labiodental nasal"},
		     {"symbol": "ɲ", "description":  "Palatal nasal"},
		     {"symbol": "ɳ", "description":  "Retroflex nasal"},
		     {"symbol": "ɴ", "description":  "Uvular nasal"},
		     {"symbol": "ɵ", "description":  "Close-mid central rounded vowel"},
		     {"symbol": "ɶ", "description":  "Open front rounded vowel"},
		     {"symbol": "ɷ", "description":  "pre-1989 form of 'ʊ' (obsolete)"},
		     {"symbol": "ɸ", "description":  "Voiceless bilabial fricative"},
		     {"symbol": "ɹ", "description":  "Alveolar approximant"},
		     {"symbol": "ɺ", "description":  "Alveolar lateral flap"},
		     {"symbol": "ɻ", "description":  "Retroflex approximant"},
		     {"symbol": "ɼ", "description":  "Alveolar trill"},
		     {"symbol": "ɽ", "description":  "Retroflex flap"},
		     {"symbol": "ɾ", "description":  "Alveolar tap"},
		     {"symbol": "ɿ", "description":  "Syllabic voiced alveolar fricative (Sinologist usage)"},
		     {"symbol": "ʀ", "description":  "Uvular trill"},
		     {"symbol": "ʁ", "description":  "Voiced uvular fricative"},
		     {"symbol": "ʂ", "description":  "Voiceless retroflex fricative"},
		     {"symbol": "ʃ", "description":  "Voiceless postalveolar fricative"},
		     {"symbol": "ʄ", "description":  "Voiced palatal implosive"},
		     {"symbol": "ʅ", "description":  "Syllabic voiced retroflex fricative (Sinologist usage)"},
		     {"symbol": "ʆ", "description":  "Voiceless alveolo-palatal fricative (obsolete)"},
		     {"symbol": "ʇ", "description":  "Dental click (obsolete)"},
		     {"symbol": "ʈ", "description":  "Voiceless retroflex plosive"},
		     {"symbol": "ʉ", "description":  "Close central rounded vowel"},
		     {"symbol": "ʊ", "description":  "Near-close near-back rounded vowel"},
		     {"symbol": "ʋ", "description":  "Labiodental approximant"},
		     {"symbol": "ʌ", "description":  "Open-mid back unrounded vowel"},
		     {"symbol": "ʍ", "description":  "Voiceless labiovelar approximant"},
		     {"symbol": "ʎ", "description":  "Palatal lateral approximant"},
		     {"symbol": "ʏ", "description":  "Near-close near-front rounded vowel"},
		     {"symbol": "ʐ", "description":  "Voiced retroflex fricative"},
		     {"symbol": "ʑ", "description":  "Voiced alveolo-palatal fricative"},
		     {"symbol": "ʒ", "description":  "Voiced postalveolar fricative"},
		     {"symbol": "ʓ", "description":  "Voiced alveolo-palatal fricative (obsolete)"},
		     {"symbol": "ʔ", "description":  "Glottal stop"},
		     {"symbol": "ʕ", "description":  "Voiced pharyngeal fricative"},
		     {"symbol": "ʖ", "description":  "Alveolar lateral click (obsolete)"},
		     {"symbol": "ʗ", "description":  "Postalveolar click (obsolete)"},
		     {"symbol": "ʘ", "description":  "Bilabial click"},
		     {"symbol": "ʙ", "description":  "Bilabial trill"},
		     {"symbol": "ʚ", "description":  "A mistake for [œ]"},
		     {"symbol": "ʛ", "description":  "Voiced uvular implosive"},
		     {"symbol": "ʜ", "description":  "Voiceless epiglottal fricative"},
		     {"symbol": "ʝ", "description":  "Voiced palatal fricative"},
		     {"symbol": "ʞ", "description":  "Velar click (obsolete)"},
		     {"symbol": "ʟ", "description":  "Velar lateral approximant"},
		     {"symbol": "ʠ", "description":  "'Voiceless' uvular implosive (obsolete)"},
		     {"symbol": "ʡ", "description":  "Epiglottal plosive"},
		     {"symbol": "ʢ", "description":  "Voiced epiglottal fricative"},
		     {"symbol": "ʣ", "description":  "Voiced alveolar affricate (obsolete)"},
		     {"symbol": "ʤ", "description":  "Voiced postalveolar affricate (obsolete)"},
		     {"symbol": "ʥ", "description":  "Voiced alveolo-palatal affricate (obsolete)"},
		     {"symbol": "ʦ", "description":  "Voiceless alveolar affricate (obsolete)"},
		     {"symbol": "ʧ", "description":  "Voiceless postalveolar affricate (obsolete)"},
		     {"symbol": "ʨ", "description":  "Voiceless alveolo-palatal affricate (obsolete)"},
		     {"symbol": "ʩ", "description":  "velopharyngeal fricative"},
		     {"symbol": "ʪ", "description":  "voiceless lateral alveolar fricative"},
		     {"symbol": "ʫ", "description":  "voiced lateral alveolar fricative"},
		     {"symbol": "ʬ", "description":  "Bilabial percussive"},
		     {"symbol": "ʭ", "description":  "Bidental percussive"},
		     {"symbol": "ʮ", "description":  "Syllabic labialized voiced alveolar fricative (Sinologist usage)"},
		     {"symbol": "ʯ", "description":  "Syllabic labialized voiced retroflex fricative (Sinologist usage)"},
		     {"symbol": "a", "description":  "Open front unrounded vowel"},
		     {"symbol": "b", "description":  "bilabial plosive"},
		     {"symbol": "c", "description":  "palatal plosive"},
		     {"symbol": "d", "description":  "alveolar plosive"},
		     {"symbol": "e", "description":  "close-mid front unrounded vowel"},
		     {"symbol": "f", "description":  "labiodental fricative"},
		     {"symbol": "g", "description":  "velar plosive Ascii g"},
		     {"symbol": "h", "description":  "glottal fricative"},
		     {"symbol": "i", "description":  "close front unrounded vowel"},
		     {"symbol": "j", "description":  "palatal approximant"},
		     {"symbol": "k", "description":  "velar plosive"},
		     {"symbol": "l", "description":  "lateral alveolar approximant"},
		     {"symbol": "l̩̩̩", "description":  "syllabic l"},
		     {"symbol": "m", "description":  "bilabial nasal"},
		     {"symbol": "m̩̩", "description":  "syllabic m"},
		     {"symbol": "n", "description":  "alveolar nasal"},
		     {"symbol": "n̩̩̩", "description":  "syllabic n"},
		     {"symbol": "o", "description":  "close-mid back rounded vowel"},
		     {"symbol": "p", "description":  "bilabial plosive"},
		     {"symbol": "q", "description":  "uvular plosive"},
		     {"symbol": "r", "description":  "alveolar trill"},
		     {"symbol": "s", "description":  "alveolar fricative"},
		     {"symbol": "t", "description":  "alveolar plosive"},
		     {"symbol": "u", "description":  "close back rounded vowel"},
		     {"symbol": "v", "description":  "labiodental fricative"},
		     {"symbol": "w", "description":  "labial-velar approximant"},
		     {"symbol": "x", "description":  "velar fricative"},
		     {"symbol": "y", "description":  "close front rounded vowel"},
		     {"symbol": "z", "description":  "alveolar fricative"},
		     {"symbol": "æ", "description":  "raised-open front unrounded vowel"},
		     {"symbol": "ç", "description":  "palatal fricative"},
		     {"symbol": "ð", "description":  "dental fricative"},
		     {"symbol": "ø", "description":  "close-mid front rounded vowel"},
		     {"symbol": "ħ", "description":  "pharyngeal fricative"},
		     {"symbol": "ŋ", "description":  "velar nasal"},
		     {"symbol": "œ", "description":  "Open-mid front rounded vowel"},
		     {"symbol": "β", "description":  "bilabial fricative"},
		     {"symbol": "θ", "description":  "dental fricative"},
		     {"symbol": "χ", "description":  "uvular fricative"},
		     {"symbol": "aː", "description":  "Open front unrounded vowel (long)"},
		     {"symbol": "eː", "description":  "close-mid front unrounded vowel (long)"},
		     {"symbol": "iː", "description":  "close front unrounded vowel (long)"},
		     {"symbol": "oː", "description":  "close-mid back rounded vowel (long)"},
		     {"symbol": "uː", "description":  "close back rounded vowel (long)"},
		     {"symbol": "yː", "description":  "close front rounded vowel (long)"},
		     {"symbol": "æː", "description":  "raised-open front unrounded vowel (long)"},
		     {"symbol": "øː", "description":  "close-mid front rounded vowel (long)"},
		     {"symbol": "œː", "description":  "Open-mid front rounded vowel (long)"},
		     {"symbol": "ɐː", "description":  "Near-open central vowel (long)"},
		     {"symbol": "ɑː", "description":  "Open back unrounded vowel (long)"},
		     {"symbol": "ɒː", "description":  "Open back rounded vowel (long)"},
		     {"symbol": "ɔː", "description":  "Open-mid back rounded vowel (long)"},
		     {"symbol": "ɘː", "description":  "Close-mid central unrounded vowel (long)"},
		     {"symbol": "əː", "description":  "Mid central vowel (long)"},
		     {"symbol": "ɚː", "description":  "Rhotacized Mid central vowel (long)"},
		     {"symbol": "ɛː", "description":  "Open-mid front unrounded vowel (long)"},
		     {"symbol": "ɜː", "description":  "Open-mid central unrounded vowel (long)"},
		     {"symbol": "ɝː", "description":  "Rhotacized Open-mid central unrounded vowel (long)"},
		     {"symbol": "ɞː", "description":  "Open-mid central rounded vowel (long)"},
		     {"symbol": "ɤː", "description":  "Close-mid back unrounded vowel (long)"},
		     {"symbol": "ɨː", "description":  "Close central unrounded vowel (long)"},
		     {"symbol": "ɪː", "description":  "Near-close near-front unrounded vowel (long)"},
		     {"symbol": "ɯː", "description":  "Close back unrounded vowel (long)"},
		     {"symbol": "ɵː", "description":  "Close-mid central rounded vowel (long)"},
		     {"symbol": "ɶː", "description":  "Open front rounded vowel (long)"},
		     {"symbol": "ʉː", "description":  "Close central rounded vowel (long)"},
		     {"symbol": "ʊː", "description":  "Near-close near-back rounded vowel (long)"},
		     {"symbol": "ʌː", "description":  "Open-mid back unrounded vowel (long)"},
		     {"symbol": "ʏː", "description":  "Near-close near-front rounded vowel (long)"}];

    self.ipaSymbols = ko.observableArray();
    self.loadIPASymbols = function () {
	self.ipaSymbols.push("");
	for(var i = 0; i < self.ipaTable.length; i++) {
	    self.ipaSymbols.push(self.ipaTable[i].symbol);
	}
    };
    
    
    self.ipaTableRows = ko.computed(function() {
	var n = self.nColumns();
	return self.createIPATableRows(n, self.dummyIPA);
    }); 
};


var adm = new ADMLD.AdminLexDefModel();
adm.loadIPASymbols();
ko.applyBindings(adm);
adm.loadLexiconNames();


// For marking the selected row in a table
$(document).on('click', '.selectable', (function(){
    $(this).addClass("selected").siblings().removeClass("selected");
}));
