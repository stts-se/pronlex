package main

import (
	"context"
	"log/syslog"
	//"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-errors/errors"
	"github.com/gorilla/mux"
	"golang.org/x/net/websocket"

	"github.com/stts-se/pronlex/dbapi"
	"github.com/stts-se/pronlex/lex"
)

func getParam(paramName string, r *http.Request) string {
	res := r.FormValue(paramName)
	if res != "" {
		return res
	}
	vars := mux.Vars(r)
	return vars[paramName]
}

func getLexRefParam(r *http.Request) (lex.LexRef, error) {
	lexRefS := getParam("lexicon_name", r)
	if strings.TrimSpace(lexRefS) == "" {
		msg := "input param <lexicon_name> must not be empty"
		return lex.LexRef{}, errors.New(msg)
	}
	lexRef, err := lex.ParseLexRef(lexRefS)
	if err != nil {
		return lex.LexRef{}, err
	}
	return lexRef, nil
}

func (rout *subRouter) addHandler(handler urlHandler) {
	rout.router.HandleFunc(handler.url, handler.handler)
	rout.handlers = append(rout.handlers, handler)
}

type subRouter struct {
	root     string
	router   *mux.Router
	handlers []urlHandler
	desc     string
}

var subRouters []*subRouter

type urlHandler struct {
	name     string
	handler  func(w http.ResponseWriter, r *http.Request)
	url      string
	help     string
	examples []string
}

// TODO: Neat URL encoding...
func urlEnc(url string) string {
	return strings.Replace(strings.Replace(strings.Replace(strings.Replace(url, " ", "%20", -1), "\n", "", -1), `"`, "%22", -1), "\t", "", -1)
}

func (h urlHandler) helpHTML(root string) string {
	s := "<h2>" + h.name + "</h2> " + h.help
	if strings.Contains(h.url, "{") {
		s = s + `<p>API URL: <code>` + root + h.url + `</code></p>`
	}
	if len(h.examples) > 0 {
		//s = s + `<p>Example invocation:`
		for _, x := range h.examples {
			urlPretty := root + x
			url := prefix + root + urlEnc(x)
			s = s + `<pre><a href="` + url + `">` + urlPretty + `</a></pre>`
		}
		//s = s + "</p>"
	}
	return s
}
func isHandeledPage(url string) bool {
	for _, sub := range subRouters {
		if sub.root == url || sub.root+"/" == url {
			return true
		}
		for _, handler := range sub.handlers {
			if sub.root+handler.url == url {
				return true
			}
		}
	}
	return false
}

var initialSlashRe = regexp.MustCompile("^/")

func removeInitialSlash(url string) string {
	return initialSlashRe.ReplaceAllString(url, "")
}

/*
	func (rout subRouter) handlerExamples() []string {
		res := []string{}
		for _, handler := range rout.handlers {
			for _, example := range handler.examples {
				res = append(res, rout.root+example)
			}
		}
		return res
	}
*/
func newSubRouter(rout *mux.Router, root string, description string) *subRouter {
	var res = subRouter{
		router: rout.PathPrefix(root).Subrouter(),
		root:   root,
		desc:   description,
	}

	helpHandler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		html := "<h1>" + removeInitialSlash(res.root) + "</h1> <em>" + res.desc + "</em>"
		for _, handler := range res.handlers {
			html = html + handler.helpHTML(res.root)
		}
		fmt.Fprint(w, html)
	}

	res.router.HandleFunc("/", helpHandler)
	subRouters = append(subRouters, &res)
	return &res
}

// protect: use this call in handlers to catch 'panic' and stack traces and returning a general error to the calling client
func protect(w http.ResponseWriter) {
	if r := recover(); r != nil {
		defer http.Error(w, "internal server error", http.StatusInternalServerError)
		fmt.Println(errors.Wrap(r, 2).ErrorStack())
		// TODO: log the actual error to a server log file (but do not return to client)
	}
}

