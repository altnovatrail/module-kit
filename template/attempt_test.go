package template

import (
	"errors"
	"io/fs"
	"maps"
	"math"
	"path"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/altnovatrail/spec"
	"github.com/altnovatrail/spec/types/formatv0"

	"github.com/altnovatrail/module-kit/internal/random"
	"github.com/altnovatrail/module-kit/internal/yaml"
)

// An attempt of a vector of templates: its seed, the parameters of the instance that the vector
// gives it, and the number of that instance, counted from them here in the order of the odometer.
type attempt struct {
	seed       uint64
	parameters map[string]int64
	instance   int
}

// The vectors of templates without a solver, by their files, each with every attempt it holds
// (format/README.md, section 23). The order of the odometer: the parameters by their names, compared
// by code units, the first changing slowest, and the combinations that the constraint leaves out
// not numbered.
var vectorAttempts = map[string][]attempt{
	// One instance: below(1) gives 0 for every seed.
	"one-instance.yaml": {
		{0, map[string]int64{"a": 4}, 0},
		{math.MaxUint64, map[string]int64{"a": 4}, 0},
	},
	// a from 1 to 2 and b from 3 to 4: the instance is 2(a − 1) + b − 3.
	"escapes.yaml": {
		{0, map[string]int64{"a": 1, "b": 4}, 1},
		{2, map[string]int64{"a": 1, "b": 3}, 0},
		{8, map[string]int64{"a": 2, "b": 3}, 2},
	},
	// first and second from 2 to 9: 8(first − 2) + second − 2.
	"missing-factor.yaml": {
		{0, map[string]int64{"first": 2, "second": 3}, 1},
		{math.MaxUint64, map[string]int64{"first": 7, "second": 5}, 43},
		{20261027, map[string]int64{"first": 6, "second": 4}, 34},
	},
	// chickens and rabbits from 1 to 9 but never alike: 8(chickens − 1), and then rabbits − 1 when it
	// is below chickens and rabbits − 2 when it is above.
	"heads-and-legs.yaml": {
		{0, map[string]int64{"chickens": 1, "rabbits": 8}, 6},
		{math.MaxUint64, map[string]int64{"chickens": 1, "rabbits": 4}, 2},
		{7, map[string]int64{"chickens": 2, "rabbits": 7}, 13},
	},
	// a1 from −3 to −1, a_1 from 0 to 1, ab from 5 to 6 and b from 1 to 2, in that order of their
	// code units: 8(a1 + 3) + 4·a_1 + 2(ab − 5) + b − 1.
	"order-of-names.yaml": {
		{0, map[string]int64{"a1": -3, "a_1": 0, "ab": 5, "b": 1}, 0},
		{math.MaxUint64, map[string]int64{"a1": -2, "a_1": 0, "ab": 6, "b": 1}, 10},
		{1, map[string]int64{"a1": -1, "a_1": 1, "ab": 5, "b": 1}, 20},
		{2, map[string]int64{"a1": -1, "a_1": 0, "ab": 6, "b": 2}, 19},
		{3, map[string]int64{"a1": -2, "a_1": 1, "ab": 6, "b": 2}, 15},
	},
}

// The instance of an attempt is the one that the vectors of templates without a solver give each
// seed. The vectors are read from spec.Format, so that the attempts here are every attempt of every
// such vector, with the parameters they give; their task gives the ID and their count the count.
// The vectors whole, with the instance that each seed gets, are the templates' (T31).
func TestAttemptInstanceOfTheVectors(t *testing.T) {
	files, err := fs.Glob(spec.Format, "vectors/templates/*.yaml")
	if err != nil || len(files) == 0 {
		t.Fatalf("the vectors of templates: %v", err)
	}
	var held []string
	for _, name := range files {
		data, err := fs.ReadFile(spec.Format, name)
		if err != nil {
			t.Fatal(err)
		}
		value, err := yaml.Read(data)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		v := value.(map[string]any)
		task := v["task"].(map[string]any)
		if _, solver := task["template"].(map[string]any)["solver"]; solver {
			continue
		}
		file := path.Base(name)
		held = append(held, file)
		id := formatv0.TaskID(task["id"].(string))
		count := int(v["count"].(int64))
		given := v["attempts"].([]any)
		attempts := vectorAttempts[file]
		if len(given) != len(attempts) {
			t.Errorf("%s holds %d attempts, and the test %d", file, len(given), len(attempts))
			continue
		}
		for i, a := range attempts {
			g := given[i].(map[string]any)
			numbers := g["numbers"].(map[string]any)
			if seed, _ := strconv.ParseUint(g["seed"].(string), 10, 64); seed != a.seed {
				t.Fatalf("%s: attempt %d has the seed %d, and the test %d", file, i+1, seed, a.seed)
			}
			for name, n := range a.parameters {
				if numbers[name] != n {
					t.Fatalf("%s: the seed %d gives %s %v, and the test %d", file, a.seed, name, numbers[name], n)
				}
			}
			if got, err := AttemptInstance(a.seed, id, count); err != nil || got != a.instance {
				t.Errorf("%s: the seed %d gives the instance %d, %v, want %d, of %v", file, a.seed, got, err, a.instance, a.parameters)
			}
		}
	}
	if want := slices.Sorted(maps.Keys(vectorAttempts)); !slices.Equal(held, want) {
		t.Errorf("the vectors without a solver are %v, and the test holds %v", held, want)
	}
}

// The instance of an attempt is below(count) of the place whose names are template, the task's ID
// and instance, in that order, and lies from 0 to count − 1 for every count from 1 to 10 000.
func TestAttemptInstanceIsItsPlace(t *testing.T) {
	g, err := random.Place(2026, "attempts")
	if err != nil {
		t.Fatal(err)
	}
	for count := 1; count <= MaxInstances; count++ {
		seed := g.Integer()
		got, err := AttemptInstance(seed, "heads-and-legs-2", count)
		if err != nil || got < 0 || got >= count {
			t.Fatalf("the seed %d of %d instances gives %d, %v", seed, count, got, err)
		}
		place, _ := random.Place(seed, "template", "heads-and-legs-2", "instance")
		if want, _ := place.Below(uint64(count)); uint64(got) != want {
			t.Fatalf("the seed %d of %d instances gives %d, and its place %d", seed, count, got, want)
		}
	}
}

// A count outside 1 to 10 000 gives ErrDomain, and so does an ID that names no place.
func TestAttemptInstanceRefuses(t *testing.T) {
	for _, count := range []int{math.MinInt, -1, 0, MaxInstances + 1, math.MaxInt} {
		if got, err := AttemptInstance(1, "a", count); !errors.Is(err, ErrDomain) || got != 0 || !strings.Contains(err.Error(), "from 1 to 10000") {
			t.Errorf("%d instances give %d, %v", count, got, err)
		}
	}
	for _, task := range []formatv0.TaskID{"", "a|b", "\xff"} {
		if got, err := AttemptInstance(1, task, 2); !errors.Is(err, ErrDomain) || got != 0 {
			t.Errorf("the task %q gives %d, %v", task, got, err)
		}
	}
}
