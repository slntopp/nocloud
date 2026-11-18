package graph

import (
	"fmt"
	"strings"
)

// accountGradeFilter turns the "grade" accounts filter into an AQL condition.
// Grade thresholds live in the admin UI settings, so the client sends the
// conditions for the selected grades:
//
//	[{"override": "A"}, {"from": 900}, {"from": 300, "to": 900}]
//
// "override" matches a manually set grade. "from"/"to" match the counted
// spending (from inclusive, to exclusive) of accounts without a manual grade.
// Conditions are OR-ed. Returns "" when there is nothing to filter by.
func accountGradeFilter(items []any) (string, map[string]any) {
	var conds []string
	vars := map[string]any{}

	for i, item := range items {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}

		if title, ok := m["override"].(string); ok {
			title = strings.TrimSpace(title)
			if title == "" {
				continue
			}
			name := fmt.Sprintf("grade_override_%d", i)
			vars[name] = title
			conds = append(conds, fmt.Sprintf("node.grade_override == @%s", name))
			continue
		}

		parts := []string{
			`(IS_NULL(node.grade_override) || node.grade_override == "")`,
			"IS_NUMBER(node.spending.amount)",
		}
		bounded := false
		if from, ok := m["from"].(float64); ok {
			name := fmt.Sprintf("grade_from_%d", i)
			vars[name] = from
			parts = append(parts, fmt.Sprintf("node.spending.amount >= @%s", name))
			bounded = true
		}
		if to, ok := m["to"].(float64); ok {
			name := fmt.Sprintf("grade_to_%d", i)
			vars[name] = to
			parts = append(parts, fmt.Sprintf("node.spending.amount < @%s", name))
			bounded = true
		}
		if !bounded {
			continue
		}
		conds = append(conds, "("+strings.Join(parts, " && ")+")")
	}

	if len(conds) == 0 {
		return "", nil
	}
	return "(" + strings.Join(conds, " || ") + ")", vars
}