// TODO should go into config file
var uploadFileArea string // = ioutil.TempDir("", filepath.Join("lexserver","upload_area"))
// var downloadFileArea string  // = ioutil.TempDir("", filepath.Join("lexserver",""download_area"))
// var symbolSetFileArea string // = filepath.Join(".", "symbol_files")
var dbLocation *string  // = filepath.Join(".", "db_files")
var staticFolder string // = "."

// TODO config stuff
func initFolders() error {
	// TODO sane error handling

	err := os.MkdirAll(filepath.Join(os.TempDir(), "lexserver"), os.ModeDir)
	if err != nil {
		return fmt.Errorf("couldn't initialize temp dir %s : %v", "/tmp/lexserver", err)
	}

	tmpDir, err := ioutil.TempDir("", "lexserver-")
	if err != nil {
		return fmt.Errorf("couldn't initialize temp dir %s : %v", "/tmp/lexserver", err)
	}

	uploadFileArea := filepath.Join(tmpDir, "upload_area")
	downloadFileArea := filepath.Join(tmpDir, "download_area")

	// If the upload area dir doesn't exist, create it
	if _, err := os.Stat(uploadFileArea); err != nil {
		if os.IsNotExist(err) {
			err2 := os.Mkdir(uploadFileArea, 0750)
			if err2 != nil {
				log.Printf("lexserver.init: failed to create %s : %v\n", uploadFileArea, err2)
			}
		} else {
			log.Printf("lexserver.init: peculiar error : %v", err)
		}
	} // else: already exists, hopefully

	// If the download area dir doesn't exist, create it
	if _, err := os.Stat(downloadFileArea); err != nil {
		if os.IsNotExist(err) {
			err2 := os.Mkdir(downloadFileArea, 0750)
			if err2 != nil {
				log.Printf("lexserver.init: failed to create %s : %v\n", downloadFileArea, err2)
			}
		} else {
			log.Printf("lexserver.init: peculiar error : %v", err)
		}
	} // else: already exists, hopefully

	// If the symbol set dir doesn't exist, create it
	// if _, err := os.Stat(symbolSetFileArea); err != nil {
	// 	if os.IsNotExist(err) {
	// 		err2 := os.Mkdir(symbolSetFileArea, 0755)
	// 		if err2 != nil {
	// 			fmt.Printf("lexserver.init: failed to create %s : %v", symbolSetFileArea, err2)
	// 		}
	// 	} else {
	// 		fmt.Printf("lexserver.init: peculiar error : %v", err)
	// 	}
	// } // else: already exists, hopefully

	if *dbEngine == "sqlite" {
		// If the db dir doesn't exist, create it
		if _, err := os.Stat(*dbLocation); err != nil {
			if os.IsNotExist(err) {
				err2 := os.Mkdir(*dbLocation, 0750)
				if err2 != nil {
					log.Printf("lexserver.init: failed to create %s : %v\n", *dbLocation, err2)
				}
			} else {
				log.Printf("lexserver.init: peculiar error : %v", err)
			}
		} // else: already exists, hopefully
	}
	return nil
}

// TODO remove pretty-print option, since you can use the JSONView plugin to Chrome instead
// pretty print if the URL paramer 'pp' has a value
func marshal(v interface{}, r *http.Request) ([]byte, error) {

	if strings.TrimSpace(getParam("pp", r)) != "" {
		return json.MarshalIndent(v, "", "  ")
	}

	return json.Marshal(v)
}

func pingHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "pronlex")
}

func versionHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, strings.Join(vInfo, "\n"))
}

// UTC time with format: yyyy-MM-dd HH:mm:ss z | %Y-%m-%d %H:%M:%S %Z
var startedTimestamp = time.Now().UTC().Format("2006-01-02 15:04:05 MST")

