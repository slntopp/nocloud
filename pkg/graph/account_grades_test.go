package graph

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAccountGradeFilter(t *testing.T) {
	cond, vars := accountGradeFilter([]any{
		map[string]any{"override": "A"},
		map[string]any{"from": 300.0, "to": 900.0},
	})
	assert.Equal(t, `(node.grade_override == @grade_override_0 || ((IS_NULL(node.grade_override) || node.grade_override == "") && IS_NUMBER(node.spending.amount) && node.spending.amount >= @grade_from_1 && node.spending.amount < @grade_to_1))`, cond)
	assert.Equal(t, map[string]any{"grade_override_0": "A", "grade_from_1": 300.0, "grade_to_1": 900.0}, vars)
}

func TestAccountGradeFilterEmpty(t *testing.T) {
	cond, vars := accountGradeFilter([]any{
		map[string]any{"override": "  "},
		map[string]any{},
		"A",
	})
	assert.Equal(t, "", cond)
	assert.Nil(t, vars)
}
