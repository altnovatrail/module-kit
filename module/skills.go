package module

import (
	"fmt"
	"maps"
	"slices"
	"strconv"
	"strings"

	"github.com/altnovatrail/spec/types/formatv0"
)

// isAnchor says whether a prerequisite names an anchor skill of the shared taxonomy, core.<name>,
// rather than a skill of the module by its local ID, which has no dot (format/README.md, section
// 13).
func isAnchor(id string) bool {
	return strings.HasPrefix(id, "core.")
}

// skillGraph checks the rules within a skill graph (format/README.md, section 12):
// skill-graph.known-skills, skill-graph.retired-ids, skill-graph.own-anchors and
// skill-graph.acyclic. teachers are the skills of the graph that teach each anchor skill.
func (c *checker) skillGraph(path string, g formatv0.SkillGraph, teachers map[formatv0.SharedTaxonomy][]string) {
	for id, skill := range g.Skills {
		for i, prerequisite := range skill.Prerequisites {
			place := at("skills", string(id), "prerequisites", strconv.Itoa(i))
			if !isAnchor(prerequisite) && !hasKey(g.Skills, formatv0.SkillID(prerequisite)) {
				c.add(RuleSkillGraphKnownSkills, path, 0, 0, place,
					fmt.Sprintf("the prerequisite %s is no skill of the module", prerequisite))
			}
			if taught := teachers[formatv0.SharedTaxonomy(prerequisite)]; isAnchor(prerequisite) && len(taught) > 0 {
				c.add(RuleSkillGraphOwnAnchors, path, 0, 0, place,
					fmt.Sprintf("the prerequisite %s is an anchor skill that %s of the module teaches: name the skill", prerequisite, strings.Join(taught, ", ")))
			}
		}
	}
	for id, retired := range g.RetiredSkills {
		if hasKey(g.Skills, id) {
			c.add(RuleSkillGraphRetiredIDs, path, 0, 0, at("retired_skills", string(id)),
				fmt.Sprintf("%s is retired and the ID of a skill too: a published ID never names another skill", id))
		}
		for i, skill := range retired.ReplacedBy {
			if !hasKey(g.Skills, skill) {
				c.add(RuleSkillGraphKnownSkills, path, 0, 0, at("retired_skills", string(id), "replaced_by", strconv.Itoa(i)),
					fmt.Sprintf("the replacement %s is no skill of the module", skill))
			}
		}
	}
	edges := edgesOf(g)
	for id, others := range cycles(edges) {
		var says []string
		if slices.Contains(edges[id], id) {
			says = append(says, "the skill names itself among its prerequisites")
		}
		if len(others) > 0 {
			says = append(says, "the skill is, through its prerequisites, a prerequisite of itself, in a cycle with "+strings.Join(others, ", "))
		}
		c.add(RuleSkillGraphAcyclic, path, 0, 0, at("skills", string(id), "prerequisites"), strings.Join(says, "; "))
	}
}

// unknownSkill says whether a part names, by its local ID, a skill that the module's skill graph
// does not have, and what the breach says: subject is how the part names it, such as "the trap
// names".
func unknownSkill(g *formatv0.SkillGraph, subject string, id formatv0.SkillID) (string, bool) {
	switch {
	case g == nil:
		return fmt.Sprintf("%s the skill %s, and the module has no skill graph", subject, id), true
	case !hasKey(g.Skills, id):
		return fmt.Sprintf("%s the skill %s, which the skill graph does not have", subject, id), true
	}
	return "", false
}

// teachersOf gives the skills of a graph that teach each anchor skill, in order.
func teachersOf(g formatv0.SkillGraph) map[formatv0.SharedTaxonomy][]string {
	teachers := map[formatv0.SharedTaxonomy][]string{}
	for _, id := range slices.Sorted(maps.Keys(g.Skills)) {
		for _, anchor := range g.Skills[id].Teaches {
			teachers[anchor] = append(teachers[anchor], string(id))
		}
	}
	return teachers
}