func getVersionInfo() []string {
	res := []string{}
	var buildInfoFile = "/wikispeech/pronlex/build_info.txt"
	if _, err := os.Stat(buildInfoFile); os.IsNotExist(err) {
		//var msg = fmt.Sprintf("lexserver: build info not defined : no such file: %s\n", buildInfoFile)
		//log.Print(msg)
		res = append(res, "Application name: pronlex")
		res = append(res, "Build timestamp: n/a")
		res = append(res, "Built by: user")
		tag, err := exec.Command("git", "describe", "--tags").Output()
		if err != nil {
			log.Printf("lexserver: couldn't retrieve git release info: %v", err)
			res = append(res, "Release: unknown")
		} else {
			branch, err := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD").Output()
			if err != nil {
				log.Printf("lexserver: couldn't retrieve git release info: %v", err)
				res = append(res, "Release: unknown")
			}
			res = append(res, strings.TrimSpace(fmt.Sprintf("Release: %s on branch %s",
				strings.TrimSpace(string(tag)),
				strings.TrimSpace(string(branch)))))
		}
	} else {

		fBytes, err := ioutil.ReadFile(buildInfoFile)
		if err != nil {
			var msg = fmt.Sprintf("lexserver: error reading file : %v", err)
			log.Print(msg)
		}
		if _, err = os.Stat(buildInfoFile); os.IsNotExist(err) {
			var msg = fmt.Sprintf("lexserver: error when reading content from timestamp file : %v", err)
			log.Print(msg)
		} else {
			res = strings.Split(strings.TrimSpace(string(fBytes)), "\n")
		}
	}
	res = append(res, "Started: "+startedTimestamp)
	//log.Println("lexserver: parsed version info", res)
	return res
}

var vInfo []string

func indexHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	html := `<h1>Lexserver</h1>`

	for _, subRouter := range subRouters {
		html = html + `<p><a href="` + prefix + subRouter.root + "/" + `"><b>` + removeInitialSlash(subRouter.root) + `</b></a>`
		html = html + " | " + subRouter.desc + "</p>\n\n"
	}
	html = html + `<p/><hr/><a href="` + prefix + `/version">About Pronlex</a>`
	fmt.Fprint(w, html)
}

// func sqlite3AnalyzeHandler(w http.ResponseWriter, r *http.Request) {
// 	log.Print("lexserver: running the Sqlite3 ANALYZE command...")
// 	_, err := db.Exec("ANALYZE")

// 	if err != nil {
// 		log.Printf("Failed to exec ANALYZE %v", err)
// 		http.Error(w, fmt.Sprintf("/admin/sqlite3_analyze failed : %v", err), http.StatusInternalServerError)
// 		return
// 	}
// 	log.Print("... done!\n")
// 	w.Write([]byte("OK"))
// }

// TODO report unused URL parameters

// TODO Gör konstanter som kan användas istället för strängar
var knownParams = map[string]int{
	"lexicons":            1,
	"entryids":            1,
	"words":               1,
	"lemmas":              1,
	"wordlike":            1,
	"wordregexp":          1,
	"entrystatus":         1,
	"users":               1,
	"wordparts":           1,
	"wordpartslike":       1,
	"wordpartsregexp":     1,
	"transcriptionlike":   1,
	"transcriptionregexp": 1,
	"partofspeechlike":    1,
	"partofspeechregexp":  1,
	"lemmalike":           1,
	"lemmaregexp":         1,
	"readinglike":         1,
	"readingregexp":       1,
	"paradigmlike":        1,
	"paradigmregexp":      1,
	"hasentryvalidation":  1,
	"validationrulelike":  1,
	"validationlevellike": 1,
	"taglike":             1,
	"languagelike":        1,
	"morphologylike":      1,
	"commentlabellike":    1,
	"commentsourcelike":   1,
	"commentlike":         1,
	"multipletags":        1,
	"page":                1,
	"pagelength":          1,
	"pp":                  1,
}

// list of values to the same param splits on comma and/or space
var splitRE = regexp.MustCompile("[, ]+")

