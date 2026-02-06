package validation

import (
	"fmt"
	"strings"
	"sync"

	"github.com/stts-se/pronlex/lex"
)

/*
Result is a validation result with the following fields:

	RuleName - arbitrary string
	Level - typically indicating severity (e.g. Info/Warning/Fatal/Format/Severe)
	Messages - arbitrary strings representing validation messages to the user
*/
type Result struct {
	RuleName string
	Level    string
	Messages []string
}

// Strings returns a slice of simple string representations of the messages in Result, including information on rule name and rule level
func (r Result) Strings() []string {
	var res = make([]string, 0)
	for _, msg := range r.Messages {
		res = append(res, fmt.Sprintf("%s (%s): %s", r.RuleName, r.Level, msg))
	}
	return res
}

// Rule interface. To create a validation.Rule, make a struct implementing Validate, ShouldAccept and ShouldReject as defined in this interface.
type Rule interface {
	Validate(lex.Entry) (Result, error)
	ShouldAccept() []lex.Entry
	ShouldReject() []lex.Entry
	Name() string
	Level() string
}

func ToString(r Rule) string {
	fs := []string{
		r.Name(),
		r.Level(),
	}
	return strings.Join(fs, "\t")
}

// Validator is a struct containing a slice of rules
type Validator struct {
	Name  string
	Rules []Rule
}

func (v Validator) NumberOfTests() int {
	a, r := v.AllTests()
	return len(a) + len(r)
}

func (v Validator) AllTests() ([]lex.Entry, []lex.Entry) {
	accept := []lex.Entry{}
	reject := []lex.Entry{}
	for _, r := range v.Rules {
		accept = append(accept, r.ShouldAccept()...)
	}
	for _, r := range v.Rules {
		reject = append(reject, r.ShouldReject()...)
	}
	return accept, reject
}

func (v Validator) String() string {
	name := "<noname>"
	if v.Name != "" {
		name = v.Name
	}
	fs := []string{"Validator: " + name}
	fs = append(fs, fmt.Sprintf("# rules: %d", len(v.Rules)))
	for _, r := range v.Rules {
		fs = append(fs, fmt.Sprintf("Rule: %s", ToString(r)))
	}
	return strings.Join(fs, "\n")
}

// IsDefined is used to check if the validator is initialized (by checking that the validator has a non-empty name).
func (v Validator) IsDefined() bool {
	return v.Name != ""
}

// TestResultContainer is a container class for accept/reject/crosscheck test result
type TestResultContainer struct {
	AcceptErrors []TestResult
	RejectErrors []TestResult
	CrossErrors  []TestResult
}

// Size returns the total number of errors in the TestResultContainer
func (tc TestResultContainer) Size() int {
	return len(tc.AcceptErrors) + len(tc.RejectErrors) + len(tc.CrossErrors)
}

// AllErrors returns all errors in the TestResultContainer
func (tc TestResultContainer) AllErrors() []TestResult {
	return append(append(tc.AcceptErrors, tc.RejectErrors...), tc.CrossErrors...)
}

// TestResult holds the test result for a tested rule suite (accept, reject, or cross tests result)
type TestResult struct {
	RuleName string
	Level    string
	Messages []string
	Input    lex.Entry
}

func (tr TestResult) String() string {
	msg := fmt.Sprintf("[%s|%s] %s", tr.RuleName, tr.Level, strings.Join(tr.Messages, "; "))
	return fmt.Sprintf("%s\nInput: %v", msg, tr.Input)
}

type acceptExample struct {
	RuleName string
	Level    string
	Entry    lex.Entry
}