// edgesOf gives the prerequisites of each skill of a graph that are skills of the module: neither
// anchor skills nor IDs that the graph does not have.
func edgesOf(g formatv0.SkillGraph) map[formatv0.SkillID][]formatv0.SkillID {
	edges := map[formatv0.SkillID][]formatv0.SkillID{}
	for id, skill := range g.Skills {
		for _, prerequisite := range skill.Prerequisites {
			if to := formatv0.SkillID(prerequisite); !isAnchor(prerequisite) && hasKey(g.Skills, to) {
				edges[id] = append(edges[id], to)
			}
		}
	}
	return edges
}

// cycles gives each node of a graph, by its edges, that lies on a cycle — a skill of the module on a
// cycle of prerequisites, a lesson on a cycle of the order of lessons — with the other nodes of its
// cycle, in order: the members of each strongly connected component of more than one node, and a
// node with an edge to itself, with no others. It finds the components as Tarjan does (SIAM Journal
// on Computing 1(2), 1972), in time linear in the graph.
func cycles[K ~string](edges map[K][]K) map[K][]string {
	found := map[K][]string{}
	index, low := map[K]int{}, map[K]int{}
	onStack := map[K]bool{}
	var stack []K
	var connect func(v K)
	connect = func(v K) {
		index[v], low[v] = len(index), len(index)
		stack = append(stack, v)
		onStack[v] = true
		for _, w := range edges[v] {
			if _, seen := index[w]; !seen {
				connect(w)
				low[v] = min(low[v], low[w])
			} else if onStack[w] {
				low[v] = min(low[v], index[w])
			}
		}
		if low[v] != index[v] {
			return
		}
		var component []string
		for {
			w := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			onStack[w] = false
			component = append(component, string(w))
			if w == v {
				break
			}
		}
		if len(component) == 1 && !slices.Contains(edges[v], v) {
			return
		}
		slices.Sort(component)
		for _, member := range component {
			found[K(member)] = slices.DeleteFunc(slices.Clone(component), func(s string) bool { return s == member })
		}
	}
	for _, id := range slices.Sorted(maps.Keys(edges)) {
		if _, seen := index[id]; !seen {
			connect(id)
		}
	}
	return found
}

// Tiers gives the tier of each skill of a skill graph (format/README.md, section 12): tier 1 for a
// skill with no prerequisite among the module's skills, and one tier above the highest of its
// prerequisites among them for every other, so that an anchor skill counts for none. The package
// holds them in its canonical view, as skill_tiers (section 25). A graph whose prerequisites name a
// skill that it does not have, or run in a cycle, has no tiers: it breaks skill-graph.known-skills
// or skill-graph.acyclic, and Tiers gives an error.
func Tiers(g formatv0.SkillGraph) (formatv0.SkillTiers, error) {
	for _, id := range slices.Sorted(maps.Keys(g.Skills)) {
		for _, prerequisite := range g.Skills[id].Prerequisites {
			if !isAnchor(prerequisite) && !hasKey(g.Skills, formatv0.SkillID(prerequisite)) {
				return nil, fmt.Errorf("the skill %s needs %s, which is no skill of the module", id, prerequisite)
			}
		}
	}
	edges := edgesOf(g)
	if found := cycles(edges); len(found) > 0 {
		id := slices.Sorted(maps.Keys(found))[0]
		return nil, fmt.Errorf("the prerequisites of %s run in a cycle: its tier is no number", id)
	}
	tiers := formatv0.SkillTiers{}
	var tierOf func(id formatv0.SkillID) int64
	tierOf = func(id formatv0.SkillID) int64 {
		if tier, ok := tiers[id]; ok {
			return tier
		}
		var highest int64
		for _, prerequisite := range edges[id] {
			highest = max(highest, tierOf(prerequisite))
		}
		tiers[id] = highest + 1
		return tiers[id]
	}
	for id := range g.Skills {
		tierOf(id)
	}
	return tiers, nil
}