func queryFromParams(r *http.Request) (dbapi.DBMQuery, error) {

	lexs := dbapi.RemoveEmptyStrings(
		splitRE.Split(getParam("lexicons", r), -1))
	words := dbapi.RemoveEmptyStrings(
		splitRE.Split(getParam("words", r), -1))
	wordParts := dbapi.RemoveEmptyStrings(
		splitRE.Split(getParam("wordparts", r), -1))
	lemmas := dbapi.RemoveEmptyStrings(
		splitRE.Split(getParam("lemmas", r), -1))
	entryIDStrings := dbapi.RemoveEmptyStrings(
		splitRE.Split(getParam("entryids", r), -1))
	entryIDs := []int64{}
	for _, s := range entryIDStrings {
		id, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return dbapi.DBMQuery{}, fmt.Errorf("couldn't create int64 from input string '%s' : %v", s, err)
		}
		entryIDs = append(entryIDs, id)
	}

	wordLike := strings.TrimSpace(getParam("wordlike", r))
	wordRegexp := strings.TrimSpace(getParam("wordregexp", r))
	wordPartsLike := strings.TrimSpace(getParam("wordpartslike", r))
	wordPartsRegexp := strings.TrimSpace(getParam("wordpartsregexp", r))
	transcriptionLike := strings.TrimSpace(getParam("transcriptionlike", r))
	transcriptionRegexp := strings.TrimSpace(getParam("transcriptionregexp", r))
	partOfSpeechLike := strings.TrimSpace(getParam("partofspeechlike", r))
	partOfSpeechRegexp := strings.TrimSpace(getParam("partofspeechregexp", r))
	languageLike := strings.TrimSpace(getParam("languagelike", r))
	morphologyLike := strings.TrimSpace(getParam("morphologylike", r))
	lemmaLike := strings.TrimSpace(getParam("lemmalike", r))
	lemmaRegexp := strings.TrimSpace(getParam("lemmaregexp", r))
	readingLike := strings.TrimSpace(getParam("readinglike", r))
	readingRegexp := strings.TrimSpace(getParam("readingregexp", r))
	paradigmLike := strings.TrimSpace(getParam("paradigmlike", r))
	paradigmRegexp := strings.TrimSpace(getParam("paradigmregexp", r))
	var entryStatus []string
	if getParam("entrystatus", r) != "" {
		entryStatus = splitRE.Split(getParam("entrystatus", r), -1)
	}
	var users []string
	if getParam("users", r) != "" {
		users = splitRE.Split(getParam("users", r), -1)
	}
	// If true, returns only entries with at least one EntryValidation issue
	hasEntryValidation := false
	if strings.ToLower(getParam("hasentryvalidation", r)) == "true" {
		hasEntryValidation = true
	}
	multipleTags := false
	if strings.ToLower(getParam("multipletags", r)) == "true" {
		multipleTags = true
	}
	validationRuleLike := strings.TrimSpace(getParam("validationrulelike", r))
	validationLevelLike := strings.TrimSpace(getParam("validationlevellike", r))

	tagLike := strings.TrimSpace(getParam("taglike", r))

	commentLabelLike := strings.TrimSpace(getParam("commentlabellike", r))
	commentSourceLike := strings.TrimSpace(getParam("commentsourcelike", r))
	commentLike := strings.TrimSpace(getParam("commentlike", r))

	// TODO report error if getParam("page", r) != ""?
	// Silently sets deafault if no value, or faulty value
	page, err := strconv.ParseInt(getParam("page", r), 10, 64)
	if err != nil {
		page = 0
		//log.Printf("failed to parse page parameter (using default value 0): %v", err)
	}

	// TODO report error if getParam("pagelength", r) != ""?
	// Silently sets deafault if no value, or faulty value
	pageLength, err := strconv.ParseInt(getParam("pagelength", r), 10, 64)
	if err != nil {
		//pageLength = 25
		pageLength = 0
	}

	lexRefs := []lex.LexRef{}
	for _, l := range lexs {
		ref, err := lex.ParseLexRef(l)
		if err != nil {
			return dbapi.DBMQuery{}, fmt.Errorf("couldn't parse lexicon reference from string %s", l)
		}
		lexRefs = append(lexRefs, ref)
	}

	q := dbapi.Query{
		Words:               words,
		WordParts:           wordParts,
		WordPartsLike:       wordPartsLike,
		WordPartsRegexp:     wordPartsRegexp,
		EntryIDs:            entryIDs,
		WordLike:            wordLike,
		WordRegexp:          wordRegexp,
		TranscriptionLike:   transcriptionLike,
		TranscriptionRegexp: transcriptionRegexp,
		PartOfSpeechLike:    partOfSpeechLike,
		PartOfSpeechRegexp:  partOfSpeechRegexp,
		MorphologyLike:      morphologyLike,
		Lemmas:              lemmas,
		LemmaLike:           lemmaLike,
		LemmaRegexp:         lemmaRegexp,
		LanguageLike:        languageLike,
		ReadingLike:         readingLike,
		ReadingRegexp:       readingRegexp,
		ParadigmLike:        paradigmLike,
		ParadigmRegexp:      paradigmRegexp,
		EntryStatus:         entryStatus,
		TagLike:             tagLike,
		CommentLabelLike:    commentLabelLike,
		CommentSourceLike:   commentSourceLike,
		CommentLike:         commentLike,
		Page:                page,
		PageLength:          pageLength,
		HasEntryValidation:  hasEntryValidation,
		MultipleTags:        multipleTags,
		ValidationRuleLike:  validationRuleLike,
		ValidationLevelLike: validationLevelLike,
		Users:               users,
	}

	dq := dbapi.DBMQuery{
		Query:   q,
		LexRefs: lexRefs,
	}
	return dq, nil
}

