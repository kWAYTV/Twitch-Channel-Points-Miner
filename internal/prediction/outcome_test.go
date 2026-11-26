package prediction

import (
	"reflect"
	"testing"
)

func TestParseOutcomes(t *testing.T) {
	t.Run("two outcomes with every field", func(t *testing.T) {
		raw := []interface{}{
			map[string]interface{}{
				"id": "a", "title": "Yes", "color": "BLUE",
				"total_users": float64(10), "total_points": float64(100),
				"top_predictors": []interface{}{
					map[string]interface{}{"points": float64(40)},
					map[string]interface{}{"points": float64(15)},
				},
			},
			map[string]interface{}{
				"id": "b", "title": "No", "color": "PINK",
				"total_users": float64(30), "total_points": float64(400),
				"top_predictors": []interface{}{
					map[string]interface{}{"points": float64(90)},
				},
			},
		}
		want := []Outcome{
			{ID: "a", Title: "Yes", Color: "BLUE", TotalUsers: 10, TotalPoints: 100, TopPoints: 40, PercentageUsers: 25, Odds: 5, OddsPercentage: 20},
			{ID: "b", Title: "No", Color: "PINK", TotalUsers: 30, TotalPoints: 400, TopPoints: 90, PercentageUsers: 75, Odds: 1.25, OddsPercentage: 80},
		}
		if got := parseOutcomes(raw); !reflect.DeepEqual(got, want) {
			t.Errorf("parseOutcomes() = %+v, want %+v", got, want)
		}
	})

	t.Run("no outcomes", func(t *testing.T) {
		want := []Outcome{}
		if got := parseOutcomes(nil); !reflect.DeepEqual(got, want) {
			t.Errorf("parseOutcomes(nil) = %#v, want %#v", got, want)
		}
	})

	t.Run("an outcome with only an id", func(t *testing.T) {
		raw := []interface{}{map[string]interface{}{"id": "a"}}
		want := []Outcome{{ID: "a"}}
		if got := parseOutcomes(raw); !reflect.DeepEqual(got, want) {
			t.Errorf("parseOutcomes() = %+v, want %+v", got, want)
		}
	})

	t.Run("an outcome nobody has bet on yet", func(t *testing.T) {
		raw := []interface{}{
			map[string]interface{}{"id": "a", "total_users": float64(4), "total_points": float64(200)},
			map[string]interface{}{"id": "b", "total_users": float64(0), "total_points": float64(0)},
		}
		want := []Outcome{
			{ID: "a", TotalUsers: 4, TotalPoints: 200, PercentageUsers: 100, Odds: 1, OddsPercentage: 100},
			{ID: "b"},
		}
		if got := parseOutcomes(raw); !reflect.DeepEqual(got, want) {
			t.Errorf("parseOutcomes() = %+v, want %+v", got, want)
		}
	})

	t.Run("numbers that arrive as integers", func(t *testing.T) {
		raw := []interface{}{
			map[string]interface{}{
				"id": "a", "total_users": 8, "total_points": 50,
				"top_predictors": []interface{}{map[string]interface{}{"points": 20}},
			},
		}
		want := []Outcome{{ID: "a", TotalUsers: 8, TotalPoints: 50, TopPoints: 20, PercentageUsers: 100, Odds: 1, OddsPercentage: 100}}
		if got := parseOutcomes(raw); !reflect.DeepEqual(got, want) {
			t.Errorf("parseOutcomes() = %+v, want %+v", got, want)
		}
	})

	t.Run("an entry that is not an object", func(t *testing.T) {
		raw := []interface{}{"a", map[string]interface{}{"id": "b"}}
		want := []Outcome{{ID: "b"}}
		if got := parseOutcomes(raw); !reflect.DeepEqual(got, want) {
			t.Errorf("parseOutcomes() = %+v, want %+v", got, want)
		}
	})

	t.Run("an empty list of top predictors", func(t *testing.T) {
		raw := []interface{}{map[string]interface{}{"id": "a", "top_predictors": []interface{}{}}}
		want := []Outcome{{ID: "a"}}
		if got := parseOutcomes(raw); !reflect.DeepEqual(got, want) {
			t.Errorf("parseOutcomes() = %+v, want %+v", got, want)
		}
	})

	t.Run("a top predictor that is not an object", func(t *testing.T) {
		raw := []interface{}{map[string]interface{}{"id": "a", "top_predictors": []interface{}{"someone"}}}
		want := []Outcome{{ID: "a"}}
		if got := parseOutcomes(raw); !reflect.DeepEqual(got, want) {
			t.Errorf("parseOutcomes() = %+v, want %+v", got, want)
		}
	})

	t.Run("a title outside ASCII", func(t *testing.T) {
		raw := []interface{}{map[string]interface{}{"id": "a", "title": "はい ✔", "color": "BLUE"}}
		want := []Outcome{{ID: "a", Title: "はい ✔", Color: "BLUE"}}
		if got := parseOutcomes(raw); !reflect.DeepEqual(got, want) {
			t.Errorf("parseOutcomes() = %+v, want %+v", got, want)
		}
	})
}

func TestOutcome_String(t *testing.T) {
	t.Run("an outcome with every figure", func(t *testing.T) {
		outcome := Outcome{Title: " Yes ", Color: "blue", TotalPoints: 1_500, TotalUsers: 25, PercentageUsers: 62.5, Odds: 1.25, OddsPercentage: 80}
		want := "Yes (BLUE), Points: 1.5k, Users: 25 (62.50%), Odds: 1.25 (80%)"
		if got := outcome.String(); got != want {
			t.Errorf("String() = %q, want %q", got, want)
		}
	})

	t.Run("an empty outcome", func(t *testing.T) {
		outcome := Outcome{}
		want := " (), Points: 0, Users: 0 (0.00%), Odds: 0 (0%)"
		if got := outcome.String(); got != want {
			t.Errorf("String() = %q, want %q", got, want)
		}
	})
}
