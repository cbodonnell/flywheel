package network

import "sort"

// removeOutlierRTTs removes outler RTTs from the recent RTTs.
// An outlier RTT is defined as an RTT that is greater than 2 times the median RTT
// and is also greater than 20ms.
func removeOutlierRTTs(recentRTTs []int64) []int64 {
	result := make([]int64, 0)
	medianRTT := medianRTT(recentRTTs)
	for i := 0; i < len(recentRTTs); i++ {
		if recentRTTs[i] > 2*medianRTT && recentRTTs[i] > 20 {
			continue
		}
		result = append(result, recentRTTs[i])
	}
	return result
}

// medianRTT returns the median RTT from a slice of RTTs.
func medianRTT(recentRTTs []int64) int64 {
	if len(recentRTTs) == 0 {
		return 0
	}
	sortedRTTs := make([]int64, len(recentRTTs))
	copy(sortedRTTs, recentRTTs)
	sort.Slice(sortedRTTs, func(i, j int) bool {
		return sortedRTTs[i] < sortedRTTs[j]
	})
	if len(sortedRTTs)%2 == 0 {
		return (sortedRTTs[len(sortedRTTs)/2-1] + sortedRTTs[len(sortedRTTs)/2]) / 2
	}
	return sortedRTTs[len(sortedRTTs)/2]
}
