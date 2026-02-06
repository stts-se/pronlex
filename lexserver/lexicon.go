package main

// The handlers of calls prefixed with '/lexicon/':

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"

	"github.com/stts-se/pronlex/lex"
)

// var lexiconValidationPage = urlHandler{
// 	name:     "validation (page)",
// 	url:      "/validation_page",
// 	help:     "Validate lexicon (GUI).",
// 	examples: []string{"/validation_page"},
// 	handler: func(w http.ResponseWriter, r *http.Request) {
// 		http.ServeFile(w, r, filepath.Join(staticFolder, "lexicon/validation_page.html"))
// 	},
// }

// var lexiconUpdateEntryURL = "/updateentry?entry={...}"
// TODO: Use a lexicon that exists!
var lexiconUpdateEntryURL = `/updateentry?entry={
    "id": 9,
    "lexRef": {
        "DBRef": "wikispeech_lexserver_testdb",
        "LexName": "sv"
    },
    "strn": "dom",
    "language": "sv",
    "partOfSpeech": "NN",
    "morphology": "UTR IND SIN",
    "wordParts": "dom",
        "lemma": {
        "id": 5,
        "strn": "dom",
        "reading": "",
        "paradigm": ""
    },
    "tag": "building",
    "transcriptions": [
    {
        "id": 12,
        "entryId": 9,
        "strn": "\" d o: m",
        "language": "sv",
        "sources": [ ]
    }
    ],
    "status": {
    "id": 9,
    "name": "demo",
    "source": "auto",
    "timestamp": "2017-11-14T09:34:30Z",
        "current": true
    },
    "entryValidations": [ ],
    "preferred": false
}`

