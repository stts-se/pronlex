package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/stts-se/pronlex/lex"
)

func runInitTests(s *http.Server, port string) error {

	go func() {
		if err := s.ListenAndServe(); err != nil {
			eFmt := fmt.Sprintf("%v", err)
			if eFmt != "http: Server closed" {
				log.Fatal(fmt.Errorf("init_tests: couldn't start test server on port %s : %v", port, err))
			}
		}
	}()

	log.Println("init_tests: waiting for server to start ...")

	time.Sleep(time.Second * 2) // TODO: This is not beautiful...

	log.Printf("init_tests: server up and running using port %s", port)
	log.Println("init_tests: running tests")

	nErrs1, nTests1, err1 := testExampleURLs(port)
	nErrs2, nTests2, err2 := testURLsWithContent(port)

	var err error
	if err1 != nil && err2 != nil {
		err = fmt.Errorf("%v, %v", err1, err2)
	} else if err1 != nil {
		err = err1
	} else if err2 != nil {
		err = err2
	}
	if err != nil {
		return err
	}

	nTests := nTests1 + nTests2
	testString := "tests"
	if nTests == 1 {
		testString = "test"
	}
	if nErrs1 > 0 || nErrs2 > 0 {
		nErrs := nErrs1 + nErrs2
		errString := "errors"
		if nErrs == 1 {
			errString = "error"
		}
		log.Printf("init_tests: %d %s completed with %d %s!", nTests, testString, nErrs, errString)
		return fmt.Errorf("INIT TESTS FAILED")
	}

	log.Printf("init_tests: %d %s completed without errors", nTests, testString)
	return nil
}

func shortenURL(url string) string {
	limit := 50
	r := []rune(url)
	if len(r) > limit {
		return string(r[0:limit]) + " (...)"
	}
	return url
}

