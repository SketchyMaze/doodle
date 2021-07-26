package doodads

import "sort"

// SortByName orders an array of loaded Doodads by their titles.
func SortByName(list []*Doodad) {
	sort.SliceStable(list, func(i, j int) bool {
		return list[i].Title < list[j].Title
	})
}
