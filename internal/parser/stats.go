package parser

import (
	"fmt"
	"sort"
	"strings"
)

func PrettyPrintStats(stats *Stats) string {
	rules := make([]string, 0, len(stats.ChoiceAltCnt))
	for r := range stats.ChoiceAltCnt {
		rules = append(rules, r)
	}
	sort.Strings(rules)
	buf := &strings.Builder{}
	for _, r := range rules {
		fmt.Fprintf(buf, "%s ", r)
		for choice, count := range stats.ChoiceAltCnt[r] {
			fmt.Fprintf(buf, "| %s->%dx ", choice, count)
		}
		buf.WriteString("\n")
	}
	fmt.Fprintf(buf, "---------------\nTotal: %d\n\n", stats.ExprCnt)
	return buf.String()
}