func testURLsWithContent(port string) (int, int, error) {

	log.Println("init_tests: testing some URLs with content")

	nFailed := 0
	nTests := 0

	lookupTests := map[string]string{
		"/lexicon/lookup?lexicons=wikispeech_lexserver_testdb:sv&wordlike=h%C3%A4st__": `[{"id":6,"lexRef":{"dbRef":"wikispeech_lexserver_testdb","lexName":"sv"},"strn":"hästar","language":"sv","partOfSpeech":"NN","morphology":"NEU IND PLU","wordParts":"hästar","lemma":{"id":4,"strn":"häst"},"transcriptions":[{"id":9,"entryId":6,"strn":"\" h E . s t a r","language":"sv"}],"status":{"id":6,"name":"demo","source":"auto","timestamp":"2020-05-25T12:44:47Z","current":true},"preferred":false,"tag":""},{"id":7,"lexRef":{"dbRef":"wikispeech_lexserver_testdb","lexName":"sv"},"strn":"hästar","language":"sv","partOfSpeech":"NN","morphology":"NEU IND PLU","wordParts":"hästar","lemma":{"id":4,"strn":"häst"},"transcriptions":[{"id":10,"entryId":7,"strn":"\" h { . s t a r","language":"sv"}],"status":{"id":7,"name":"demo","source":"auto","timestamp":"2020-05-25T12:44:47Z","current":true},"preferred":false,"tag":""}]`,

		"/lexicon/lookup?lexicons=wikispeech_lexserver_testdb:sv&wordpartsregexp=h%C3%A4st": `[{"id":5,"lexRef":{"dbRef":"wikispeech_lexserver_testdb","lexName":"sv"},"strn":"häst","language":"sv","partOfSpeech":"NN","morphology":"NEU IND SIN","wordParts":"häst","lemma":{"id":4,"strn":"häst"},"transcriptions":[{"id":8,"entryId":5,"strn":"\" h E s t","language":"sv"}],"status":{"id":5,"name":"demo","source":"auto","timestamp":"2020-05-25T12:46:16Z","current":true},"preferred":false,"tag":""},{"id":6,"lexRef":{"dbRef":"wikispeech_lexserver_testdb","lexName":"sv"},"strn":"hästar","language":"sv","partOfSpeech":"NN","morphology":"NEU IND PLU","wordParts":"hästar","lemma":{"id":4,"strn":"häst"},"transcriptions":[{"id":9,"entryId":6,"strn":"\" h E . s t a r","language":"sv"}],"status":{"id":6,"name":"demo","source":"auto","timestamp":"2020-05-25T12:46:16Z","current":true},"preferred":false,"tag":""},{"id":7,"lexRef":{"dbRef":"wikispeech_lexserver_testdb","lexName":"sv"},"strn":"hästar","language":"sv","partOfSpeech":"NN","morphology":"NEU IND PLU","wordParts":"hästar","lemma":{"id":4,"strn":"häst"},"transcriptions":[{"id":10,"entryId":7,"strn":"\" h { . s t a r","language":"sv"}],"status":{"id":7,"name":"demo","source":"auto","timestamp":"2020-05-25T12:46:16Z","current":true},"preferred":false,"tag":""}]`,

		"/lexicon/lookup?lemmas=kex&lexicons=wikispeech_lexserver_testdb:sv": `[{"id":1,"lexRef":{"dbRef":"wikispeech_lexserver_testdb","lexName":"sv"},"strn":"kex","language":"sv","partOfSpeech":"NN","morphology":"NEU IND SIN","wordParts":"kex","lemma":{"id":1,"strn":"kex"},"transcriptions":[{"id":1,"entryId":1,"strn":"\" k e k s","language":"sv"},{"id":2,"entryId":1,"strn":"\" C e k s","language":"sv"}],"status":{"id":1,"name":"demo","source":"auto","timestamp":"2020-05-25T12:46:40Z","current":true},"preferred":false,"tag":""},{"id":2,"lexRef":{"dbRef":"wikispeech_lexserver_testdb","lexName":"sv"},"strn":"kexet","language":"sv","partOfSpeech":"NN","morphology":"NEU DEF SIN","wordParts":"kexet","lemma":{"id":1,"strn":"kex"},"transcriptions":[{"id":3,"entryId":2,"strn":"\" k e k . s @ t","language":"sv"},{"id":4,"entryId":2,"strn":"\" C e k . s @ t","language":"sv"}],"status":{"id":2,"name":"demo","source":"auto","timestamp":"2020-05-25T12:46:40Z","current":true},"preferred":false,"tag":""}]`,

		"/lexicon/lookup?lexicons=wikispeech_lexserver_testdb:sv&words=dom&transcriptionlike=%25o:%25&pp=yes": `[   {     "id": 9,     "lexRef": {       "dbRef": "wikispeech_lexserver_testdb",       "lexName": "sv"     },     "strn": "dom",     "language": "sv",     "partOfSpeech": "NN",     "morphology": "UTR IND SIN",     "wordParts": "dom",     "lemma": {       "id": 5,       "strn": "dom"     },     "transcriptions": [       {         "id": 12,         "entryId": 9,         "strn": "\" d o: m",         "language": "sv"       }     ],     "status": {       "id": 11,       "name": "demo",       "source": "auto",       "timestamp": "2020-05-25T12:47:04Z",       "current": true     },         "preferred": false,     "tag": "building" } ]`}

	jsonMapTests := map[string]string{
		// "/mapper/map/sv-se_ws-sampa-DEMO/sv-se_sampa_mary-DEMO/%22%22%20p%20O%20j%20.%20k%20@": `{"From":"sv-se_ws-sampa-DEMO","To":"sv-se_sampa_mary-DEMO","Input":"\"\" p O j . k @","Result":"\" p O j - k @"}`,
		// "/mapper/map/sv-se_sampa_mary-DEMO/sv-se_ws-sampa-DEMO/%22%20p%20O%20j%20-%20k%20@":    `{"From":"sv-se_sampa_mary-DEMO","To":"sv-se_ws-sampa-DEMO","Input":"\" p O j - k @","Result":"\"\" p O j . k @"}`,
		//`/validation/validateentry?symbolsetname=sv-se_ws-sampa-DEMO&entry={%22id%22:371546,%22lexiconId%22:1,%22strn%22:%22h%C3%A4st%22,%22language%22:%22SWE%22,%22partOfSpeech%22:%22NN%20SIN|IND|NOM|UTR%22,%22wordParts%22:%22h%C3%A4st%22,%22lemma%22:{%22id%22:42815,%22strn%22:%22h%C3%A4st%22,%22reading%22:%22%22,%22paradigm%22:%22s2q-lapp%22},%22transcriptions%22:[{%22id%22:377191,%22entryId%22:371546,%22strn%22:%22\%22%20h%20E%20s%20t%22,%22language%22:%22SWE%22,%22sources%22:[]}],%22status%22:{%22id%22:371546,%22name%22:%22imported%22,%22source%22:%22nst%22,%22timestamp%22:%222016-09-06T12:54:12Z%22,%22current%22:true}}`: `{"id":371546,"lexRef":{},"strn":"häst","language":"SWE","partOfSpeech":"NN SIN|IND|NOM|UTR","wordParts":"häst","lemma":{"id":42815,"strn":"häst","paradigm":"s2q-lapp"},"transcriptions":[{"id":377191,"entryId":371546,"strn":"\" h E s t","language":"SWE"}],"status":{"id":371546,"name":"imported","source":"nst","timestamp":"2016-09-06T12:54:12Z","current":true}}`,
	}

	jsonListTestsMustContain := map[string][]string{
		"/admin/list_dbs": {"wikispeech_lexserver_testdb"},
		// "/mapper/list":     {"sv-se_ws-sampa-DEMO - sv-se_sampa_mary-DEMO", "sv-se_sampa_mary-DEMO - sv-se_ws-sampa-DEMO"},
		// "/symbolset/list":  {"sv-se_sampa_mary-DEMO", "sv-se_ws-sampa-DEMO", "sv-se_nst-xsampa-DEMO"},
		// "/validation/list": {"sv-se_ws-sampa-DEMO"},
	}

	jsonBoolTests := map[string]bool{
		// "/validation/has_validator/sv-se_ws-sampa-DEMO": true,
		// "/validation/has_validator/ar_ws-sampa-DEMO":    false,
	}

	mustExistTests := []string{
		"/ipa_table.txt",
	}

	log.Printf("init_tests: testing 200 status: %d", len(mustExistTests))
	for _, url := range mustExistTests {
		nTests = nTests + 1
		ok, err := mustExistTest(port, url)
		if !ok {
			nFailed = nFailed + 1
		}
		if err != nil {
			return nFailed, nTests, err
		}
	}

	log.Printf("init_tests: testing entry lookup: %d", len(lookupTests))
	for url, expect := range lookupTests {
		nTests = nTests + 1
		ok, err := lookupTest(port, url, expect)
		if !ok {
			nFailed = nFailed + 1
		}
		if err != nil {
			return nFailed, nTests, err
		}
	}

	log.Printf("init_tests: testing json map results: %d", len(jsonMapTests))
	for url, expect := range jsonMapTests {
		nTests = nTests + 1
		ok, err := jsonMapTest(port, url, expect)
		if !ok {
			nFailed = nFailed + 1
		}
		if err != nil {
			return nFailed, nTests, err
		}
	}

	log.Printf("init_tests: testing json list results: %d", len(jsonListTestsMustContain))
	for url, expect := range jsonListTestsMustContain {
		nTests = nTests + 1
		ok, err := jsonListTestMustContain(port, url, expect)
		if !ok {
			nFailed = nFailed + 1
		}
		if err != nil {
			return nFailed, nTests, err
		}
	}

	log.Printf("init_tests: testing boolean results: %d", len(jsonBoolTests))
	for url, expect := range jsonBoolTests {
		nTests = nTests + 1
		ok, err := jsonTestBool(port, url, expect)
		if !ok {
			nFailed = nFailed + 1
		}
		if err != nil {
			return nFailed, nTests, err
		}
	}

	return nFailed, nTests, nil
}

