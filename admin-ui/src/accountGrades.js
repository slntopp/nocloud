// Account grades. Billing keeps account.spending (average real money paid per
// month over the last 13 months, in the default currency); the grade is derived from
// the "account-grades" setting, so changing the thresholds applies at once.
//
// Setting value:
// {
//   "grades": [
//     { "title": "A+", "color": "#8E24AA" },          // no "min": manual only
//     { "title": "A", "min": 900, "color": "#2E7D32" },
//     { "title": "B", "min": 300, "color": "#1565C0" },
//     { "title": "C", "min": 0, "color": "#757575" }
//   ]
// }

export const ACCOUNT_GRADES_KEY = "account-grades";

export const DEFAULT_ACCOUNT_GRADES = [
  { title: "A+", color: "#8E24AA" },
  { title: "A", min: 900, color: "#2E7D32" },
  { title: "B", min: 300, color: "#1565C0" },
  { title: "C", min: 0, color: "#757575" },
];

export function parseAccountGrades(raw) {
  try {
    const grades = JSON.parse(raw || "null")?.grades;
    if (Array.isArray(grades)) {
      const valid = grades.filter((g) => g && typeof g.title === "string");
      if (valid.length) return valid;
    }
  } catch {
    // fall through to the defaults
  }
  return DEFAULT_ACCOUNT_GRADES;
}

const hasMin = (g) => g.min !== undefined && g.min !== null && g.min !== "";

// Returns { title, color, manual, spending } or null when the account has no
// spending yet (billing has not counted it) and no manual grade.
export function getAccountGrade(account, grades = DEFAULT_ACCOUNT_GRADES) {
  if (!account) return null;
  const spending = account.spending;
  const override = (account.gradeOverride || "").trim();

  if (override) {
    const grade = grades.find((g) => g.title === override);
    return { title: override, color: grade?.color, manual: true, spending };
  }
  if (!spending) return null;

  const amount = +spending.amount || 0;
  const grade = grades
    .filter(hasMin)
    .sort((a, b) => +b.min - +a.min)
    .find((g) => amount >= +g.min);
  if (!grade) return null;

  return { title: grade.title, color: grade.color, manual: false, spending };
}

// Builds the accounts "grade" filter for the selected grade titles: a grade
// matches accounts with that manual grade, and accounts without a manual grade
// whose spending is in [min, next higher min).
export function gradeFilterConditions(titles, grades = DEFAULT_ACCOUNT_GRADES) {
  if (!Array.isArray(titles) || !titles.length) return undefined;

  const mins = [...new Set(grades.filter(hasMin).map((g) => +g.min))].sort(
    (a, b) => a - b,
  );

  const conditions = [];
  titles.forEach((title) => {
    conditions.push({ override: title });
    const grade = grades.find((g) => g.title === title);
    if (!grade || !hasMin(grade)) return;
    const from = +grade.min;
    const to = mins.find((m) => m > from);
    conditions.push(to === undefined ? { from } : { from, to });
  });
  return conditions;
}