// Remove initial and trailing " or ' from string
func delQuote(s string) string {
	res := s
	res = strings.TrimPrefix(res, `"`)
	res = strings.TrimPrefix(res, `'`)
	res = strings.TrimSuffix(res, `"`)
	res = strings.TrimSuffix(res, `'`)

	return strings.TrimSpace(res)
}

//var wsChan = make(chan string)

// See https://blog.golang.org/go-maps-in-action "Concurrency"
var webSocks = struct {
	sync.RWMutex
	clients map[string]*websocket.Conn
}{clients: make(map[string]*websocket.Conn)}

func deleteWebSocketClient(id string) {
	webSocks.Lock()
	//defer webSocks.Unlock()
	delete(webSocks.clients, id)
	webSocks.Unlock()
}

// func messageToClientWebSock(clientUUID string, msg string) {
// 	if strings.TrimSpace(clientUUID) != "" {
// 		webSocks.Lock()
// 		if ws, ok := webSocks.clients[clientUUID]; ok {
// 			err := websocket.Message.Send(ws, msg)
// 			log.Printf("websocket send error : %v", err)
// 		} else {
// 			log.Printf("messageToClientWebSock called with unknown UUID string '%s'", clientUUID)
// 		}
// 		webSocks.Unlock()
// 	} else {
// 		log.Printf("messageToClientWebSock called with empty UUID string and message '%s'", msg)
// 	}
// }

func webSockRegHandler(ws *websocket.Conn) {
	var id string
	for {
		var msg string
		err := websocket.Message.Receive(ws, &msg)
		if err != nil {
			log.Printf("webSockRegHandler error : %v\n", err)
			log.Printf("webSockRegHandler removing socket with id %s", id)
			deleteWebSocketClient(id)
			//id = ""
			break
		}

		log.Printf("webSockRegHandler: %s", msg)

		var pref = "CLIENT_ID: "
		//var id string
		if strings.HasPrefix(msg, pref) {
			id = strings.TrimPrefix(strings.TrimSpace(msg), pref)
			webSocks.Lock()
			//defer webSocks.Unlock()
			webSocks.clients[id] = ws
			webSocks.Unlock()
			log.Printf("Processed id: %v", id)
		}

		log.Printf("id is now: %v", id)

		var reply = "HI THERE! " + time.Now().Format("Mon, 02 Jan 2006 15:04:05 PST")
		err = websocket.Message.Send(ws, reply)
		if err != nil {
			log.Printf("webSockRegHandler error : %v\n", err)
			deleteWebSocketClient(id)
			//id = ""
			break
		}
	}
}