func jsonMapTest(port string, url string, expect string) (bool, error) {
	url = "http://localhost" + port + url
	/* #nosec G107 */
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("** FAILED TEST ** for %s : couldn't retrieve URL : %v\n", url, err)
		return false, nil
	}
	defer resp.Body.Close()
	log.Printf("init_tests: jsonMap %s", url)

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("** FAILED TEST ** for %s : expected response code 200, found %d\n", url, resp.StatusCode)
		return false, nil
	}

	got, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return false, fmt.Errorf("couldn't read response body : %v", err)
	}

	var gotJ map[string]interface{}
	err = json.Unmarshal([]byte(got), &gotJ)
	if err != nil {
		return false, fmt.Errorf("couldn't convert response to json : %v", err)
	}

	var expJ map[string]interface{}
	err = json.Unmarshal([]byte(expect), &expJ)
	if err != nil {
		return false, fmt.Errorf("couldn't convert expected result to json : %v", err)
	}

	if !reflect.DeepEqual(gotJ, expJ) {
		fmt.Printf("** FAILED TEST ** for %s :\n >> EXPECTED RESPONSE:\n%s\n >> FOUND:\n%s\n", url, expect, string(got))
		return false, nil
	}
	return true, nil
}

func jsonTestBool(port string, url string, expect bool) (bool, error) {
	url = "http://localhost" + port + url
	/* #nosec G107 */
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("** FAILED TEST ** for %s : couldn't retrieve URL : %v\n", url, err)
		return false, nil
	}
	defer resp.Body.Close()
	log.Printf("init_tests: jsonMap %s", url)

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("** FAILED TEST ** for %s : expected response code 200, found %d\n", url, resp.StatusCode)
		return false, nil
	}

	got, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return false, fmt.Errorf("couldn't read response body : %v", err)
	}

	var gotJ bool
	err = json.Unmarshal([]byte(got), &gotJ)
	if err != nil {
		return false, fmt.Errorf("couldn't convert response to json : %v", err)
	}
	if gotJ != expect {
		fmt.Printf("** FAILED TEST ** for %s :\n >> EXPECTED RESPONSE: %v\n >> FOUND: %v\n", url, expect, gotJ)
		return false, nil
	}
	return true, nil
}