var lexiconUpdateEntry = urlHandler{
	name:     "updateentry",
	url:      "/updateentry",
	help:     "Updates an entry in the database. Input is an entry variable in JSON format. For examples, see <a href=\"https://godoc.org/github.com/stts-se/pronlex/lex\">package documentation</a>.",
	examples: []string{lexiconUpdateEntryURL},
	handler: func(w http.ResponseWriter, r *http.Request) {
		entryJSON := getParam("entry", r)
		//body, err := ioutil.ReadAll(r.Body)
		var e lex.Entry
		err := json.Unmarshal([]byte(entryJSON), &e)
		if err != nil {
			log.Printf("lexserver: Failed to unmarshal json: %v", err)
			http.Error(w, fmt.Sprintf("failed to process incoming Entry json : %v", err), http.StatusInternalServerError)
			return
		}

		// Underscore below matches bool indicating if any update has taken place. Return this info?
		res, _, err2 := dbm.UpdateEntry(e)
		if err2 != nil {
			log.Printf("lexserver: Failed to update entry : %v", err2)
			http.Error(w, fmt.Sprintf("failed to update Entry : %v", err2), http.StatusInternalServerError)
			return
		}

		res0, err3 := json.Marshal(res)
		if err3 != nil {
			log.Printf("lexserver: Failed to marshal entry : %v", err3)
			http.Error(w, fmt.Sprintf("failed return updated Entry : %v", err3), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		fmt.Fprint(w, string(res0))
	},
}

var lexiconUpdateValidation = urlHandler{
	name:     "updatevalidation",
	url:      "/updatevalidation",
	help:     "Updates the validation for an entry in the database. Input is an entry variable in JSON format. For examples, see <a href=\"https://godoc.org/github.com/stts-se/pronlex/lex\">package documentation</a>.",
	examples: []string{},
	handler: func(w http.ResponseWriter, r *http.Request) {
		entryJSON := getParam("entry", r)
		if entryJSON == "" {
			log.Printf("lexiconUpdateValidation got empty input entry")
			http.Error(w, "failed to process incoming Entry json : empty input entry", http.StatusInternalServerError)
			return
		}
		//body, err := ioutil.ReadAll(r.Body)
		var e lex.Entry
		err := json.Unmarshal([]byte(entryJSON), &e)
		if err != nil {
			log.Printf("lexserver: Failed to unmarshal json: %v", err)
			http.Error(w, fmt.Sprintf("failed to process incoming Entry json : %v", err), http.StatusInternalServerError)
			return
		}

		err2 := dbm.UpdateValidation(e)
		if err2 != nil {
			log.Printf("lexserver: Failed to update entry : %v", err2)
			http.Error(w, fmt.Sprintf("failed to update Entry : %v", err2), http.StatusInternalServerError)
			return
		}

		//w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		//fmt.Fprint(w, "Validation updated")
	},
}

// LexWithEntryCount is a struct for collecting lexicon info for json result
type LexWithEntryCount struct {
	Name          string `json:"name"`
	SymbolSetName string `json:"symbolSetName"`
	Locale        string `json:"locale"`
	EntryCount    int64  `json:"entryCount"`
}

var lexiconList = urlHandler{
	name:     "list",
	url:      "/list",
	help:     "Lists available lexicons along with some basic info.",
	examples: []string{"/list"},
	handler: func(w http.ResponseWriter, r *http.Request) {
		lexs0, err := dbm.ListLexicons() // TODO error handling
		if err != nil {
			http.Error(w, fmt.Sprintf("list lexicons failed : %v", err), http.StatusInternalServerError)
			return
		}
		var lexs []LexWithEntryCount = []LexWithEntryCount{}
		for _, lex := range lexs0 {
			entryCount, err := dbm.EntryCount(lex.LexRef)
			if err != nil {
				http.Error(w, fmt.Sprintf("lexicon stats failed : %v", err), http.StatusInternalServerError)
				return
			}

			locale, err := dbm.Locale(lex.LexRef)
			if err != nil {
				http.Error(w, fmt.Sprintf("lexicon stats failed : %v", err), http.StatusInternalServerError)
				return
			}
			lexs = append(lexs, LexWithEntryCount{Name: lex.LexRef.String(), SymbolSetName: lex.SymbolSetName, Locale: locale, EntryCount: entryCount})
		}
		jsn, err := marshal(lexs, r)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed marshalling : %v", err), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		fmt.Fprint(w, string(jsn))
	},
}

var lexiconListCurrentEntryStatuses = urlHandler{
	name:     "list_current_entry_statuses",
	url:      "/list_current_entry_statuses/{lexicon_name}",
	help:     "List current entry statuses. Optional param freq set to true will include frequencies for each status.",
	examples: []string{"/list_current_entry_statuses/wikispeech_lexserver_testdb:sv"},
	handler: func(w http.ResponseWriter, r *http.Request) {
		lexRef, err := getLexRefParam(r)
		if err != nil {
			log.Println(err)
			http.Error(w, fmt.Sprintf("couldn't parse lexicon ref %v : %v", lexRef, err), http.StatusInternalServerError)
			return
		}
		freq := getParam("freq", r)
		if freq == "true" {
			statuses, err := dbm.ListCurrentEntryStatusesWithFreq(lexRef)
			if err != nil {
				http.Error(w, fmt.Sprintf("listCurrentEntryStatuses : %v", err), http.StatusInternalServerError)
				return
			}
			j, err := json.Marshal(statuses)
			if err != nil {
				http.Error(w, fmt.Sprintf("listCurrentEntryStatuses : %v", err), http.StatusInternalServerError)
				return
			}

			fmt.Fprint(w, string(j))
		} else {
			statuses, err := dbm.ListCurrentEntryStatuses(lexRef)
			if err != nil {
				http.Error(w, fmt.Sprintf("listCurrentEntryStatuses : %v", err), http.StatusInternalServerError)
				return
			}
			j, err := json.Marshal(statuses)
			if err != nil {
				http.Error(w, fmt.Sprintf("listCurrentEntryStatuses : %v", err), http.StatusInternalServerError)
				return
			}

			fmt.Fprint(w, string(j))
		}
	},
}

var lexiconListCommentLabels = urlHandler{
	name:     "list_comment_labels",
	url:      "/list_comment_labels/{lexicon_name}",
	help:     "List comment labels.",
	examples: []string{"/list_current_entry_statuses/wikispeech_lexserver_testdb:sv"},
	handler: func(w http.ResponseWriter, r *http.Request) {
		lexRef, err := getLexRefParam(r)
		if err != nil {
			log.Println(err)
			http.Error(w, fmt.Sprintf("couldn't parse lexicon ref %v : %v", lexRef, err), http.StatusInternalServerError)
			return
		}

		statuses, err := dbm.ListCommentLabels(lexRef)
		if err != nil {
			http.Error(w, fmt.Sprintf("listCommentLabels: %v", err), http.StatusInternalServerError)
			return
		}
		j, err := json.Marshal(statuses)
		if err != nil {
			http.Error(w, fmt.Sprintf("listCommentLabels : %v", err), http.StatusInternalServerError)
			return
		}

		fmt.Fprint(w, string(j))
	},
}

var lexiconListCurrentEntryUsers = urlHandler{
	name:     "list_current_entry_users",
	url:      "/list_current_entry_users/{lexicon_name}",
	help:     "List current entry users. Optional param freq set to true will include frequencies for each user.",
	examples: []string{"/list_current_entry_users/wikispeech_lexserver_testdb:sv?freq=true"},
	handler: func(w http.ResponseWriter, r *http.Request) {
		lexRef, err := getLexRefParam(r)
		if err != nil {
			log.Println(err)
			http.Error(w, fmt.Sprintf("couldn't parse lexicon ref %v : %v", lexRef, err), http.StatusInternalServerError)
			return
		}
		freq := getParam("freq", r)
		if freq == "true" {
			users, err := dbm.ListCurrentEntryUsersWithFreq(lexRef)
			if err != nil {
				http.Error(w, fmt.Sprintf("listCurrentEntryUsers : %v", err), http.StatusInternalServerError)
				return
			}
			j, err := json.Marshal(users)
			if err != nil {
				http.Error(w, fmt.Sprintf("listCurrentEntryUsers : %v", err), http.StatusInternalServerError)
				return
			}
			fmt.Fprint(w, string(j))

		} else {
			users, err := dbm.ListCurrentEntryUsers(lexRef)
			if err != nil {
				http.Error(w, fmt.Sprintf("listCurrentEntryUsers : %v", err), http.StatusInternalServerError)
				return
			}
			j, err := json.Marshal(users)
			if err != nil {
				http.Error(w, fmt.Sprintf("listCurrentEntryUsers : %v", err), http.StatusInternalServerError)
				return
			}
			fmt.Fprint(w, string(j))
		}

	},
}

// TODO cut-n-paste from above
var lexiconListAllEntryStatuses = urlHandler{
	name:     "list_all_entry_statuses",
	url:      "/list_all_entry_statuses/{lexicon_name}",
	help:     "List all entry statuses.",
	examples: []string{"/list_all_entry_statuses/wikispeech_lexserver_testdb:sv"},
	handler: func(w http.ResponseWriter, r *http.Request) {
		lexRef, err := getLexRefParam(r)
		if err != nil {
			log.Println(err)
			http.Error(w, fmt.Sprintf("couldn't parse lexicon ref %v : %v", lexRef, err), http.StatusInternalServerError)
			return
		}

		statuses, err := dbm.ListAllEntryStatuses(lexRef)
		if err != nil {
			http.Error(w, fmt.Sprintf("listAllEntryStatuses : %v", err), http.StatusInternalServerError)
			return
		}
		j, err := json.Marshal(statuses)
		if err != nil {
			http.Error(w, fmt.Sprintf("listAllEntryStatuses : %v", err), http.StatusInternalServerError)
			return
		}

		fmt.Fprint(w, string(j))
	},
}

// LexInfo is a struct for collecting lexicon info for json result
type LexInfo struct {
	Name          string `json:"name"`
	SymbolSetName string `json:"symbolSetName"`
}

var lexiconInfo = urlHandler{
	name:     "info",
	url:      "/info/{lexicon_name}",
	help:     "Get some basic lexicon info.",
	examples: []string{"/info/wikispeech_lexserver_testdb:sv"},
	handler: func(w http.ResponseWriter, r *http.Request) {
		lexRef, err := getLexRefParam(r)
		if err != nil {
			log.Println(err)
			http.Error(w, fmt.Sprintf("couldn't parse lexicon ref %v : %v", lexRef, err), http.StatusInternalServerError)
			return
		}

		lex, err := dbm.GetLexicon(lexRef)
		if err != nil {
			http.Error(w, fmt.Sprintf("get lexicon failed : %v", err), http.StatusInternalServerError)
			return
		}
		li := LexInfo{Name: lexRef.String(), SymbolSetName: lex.SymbolSetName}
		jsn, err := marshal(li, r)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed marshalling : %v", err), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		fmt.Fprint(w, string(jsn))
	},
}

var lexiconStats = urlHandler{
	name:     "stats",
	url:      "/stats/{lexicon_name}",
	help:     "Lists lexicon stats.",
	examples: []string{"/stats/wikispeech_lexserver_testdb:sv"},
	handler: func(w http.ResponseWriter, r *http.Request) {
		lexRef, err := getLexRefParam(r)
		if err != nil {
			log.Println(err)
			http.Error(w, fmt.Sprintf("couldn't parse lexicon ref %v : %v", lexRef, err), http.StatusInternalServerError)
			return
		}

		stats, err := dbm.LexiconStats(lexRef)
		if err != nil {
			http.Error(w, fmt.Sprintf("lexiconStatsHandler: call to  dbapi.LexiconStats failed : %v", err), http.StatusInternalServerError)
			return
		}
		res, err := json.Marshal(stats)

		if err != nil {
			http.Error(w, fmt.Sprintf("lexiconStatsHandler: failed to marshal struct : %v", err), http.StatusInternalServerError)
			return
		}

		fmt.Fprint(w, string(res))
	},
}

var lexiconLookup = urlHandler{
	name:     "lookup",
	url:      "/lookup",
	help:     "Lookup in lexicon.",
	examples: []string{"/lookup"},
	handler: func(w http.ResponseWriter, r *http.Request) {

		var err error

		u, err := url.Parse(r.URL.String())
		if err != nil {
			log.Printf("lexLookUpHandler failed to get params: %v", err)
			http.Error(w, fmt.Sprintf("%v", err), http.StatusInternalServerError)
			return
		}
		params := u.Query()
		if len(params) == 0 {
			log.Print("lexiconLookup: zero params, serving lexlookup.html")
			http.ServeFile(w, r, filepath.Join(staticFolder, "lexlookup.html"))
			return
		}
		for k, v := range params {
			if _, ok := knownParams[k]; !ok {
				log.Printf("lexiconLookup: unknown URL parameter: '%s': '%s'", k, v)
				http.Error(w, fmt.Sprintf("lexiconLookup: unknown URL parameter: '%s': '%s'", k, v), http.StatusBadRequest)
				return // NB: only informs about the first unknown param...
			}
		}

		q, err := queryFromParams(r)

		if err != nil {
			log.Printf("failed to process query params: %v", err)
			http.Error(w, fmt.Sprintf("%v", err), http.StatusBadRequest)
			return
		}

		res, err := dbm.LookUpIntoSlice(q)

		if err != nil {
			log.Printf("lexserver: Failed to get entries: %v", err)
			http.Error(w, fmt.Sprintf("%v", err), http.StatusInternalServerError)
			return
		}

		jsn, err := marshal(res, r)
		if err != nil {
			log.Printf("lexserver: Failed to marshal json: %v", err)
			http.Error(w, fmt.Sprintf("%v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		fmt.Fprint(w, string(jsn))
	},
}

type MiniEntry struct {
	Orth   string `json:"orth"`
	Tag    string `json:"tag"`
	Status string `json:"status"`
	Pref   bool   `json:"pref"`
}

var lexiconEntriesExist = urlHandler{
	name:     "entries_exist",
	url:      "/entries_exist",
	help:     "Lookup orthographies in the db and see if they exist as entries.",
	examples: []string{"/entries_exist?lexicons=wikispeech_lexserver_testdb:sv&words=hund,h%C3%A4st,hunnd"},
	handler: func(w http.ResponseWriter, r *http.Request) {

		var err error

		u, err := url.Parse(r.URL.String())
		if err != nil {
			log.Printf("entriesExistHandler failed to get params: %v", err)
			http.Error(w, fmt.Sprintf("%v", err), http.StatusInternalServerError)
			return
		}
		params := u.Query()
		for k, v := range params {
			// if _, ok := knownParams[k]; !ok {
			// 	log.Printf("entriesExist: unknown URL parameter: '%s': '%s'", k, v)
			// 	http.Error(w, fmt.Sprintf("lexiconLookup: unknown URL parameter: '%s': '%s'", k, v), http.StatusBadRequest)
			// 	return // NB: only informs about the first unknown param...
			// }
			if !(k == "lexicons" || k == "words" || k == "pagelength") {
				log.Printf("entriesExist: unknown URL parameter: '%s': '%s'. Valid parameters :  %s and %s", k, v, "lexicons", "words")
				http.Error(w, fmt.Sprintf("lexiconLookup: unknown URL parameter: '%s': '%s'", k, v), http.StatusBadRequest)
				return // NB: only informs about the first unknown param...
			}
		}

		q, err := queryFromParams(r)

		if err != nil {
			log.Printf("failed to process query params: %v", err)
			http.Error(w, fmt.Sprintf("%v", err), http.StatusBadRequest)
			return
		}
		var res = make(map[string][]MiniEntry)
		writer := lex.EntrySliceWriter{}
		err = dbm.LookUp(q, &writer)
		if err != nil {
			log.Printf("lexserver: Failed to get entries: %v", err)
			http.Error(w, fmt.Sprintf("%v", err), http.StatusInternalServerError)
			return
		}
		for _, e := range writer.Entries {
			orth := e.Strn
			me := MiniEntry{Orth: orth, Tag: e.Tag, Status: e.EntryStatus.Name, Pref: e.Preferred}
			res[orth] = append(res[orth], me)
		}
		for _, orth := range q.Query.Words {
			if _, ok := res[orth]; !ok {
				res[orth] = []MiniEntry{}
			}
		}
		jsn, err := marshal(res, r)
		if err != nil {
			log.Printf("lexserver: Failed to marshal json: %v", err)
			http.Error(w, fmt.Sprintf("%v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		fmt.Fprint(w, string(jsn))
	},
}

var lexiconAddEntryURL = `/addentry?lexicon_name=wikispeech_lexserver_testdb:sv&entry={
    "strn": "flesk",
    "language": "sv-se",
    "partOfSpeech": "NN",
    "morphology": "SIN-PLU|IND|NOM|NEU",
    "wordParts": "flesk",
    "lemma": {
	"strn": "flesk",
	"reading": "",
	"paradigm": "s7n-övriga ex träd"
    },
    "transcriptions": [
	{
	    "strn": "\" f l E s k",
	    "language": "sv-se"
	}
    ]
}
`

type IDs struct {
	IDs []int64 `json:"ids"`
}

var lexiconAddEntry = urlHandler{
	name:     "addentry",
	url:      "/addentry",
	help:     "Add an entry to the database. Input entry in JSON format. For examples, see <a href=\"https://godoc.org/github.com/stts-se/pronlex/lex\">package documentation</a>.",
	examples: []string{lexiconAddEntryURL},
	handler: func(w http.ResponseWriter, r *http.Request) {
		lexRef, err := getLexRefParam(r)
		if err != nil {
			log.Println(err)
			http.Error(w, fmt.Sprintf("couldn't parse lexicon ref %v : %v", lexRef, err), http.StatusInternalServerError)
			return
		}

		entryJSON := getParam("entry", r)
		var e lex.Entry
		err = json.Unmarshal([]byte(entryJSON), &e)
		if err != nil {
			log.Printf("lexserver: Failed to unmarshal json: %v", err)
			http.Error(w, fmt.Sprintf("failed to process incoming Entry json : %v", err), http.StatusInternalServerError)
			return
		}

		ids, err := dbm.InsertEntries(lexRef, []lex.Entry{e})
		if err != nil {
			msg := fmt.Sprintf("lexserver failed to update entry : %v", err)
			log.Println(msg)
			http.Error(w, msg, http.StatusInternalServerError)
			return
		}
		jsids := IDs{ids}
		res, err := json.Marshal(jsids)
		if err != nil {
			msg := fmt.Sprintf("lexserver: Failed to marshal ids : %v", err)
			log.Print(msg)
			http.Error(w, msg, http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		fmt.Fprint(w, string(res))
	},
}

var deleteEntry = func(w http.ResponseWriter, r *http.Request) {
	lexRef, err := getLexRefParam(r)

	if err != nil {
		log.Println(err)
		http.Error(w, fmt.Sprintf("couldn't parse lexicon ref %v : %v", lexRef, err), http.StatusInternalServerError)
		return
	}

	entryID := getParam("entry_id", r)
	id, err := strconv.ParseInt(entryID, 10, 64)
	if err != nil {
		log.Println(err)
		http.Error(w, fmt.Sprintf("failed to parse entry id %s : %v", entryID, err), http.StatusInternalServerError)
		return
	}

	idRes, err := dbm.DeleteEntry(id, lexRef)
	if err != nil {
		log.Println(err)
		http.Error(w, fmt.Sprintf("failed to detele entry id '%s' in lexicon '%s' : %v", entryID, lexRef.LexName, err), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "deleted entry id '%d' from lexicon '%s'\n", idRes, lexRef.LexName)
}

var lexiconDeleteEntry = urlHandler{
	name:     "delete_entry",
	url:      "/delete_entry/{lexicon_name}/{entry_id}",
	help:     "Delete an entry from the database.",
	examples: []string{},
	handler:  deleteEntry,
}

// var lexiconValidation = urlHandler{
// 	name:     "validation (api)",
// 	url:      "/validation/{lexicon_name}",
// 	help:     "Validate lexicon (API). Requires POST request.",
// 	examples: []string{},
// 	handler: func(w http.ResponseWriter, r *http.Request) {
// 		if r.Method != "POST" {
// 			http.Error(w, fmt.Sprintf("lexiconfileupload only accepts POST request, got %s", r.Method), http.StatusBadRequest)
// 			return
// 		}

// 		start := time.Now()

// 		clientUUID := getParam("client_uuid", r)

// 		if strings.TrimSpace(clientUUID) == "" {
// 			msg := "lexiconValidation got no client uuid"
// 			log.Println(msg)
// 			http.Error(w, msg, http.StatusBadRequest)
// 			return
// 		}
// 		conn, ok := webSocks.clients[clientUUID]
// 		if !ok {
// 			msg := fmt.Sprintf("lexiconValidation couldn't find connection for uuid %v", clientUUID)
// 			log.Println(msg)
// 			http.Error(w, msg, http.StatusBadRequest)
// 			return
// 		}
// 		logger := dbapi.NewWebSockLogger(conn)

// 		lexRef, err := getLexRefParam(r)
// 		if err != nil {
// 			log.Println(err)
// 			http.Error(w, fmt.Sprintf("couldn't parse lexicon ref %v : %v", lexRef, err), http.StatusInternalServerError)
// 			return
// 		}

// 		lexicon, err := dbm.GetLexicon(lexRef)
// 		if err != nil {
// 			http.Error(w, fmt.Sprintf("couldn't retrive lexicon : %v", err), http.StatusInternalServerError)
// 			return
// 		}
// 		vMut.Lock()
// 		v, err := vMut.service.ValidatorForName(lexicon.SymbolSetName)
// 		vMut.Unlock()
// 		if err != nil {
// 			msg := fmt.Sprintf("lexiconValidation failed to get validator for symbol set %v : %v", lexicon.SymbolSetName, err)
// 			log.Println(msg)
// 			http.Error(w, msg, http.StatusBadRequest)
// 			return
// 		}

// 		q := dbapi.Query{}
// 		stats, err := dbm.Validate(lexRef, logger, *v, q)
// 		if err != nil {
// 			msg := fmt.Sprintf("lexiconValidation failed validate : %v", err)
// 			log.Println(msg)
// 			http.Error(w, msg, http.StatusBadRequest)
// 			return
// 		}
// 		dur := round(time.Since(start), time.Second)
// 		fmt.Fprintf(w, "\nDuration %v\n", dur)
// 		fmt.Fprint(w, stats)
// 	},
// }

// func round(d, r time.Duration) time.Duration {
// 	if r <= 0 {
// 		return d
// 	}
// 	neg := d < 0
// 	if neg {
// 		d = -d
// 	}
// 	if m := d % r; m+m < r {
// 		d = d - m
// 	} else {
// 		d = d + r - m
// 	}
// 	if neg {
// 		return -d
// 	}
// 	return d
// }