// RunTests runs accept/reject tests for all individual rules, and cross checks all accept tests against the other rules
func (v Validator) RunTests() (TestResultContainer, error) {
	var result TestResultContainer
	var allAccept []acceptExample
	for _, rule := range v.Rules {
		for _, e := range rule.ShouldAccept() {
			res, err := rule.Validate(e)
			allAccept = append(allAccept, acceptExample{RuleName: res.RuleName, Level: res.Level, Entry: e})
			if err != nil {
				return result, err
			}
			var messages []string
			for _, msg := range res.Messages {
				messages = append(messages,
					fmt.Sprintf("Accept example was rejected for rule %s (%s). Message: %s", res.RuleName, res.Level, msg))
			}
			if len(messages) > 0 {
				result.AcceptErrors = append(result.AcceptErrors,
					TestResult{RuleName: res.RuleName, Level: res.Level, Messages: messages, Input: e})
			}
		}
		for _, e := range rule.ShouldReject() {
			res, err := rule.Validate(e)
			if err != nil {
				return result, err
			}
			if len(res.Messages) == 0 {
				messages := []string{fmt.Sprintf("Reject example was accepted for rule %s (%s)", res.RuleName, res.Level)}
				result.RejectErrors = append(result.RejectErrors,
					TestResult{RuleName: res.RuleName, Level: res.Level, Messages: messages, Input: e})
			}
		}
	}

	for _, accept := range allAccept {
		for _, rule := range v.Rules {
			if strings.ToLower(rule.Level()) == "info" {
				continue
			}

			// TODO: no need to test the rule's own accept examples, that is already taken care of above
			res, err := rule.Validate(accept.Entry)
			if err != nil {
				return result, err
			}
			var messages []string
			for _, msg := range res.Messages {
				messages = append(messages,
					fmt.Sprintf("Accept example for rule %s (%s) was rejected by rule %s (%s). Message: %s", accept.RuleName, accept.Level, res.RuleName, res.Level, msg))
			}
			if len(messages) > 0 {
				result.CrossErrors = append(result.CrossErrors,
					TestResult{RuleName: res.RuleName, Level: res.Level, Messages: messages, Input: accept.Entry})
			}
		}

	}

	return result, nil
}

/*
func (v Validator) validateEntryOLD(e lex.Entry) (lex.Entry, bool) {
	e.EntryValidations = make([]lex.EntryValidation, 0)
	for _, rule := range v.Rules {
		res, err := rule.Validate(e)
		if err != nil {
			var ev = lex.EntryValidation{
				RuleName: "System",
				Level:    "Error",
				Message:  fmt.Sprintf("error when validating word '%s' with rule %s : %v", e.Strn, res.RuleName, err),
			}
			e.EntryValidations = append(e.EntryValidations, ev)
		} else {
			for _, msg := range res.Messages {
				var ev = lex.EntryValidation{
					RuleName: res.RuleName,
					Level:    res.Level,
					Message:  msg,
				}
				e.EntryValidations = append(e.EntryValidations, ev)
			}
		}
	}
	return e, len(e.EntryValidations) == 0
}

*/

// ValidateEntry is used to validate single entries. Any validation
// errors are added to the entry's EntryValidations field.
func (v Validator) ValidateEntry(e *lex.Entry) {
	e.EntryValidations = make([]lex.EntryValidation, 0)
	for _, rule := range v.Rules {
		res, err := rule.Validate(*e)
		if err != nil {
			var ev = lex.EntryValidation{
				RuleName: "System",
				Level:    "Error",
				Message:  fmt.Sprintf("error when validating word '%s' with rule %s : %v", e.Strn, res.RuleName, err),
			}
			e.EntryValidations = append(e.EntryValidations, ev)
		} else {
			for _, msg := range res.Messages {
				var ev = lex.EntryValidation{
					RuleName: res.RuleName,
					Level:    res.Level,
					Message:  msg,
				}
				e.EntryValidations = append(e.EntryValidations, ev)
			}
		}
	}
}

func (v Validator) validateEntriesConcurrent(entries []lex.Entry) ([]lex.Entry, bool) {
	var wg sync.WaitGroup
	var work = []*lex.Entry{}
	for _, e := range entries {
		ex := e
		work = append(work, &ex)
	}
	for _, e := range work {
		wg.Add(1)
		go func(ee *lex.Entry) {
			defer wg.Done()
			v.ValidateEntry(ee)
		}(e)
	}

	wg.Wait()

	valid := true
	var res = []lex.Entry{}
	for _, e := range work {
		res = append(res, *e)
		if len(e.EntryValidations) > 0 {
			valid = false
		}
	}
	return res, valid
}

// ValidateEntries is used to validate a slice of entries.  Any validation
// errors are added to each entry's EntryValidations field. The
// function returns true if the entry is valid (i.e., no validation
// issues are found), otherwise false.
func (v Validator) ValidateEntries(entries []lex.Entry) ([]lex.Entry, bool) {
	return v.validateEntriesConcurrent(entries)
}

/*
func (v Validator) validateEntriesOLD(entries []lex.Entry) ([]lex.Entry, bool) {
	var res []lex.Entry
	var valid = true
	for _, e0 := range entries {
		var e, ok = v.validateEntryOLD(e0)
		if !ok {
			valid = false
		}
		res = append(res, e)
	}
	return res, valid
}
*/