func contains(slice []string, item string) bool {
	for _, a := range slice {
		if a == item {
			return true
		}
	}
	return false
}

func jsonListTestMustContain(port string, url string, expect []string) (bool, error) {
	url = "http://localhost" + port + url
	/* #nosec G107 */
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("** FAILED TEST ** for %s : couldn't retrieve URL : %v\n", url, err)
		return false, nil
	}
	defer resp.Body.Close()
	log.Printf("init_tests: jsonList %s", url)

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("** FAILED TEST ** for %s : expected response code 200, found %d\n", url, resp.StatusCode)
		return false, nil
	}

	got, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return false, fmt.Errorf("couldn't read response body : %v", err)
	}

	var gotJ []string
	err = json.Unmarshal([]byte(got), &gotJ)
	if err != nil {
		return false, fmt.Errorf("couldn't convert response to json : %v", err)
	}

	ok := true
	for _, exp := range expect {
		if !contains(gotJ, exp) {
			ok = false
		}
	}
	if !ok {
		fmt.Printf("** FAILED TEST ** for %s :\n >> EXPECTED RESPONSE TO CONTAIN ALL OF:\n%s\n >> FOUND:\n%s\n", url, expect, string(got))
		return false, nil
	}

	return true, nil
}

func lookupTest(port string, url string, expect string) (bool, error) {
	url = "http://localhost" + port + url
	/* #nosec G107 */
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("** FAILED TEST ** for %s : couldn't retrieve URL : %v\n", url, err)
		return false, nil
	}
	defer resp.Body.Close()
	log.Printf("init_tests: lookup/entry %s", url)

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("** FAILED TEST ** for %s : expected response code 200, found %d\n", url, resp.StatusCode)
		return false, nil
	}

	got, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return false, fmt.Errorf("couldn't read response body : %v", err)
	}

	var gotEs []lex.Entry
	err = json.Unmarshal(got, &gotEs)
	if err != nil {
		return false, fmt.Errorf("couldn't parse json : %v", err)
	}
	var expEs []lex.Entry
	err = json.Unmarshal([]byte(expect), &expEs)
	if err != nil {
		return false, fmt.Errorf("couldn't parse expect string : %v", err)
	}
	for i, e := range gotEs {
		e.EntryStatus.Timestamp = ""
		gotEs[i] = e
	}
	for i, e := range expEs {
		e.EntryStatus.Timestamp = ""
		expEs[i] = e
	}

	if !reflect.DeepEqual(gotEs, expEs) {
		fmt.Println("EXP", expEs)
		fmt.Println("GOT", gotEs)

		fmt.Printf("** FAILED TEST ** for %s :\n >> EXPECTED RESPONSE:\n%s\n >> FOUND:\n%s\n", url, expect, string(got))
		return false, nil
	}
	return true, nil
}

