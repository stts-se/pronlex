package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	//"os"
	"encoding/json"
	"path/filepath"
	"strings"

	"github.com/stts-se/pronlex/symbolset"
)

// TODO Mutex this variable
var mapperService = symbolset.MapperService{}

func mapMapperHandler(w http.ResponseWriter, r *http.Request) {
	fromName := r.FormValue("from")
	toName := r.FormValue("to")
	trans := r.FormValue("trans")
	result, err := mapperService.Map(fromName, toName, trans)
	if err != nil {
		msg := fmt.Sprintf("failed mapping transcription : %v", err)
		log.Println(msg)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}
	fmt.Fprint(w, result)
}

func symbolSetMapperHandler(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	result, ok := mapperService.SymbolSets[name]
	if !ok {
		msg := fmt.Sprintf("failed getting symbol set : %v", name)
		log.Println(msg)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}
	fmt.Fprint(w, result)
}

func mapTableMapperHandler(w http.ResponseWriter, r *http.Request) {
	fromName := r.FormValue("from")
	toName := r.FormValue("to")
	result, err := mapperService.GetMapTable(fromName, toName)
	if err != nil {
		msg := fmt.Sprintf("failed getting map table : %v", err)
		log.Println(msg)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}
	fmt.Fprint(w, result)
}

func loadMapperHandler(w http.ResponseWriter, r *http.Request) {
	// list files in symbol set dir
	fileInfos, err := ioutil.ReadDir(symbolSetFileArea)
	if err != nil {
		msg := fmt.Sprintf("failed reading symbol set dir : %v", err)
		log.Println(msg)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}

	var fErrs error
	var symSets []symbolset.SymbolSet
	for _, fi := range fileInfos {
		if strings.HasSuffix(fi.Name(), ".tab") {
			symset, err := symbolset.LoadSymbolSet(filepath.Join(symbolSetFileArea, fi.Name()))
			if err != nil {
				if fErrs != nil {
					fErrs = fmt.Errorf("%v : %v", fErrs, err)
				} else {
					fErrs = err
				}
			} else {
				symSets = append(symSets, symset)
			}
		}
	}

	//msg := fmt.Sprintf("failed to load symbol set file : %v", err)
	//http.Error(w, msg, http.StatusInternalServerError)
	//return

	if fErrs != nil {
		msg := fmt.Sprintf("failed to load symbol set : %v", fErrs)
		log.Println(msg)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}

	// TODO nuke symbolSets and replace by symbolSetsMap
	//symbolSets = symSets
	var symbolSetsMap = make(map[string]symbolset.SymbolSet)
	for _, z := range symSets {
		// TODO check that x.Name doesn't already exist
		symbolSetsMap[z.Name] = z
	}
	mapperService.SymbolSets = symbolSetsMap

	j, err := json.Marshal(symbolSetNames(symbolSetsMap))
	if err != nil {
		msg := fmt.Sprintf("json marshalling error : %v", err)
		log.Println(msg)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}

	fmt.Fprint(w, string(j))
}

func listMapperHandler(w http.ResponseWriter, r *http.Request) {
	ss := symbolSetNames(mapperService.SymbolSets)
	j, err := json.Marshal(ss)
	if err != nil {
		msg := fmt.Sprintf("failed to marshal struct : %v", err)
		log.Println(msg)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	fmt.Fprint(w, string(j))
}

type SymbolSetNames struct {
	SymbolSetNames []string `json:symbol_set_names`
}

func symbolSetNames(sss map[string]symbolset.SymbolSet) SymbolSetNames {
	var ssNames []string
	for ss, _ := range sss {
		ssNames = append(ssNames, ss)
	}
	return SymbolSetNames{SymbolSetNames: ssNames}
}

// func symbolSetNames(sss []symbolset.SymbolSet) SymbolSetNames {
// 	var ssNames []string
// 	for _, ss := range sss {
// 		ssNames = append(ssNames, ss.Name)
// 	}
// 	return SymbolSetNames{SymbolSetNames: ssNames}
// }