func keepClientsAlive() {
	c := time.Tick(67 * time.Second)
	for range c {

		webSocks.Lock()
		//log.Printf("keepClientsAlive: pinging number of clients: %v\n", len(webSocks.clients))
		//defer webSocks.Unlock()
		for client, ws := range webSocks.clients {

			err := websocket.Message.Send(ws, "WS_KEEPALIVE") //+time.Now().Format("Mon, 02 Jan 2006 15:04:05 PST"))
			if err != nil {
				log.Printf("keepClientsAlive got error from websocket send : %v", err)
				delete(webSocks.clients, client)
				log.Printf("keepClientsAlive closed socket to client %s", client)
			}
		}
		webSocks.Unlock()
	}
}

var dbm *dbapi.DBManager
var dbEngine *string

/*
func keepAlive(wsC chan string) {
	c := time.Tick(57 * time.Second)
	for range c {
		wsC <- "WS_KEEPALIVE"
	}
}
*/

/*
func apiChangedHandler(msg string) func(http.ResponseWriter, *http.Request) {

		return func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, fmt.Sprintf("the API has changed: %s\n", msg), http.StatusBadRequest)
			return
		}
	}
*/
func isStaticPage(url string) bool {
	return url == "/" || strings.Contains(url, "externals") || strings.Contains(url, "built") || url == "/websockreg" || url == "/favicon.ico" || url == "/static/" || url == "/ipa_table.txt" || url == "/ping" || url == "/version"
}

var prefix string