func mustExistTest(port string, url string) (bool, error) {
	url = "http://localhost" + port + url
	/* #nosec G107 */
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("** FAILED TEST ** for %s : couldn't retrieve URL : %v\n", url, err)
		return false, nil
	}
	defer resp.Body.Close()
	log.Printf("init_tests: lookup/entry %s", url)

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("** FAILED TEST ** for %s : expected response code 200, found %d\n", url, resp.StatusCode)
		return false, nil
	}
	return true, nil

}

func testExampleURLs(port string) (int, int, error) {

	log.Println("init_tests: testing response codes for built-in example URLs")

	nFailed := 0
	nTests := 0

	/* #nosec G107 */
	resp, err := http.Get("http://localhost" + port + "/meta/examples")
	if err != nil {
		return nFailed, nTests, fmt.Errorf("couldn't retrieve server's url examples : %v", err)
	}
	defer resp.Body.Close()
	js, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nFailed, nTests, fmt.Errorf("couldn't retrieve server's url examples : %v", err)
	}

	var res []JSONURLExample
	err = json.Unmarshal([]byte(js), &res)
	if err != nil {
		return nFailed, nTests, fmt.Errorf("couldn't unmarshal json : %v", err)
	}

	for _, example := range res {
		nTests = nTests + 1
		url := "http://localhost" + port + urlEnc(example.URL)
		/* #nosec G107 */
		resp, err = http.Get(url)
		if err != nil {
			fmt.Printf("** FAILED TEST ** for %s : couldn't retrieve URL : %v\n", url, err)
			nFailed = nFailed + 1
		}
		defer resp.Body.Close()
		log.Printf("init_tests: %s => %s : %s", example.Template, shortenURL(example.URL), resp.Status)

		if resp.StatusCode != http.StatusOK {
			fmt.Printf("** FAILED TEST ** for %s : expected response code 200, found %d\n", url, resp.StatusCode)
			nFailed = nFailed + 1
		} else {
			got, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return nFailed, nTests, fmt.Errorf("couldn't read response body : %v", err)
			}

			if strings.TrimSpace(string(got)) == "" {
				fmt.Printf("** FAILED TEST ** for %s : expected non-empty response\n", url)
				nFailed = nFailed + 1
			}
		}
	}

	return nFailed, nTests, nil
}
