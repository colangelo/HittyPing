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

func TestSkillFrontmatterNamesHp(t *testing.T) {
	if !strings.HasPrefix(skillText, "---\nname: hp\ndescription: ") {
		t.Errorf("skills/hp/SKILL.md must open with frontmatter naming hp; got %q", skillText[:min(40, len(skillText))])
	}
}