func main() {

	port := ":8787"
	testPort := ":8799"
	tag := "standard"
	vInfo = getVersionInfo()

	defaultSqliteLocation := filepath.Join(".", "db_files")
	defaultMariaDBLocation := "speechoid:@tcp(127.0.0.1:3306)"

	var test = flag.Bool("test", false, "run server tests")
	dbEngine = flag.String("db_engine", "sqlite", "db engine (sqlite or mariadb)")
	var maxOpenConns = flag.Int("max_open_conns", 0, "max open connections to one db")
	dbLocation = flag.String("db_location", "", fmt.Sprintf("db location (default \"%s\" for sqlite; \"%s\" for mariadb)", defaultSqliteLocation, defaultMariaDBLocation))
	var logger = flag.String("logger", "stderr", "System `logger` (stderr, syslog or filename)")
	var prefixFlag = flag.String("prefix", "", "Explicit server prefix (e.g. /lexserver)")
	var static = flag.String("static", filepath.Join(".", "static"), "location for static html files")
	var version = flag.Bool("version", false, "print version and exit")
	var help = flag.Bool("help", false, "print usage/help and exit")

	var printUsage = func() {
		fmt.Fprintf(os.Stderr, `Usage:
     $ lexserver [flags] <PORT>
     $ lexserver [flags]
      - use default port

Flags:
`)

		flag.PrintDefaults()

		fmt.Fprintf(os.Stderr, `
Default ports:
     %s  for the standard server
     %s  for the test server
`, port, testPort)
	}

	flag.Usage = func() {
		printUsage()
		os.Exit(0)
	}

	flag.Parse()

	if *version {
		for _, info := range vInfo {
			fmt.Println(info)
		}
		os.Exit(0)
	}

	if *help {
		printUsage()
		os.Exit(1)
	}

	if *test {
		port = testPort
		tag = "test"
	}

	if len(flag.Args()) > 1 {
		printUsage()
		os.Exit(1)
	} else if len(flag.Args()) == 1 {
		port = flag.Args()[0]
	}
	if !strings.HasPrefix(port, ":") {
		port = ":" + port
	}

	if *logger == "stderr" {
		// default logger
	} else if *logger == "syslog" {
		writer, err := syslog.New(syslog.LOG_INFO, "lexserver")
		if err != nil {
			log.Fatalf("Couldn't create logger: %v", err)
		}
		log.SetOutput(writer)
		log.SetFlags(0) // no timestamps etc, since syslog already prints that
	} else {
		f, err := os.OpenFile(*logger, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
		if err != nil {
			log.Fatalf("Couldn't create logger: %v", err)
		}
		defer func() {
			err = f.Close()
			if err != nil {
				log.Fatalf("Couldn't close logger: %v", err)
			}
		}()
		log.SetOutput(f)
	}
	log.Println("lexserver: created logger for " + *logger)

	prefix = *prefixFlag
	if prefix != "" {
		if !strings.HasPrefix(prefix, "/") {
			prefix = "/" + prefix
		}
		// if !strings.HasSuffix(prefix, "/") {
		// 	prefix = prefix + "/"
		// }
	}

	//symbolSetFileArea = *ssFiles
	//dbLocation = *dbFiles
	staticFolder = *static

	var engine dbapi.DBEngine
	if *dbEngine == "sqlite" {
		engine = dbapi.Sqlite
		if *dbLocation == "" {
			dbLocation = &defaultSqliteLocation
		}
	} else if *dbEngine == "mariadb" {
		engine = dbapi.MariaDB
		if *dbLocation == "" {
			dbLocation = &defaultMariaDBLocation
		}
	} else {
		log.Fatalf("Invalid db engine: %s", *dbEngine)
	}
	log.Printf("lexserver: db_location = %s", *dbLocation)

	err := initFolders()
	if err != nil {
		log.Fatal(fmt.Errorf("lexserver: couldn't initialize folders : %v", err))
		os.Exit(1)
	}

	dbm, err = dbapi.NewDBManager(engine)
	if err != nil {
		log.Fatal(fmt.Errorf("lexserver: couldn't initialize db manager : %v", err))
		os.Exit(1)
	}
	dbm.MaxOpenConns = *maxOpenConns
	if engine == dbapi.Sqlite {
		dbapi.Sqlite3WithRegex()
	}

	log.Println("lexserver: started")

	err = setupDemoDB(engine)
	if err != nil {
		log.Printf("COULDN'T INITIALISE DEMO DB : %v\n", err)
		os.Exit(1)
	}

	log.Printf("lexserver: creating %s server on port %s", tag, port)
	s, err := createServer(port)
	if err != nil {
		log.Printf("COULDN'T CREATE SERVER : %v\n", err)
		os.Exit(1)
	}

	if *test {
		err = runInitTests(s, port)
		defer shutdown(s)
		if err != nil {
			log.Printf("lexserver: %v", err)
			os.Exit(1)
		}
	} else { // start the standard server
		stop := make(chan os.Signal, 1)

		signal.Notify(stop, os.Interrupt)
		go func() {
			if err := s.ListenAndServe(); err != nil {
				log.Fatal(fmt.Errorf("lexserver: couldn't start server on port %s : %v", port, err))
			}
		}()
		log.Printf("lexserver: server up and running using port %s", port)

		<-stop

		// This happens after Ctrl-C
		fmt.Fprintf(os.Stderr, "\n")
		shutdown(s)
	}
	log.Println("lexserver: BYE!")
}

func shutdown(s *http.Server) {
	log.Println("lexserver: shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	defer s.Shutdown(ctx)

	// shut down databases nicely
	dbNames, err := dbm.ListDBNames()
	if err != nil {
		log.Printf("couldn't close databases properly, will exit anyway : %v", err)
	}
	for _, dbName := range dbNames {
		err = dbm.CloseDB(dbName)
		if err != nil {
			log.Printf("couldn't close database %s properly, will exit anyway : %v", string(dbName), err)
		}
		log.Printf("lexserver: closed database %s", string(dbName))
	}
	log.Println("lexserver: shutdown completed")
}

func createServer(port string) (*http.Server, error) {

	var s *http.Server

	if !strings.HasPrefix(port, ":") {
		port = ":" + port
	}
	rout := mux.NewRouter().StrictSlash(false)

	var err error // återanvänds för alla fel

	log.Print("lexserver: loading dbs from location ", *dbLocation)
	err = dbm.FirstTimePopulateDBCache(*dbLocation)
	if err != nil {
		return s, err
	}
	// err = loadValidators(symbolSetFileArea)
	// if err != nil {
	// 	return s, fmt.Errorf("failed to load validators : %v", err)
	// }
	// log.Printf("lexserver: loaded validators : %v", validatorNames())

	lexicon := newSubRouter(rout, "/lexicon", "Lexicon management/admin, including full validation")
	lexicon.addHandler(lexiconList)
	lexicon.addHandler(lexiconLookup) // has its own index page in static/
	lexicon.addHandler(lexiconEntriesExist)
	lexicon.addHandler(lexiconInfo)
	lexicon.addHandler(lexiconStats)
	lexicon.addHandler(lexiconListCommentLabels)
	lexicon.addHandler(lexiconListCurrentEntryUsers)
	lexicon.addHandler(lexiconListCurrentEntryStatuses)
	lexicon.addHandler(lexiconListAllEntryStatuses)
	// lexicon.addHandler(lexiconValidationPage)
	// lexicon.addHandler(lexiconValidation)
	lexicon.addHandler(lexiconUpdateEntry)
	lexicon.addHandler(lexiconUpdateValidation)
	lexicon.addHandler(lexiconAddEntry)
	lexicon.addHandler(lexiconDeleteEntry)

	admin := newSubRouter(rout, "/admin", "Misc admin tools")
	admin.addHandler(adminLexImportPage)
	admin.addHandler(adminLexImport)
	admin.addHandler(adminListDBs)
	admin.addHandler(adminCreateDB)
	admin.addHandler(adminDefineLex)
	admin.addHandler(adminMoveNewEntries)
	admin.addHandler(adminDeleteLex)
	// // admin.addHandler(adminSuperDeleteLex)
	admin.addHandler(adminListIDs)

	// Sqlite3 ANALYZE command in some instances make search quicker,
	// but it takes a while to perform. TODO: Re-add this call?
	//rout.HandleFunc("/admin/sqlite3_analyze", sqlite3AnalyzeHandler)

	rout.HandleFunc("/", indexHandler)
	rout.HandleFunc("/ping", pingHandler)
	rout.HandleFunc("/version", versionHandler)
	rout.Handle("/websockreg", websocket.Handler(webSockRegHandler))

	// static
	rout.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join(staticFolder, "favicon.ico"))
	})
	rout.HandleFunc("/ipa_table.txt", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join(staticFolder, "ipa_table.txt"))
	})
	rout.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir(staticFolder))))

	var urls = []string{}
	errW := rout.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		url, err := route.GetPathTemplate()
		if err != nil {
			return err
		}
		if !isStaticPage(url) && !isHandeledPage(url) {
			log.Print("Unhandeled url: ", url)
			urls = append(urls, url+" (UNHANDELED)")
		} else {
			urls = append(urls, url)
		}
		return nil
	})

	if errW != nil {
		log.Printf("server failed to walk through route handlers : %v", err)
	}

	meta := newSubRouter(rout, "/meta", "Meta API calls (list served URLs, etc)")
	meta.addHandler(metaURLsHandler(urls))
	meta.addHandler(metaExamplesHandler)

	// Pinging connected websocket clients
	go keepClientsAlive()

	log.Print("lexserver: server created but not started for port ", port)

	s = &http.Server{
		Addr:           port,
		Handler:        rout,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	return s, nil
}
