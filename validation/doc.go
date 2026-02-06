// Package validation is used to define entry validators and rules. It contains a few general rule types (see validation/rules/rules.go), and you can also create new rules using the Rule interface.
//
// To create a validating rule suite, initialize the Validator struct using a slice of Rule instances.
// Use the Validator.Validate or ValidateEntry functions to have one or more entries validated. Implemented validators are found in sub package validation/validators.
package validation
