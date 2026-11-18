package main

import (
	"strings"
	"testing"

	flag "github.com/spf13/pflag"
)

func TestSkillDocumentsEveryFlag(t *testing.T) {
	fs := flag.NewFlagSet("hp", flag.ContinueOnError)
	defineFlags(fs)
	fs.VisitAll(func(f *flag.Flag) {
		if !strings.Contains(skillText, "`--"+f.Name+"`") {
			t.Errorf("skills/hp/SKILL.md does not document --%s", f.Name)
		}
		if f.Shorthand != "" && !strings.Contains(skillText, "`-"+f.Shorthand+"`, `--"+f.Name+"`") {
			t.Errorf("skills/hp/SKILL.md does not pair -%s with --%s", f.Shorthand, f.Name)
		}
	})
}

// The description is a folded block scalar: a plain one-line value holding
// ": " is invalid YAML, which strict loaders (Pi, pyyaml) reject.
func TestSkillFrontmatterNamesHp(t *testing.T) {
	const head = "---\nname: hp\ndescription: >-\n"
	if !strings.HasPrefix(skillText, head) {
		t.Fatalf("skills/hp/SKILL.md must open with %q; got %q", head, skillText[:min(60, len(skillText))])
	}
	block, _, found := strings.Cut(skillText[len(head):], "\n---\n")
	if !found {
		t.Fatal("skills/hp/SKILL.md frontmatter has no closing ---")
	}
	var words []string
	for _, line := range strings.Split(block, "\n") {
		if !strings.HasPrefix(line, "  ") {
			t.Fatalf("description line %q is not indented under the block scalar", line)
		}
		words = append(words, strings.TrimSpace(line))
	}
	if n := len(strings.Join(words, " ")); n > 1024 {
		t.Errorf("description is %d characters; Pi rejects more than 1024", n)
	}
}
