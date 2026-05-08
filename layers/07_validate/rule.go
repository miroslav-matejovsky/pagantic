package validate

// Rule is one named check.
type Rule struct {
	// Name says which rule ran.
	Name string
	// Check tests output. Return error on fail.
	Check func(output string) error
}

// RuleValidator runs deterministic rules on output.
type RuleValidator struct {
	// Rules holds checks to run.
	Rules []Rule
}

// NewRuleValidator creates validator with given rules.
func NewRuleValidator(rules ...Rule) *RuleValidator {
	copied := append([]Rule(nil), rules...)
	return &RuleValidator{Rules: copied}
}

// Validate runs all rules. It returns only failures.
func (rv *RuleValidator) Validate(output string) []error {
	if rv == nil {
		return nil
	}

	var errs []error
	for _, rule := range rv.Rules {
		if rule.Check == nil {
			continue
		}
		if err := rule.Check(output); err != nil {
			errs = append(errs, err)
		}
	}

	return errs
}
