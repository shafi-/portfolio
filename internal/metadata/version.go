package metadata

import "strings"

// versionSpecKinds are the literal constraint operators a manifest may prefix a
// version with, longest first so multi-char operators match before their
// single-char prefixes (e.g. ">=" before ">", "~>" before "~").
var versionSpecKinds = []string{">=", "<=", "~>", "==", "!=", "^", "~", ">", "<", "="}

// parseVersionSpec splits a dependency version specifier (exactly as declared in
// a manifest) into its value and the literal constraint kind. This is a literal
// decomposition, not a semantic one — Cargo's implicit caret (the "1.0" in
// serde = "1.0") is recorded as "exact" because the manifest pins it literally;
// interpreting it as a caret is the AI agent's job, not the engine's.
//
// The returned kind is one of:
//   - the leading operator verbatim ("^", "~", "~>", ">=", "<=", "==", "!=", "=", ">", "<")
//   - "exact"  for a bare pinned version (no operator), e.g. "4.0.0", "v1.2.3"
//   - "range"  for a compound spec, e.g. ">=1.0,<2.0", ">=1.0 <2.0"
//   - "any"    for "*", "latest", and x-ranges, e.g. "1.x", "3.+"
//   - ""       when the spec is empty/unknown
//
// version holds the operator-stripped value when kind is an operator or "exact";
// for "range"/"any" it holds the whole declared spec (there is no single value).
func parseVersionSpec(raw string) (version, kind string) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return "", ""
	}

	for _, op := range versionSpecKinds {
		if strings.HasPrefix(s, op) {
			rest := strings.TrimSpace(strings.TrimPrefix(s, op))
			// A second operator or comma means a compound range, e.g.
			// ">=1.0,<2.0" — there is no single version value to extract.
			if rest == "" || containsConstraint(rest) {
				return s, "range"
			}
			return rest, op
		}
	}

	// No leading operator.
	if s == "*" || s == "latest" {
		return s, "any"
	}
	if strings.Contains(s, ",") || strings.Contains(s, " - ") {
		// comma-separated or hyphen range, e.g. "1.0,2.0", "1.0 - 2.0"
		return s, "range"
	}
	if strings.ContainsAny(s, "*xX+") {
		// x-range / wildcard, e.g. "1.x", "3.+", "1.*"
		return s, "any"
	}
	return s, "exact"
}

// containsConstraint reports whether s holds a second version constraint
// operator or a comma, indicating a compound range after the first operator was
// stripped (e.g. rest "1.0,<2.0" or "1.0 <2.0" from ">=1.0,<2.0").
func containsConstraint(s string) bool {
	if strings.ContainsAny(s, "<>=~^") {
		return true
	}
	return strings.Contains(s, ",")
}

// splitNameSpec splits a flat dependency line (e.g. a requirements.txt entry)
// into the distribution name and the trailing version spec, found at the first
// constraint-operator character. "django==4.0" -> ("django", "==4.0");
// "django>=4.0" -> ("django", ">=4.0"); "flask" -> ("flask", ""). It does not
// strip PEP 508 extras — callers do that on the returned name.
func splitNameSpec(line string) (name, spec string) {
	if i := strings.IndexAny(line, "=<>!~"); i != -1 {
		return line[:i], line[i:]
	}
	return line, ""
}

// extractTableVersion reads the declared version spec from the right-hand side
// of a TOML dependency assignment. It handles both the string form
// (`requests = "^2.28.0"` → rhs `"^2.28.0"`) and the inline-table form used by
// Cargo and Poetry (`tokio = { version = "1.0", features = [...] }`). It returns
// the raw spec string (still including any operator); callers pass it through
// parseVersionSpec to split value from kind.
func extractTableVersion(rhs string) string {
	s := strings.TrimSpace(rhs)
	if s == "" {
		return ""
	}
	if strings.HasPrefix(s, "{") {
		// Inline-table form: pull the value of the inner "version" key.
		i := strings.Index(s, "version")
		if i < 0 {
			return ""
		}
		rest := s[i+len("version"):]
		eq := strings.Index(rest, "=")
		if eq < 0 {
			return ""
		}
		rest = rest[eq+1:]
		q := strings.IndexAny(rest, "\"'")
		if q < 0 {
			return ""
		}
		quote := rest[q]
		end := strings.IndexByte(rest[q+1:], quote)
		if end < 0 {
			return ""
		}
		return rest[q+1 : q+1+end]
	}
	// String form: strip a single pair of surrounding quotes.
	if len(s) >= 2 && (s[0] == '"' || s[0] == '\'') && s[len(s)-1] == s[0] {
		return s[1 : len(s)-1]
	}
	return s
}
