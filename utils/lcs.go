package utils

func DetermineLCS(lst1, lst2 []string) []string {
	len1 := len(lst1)
	len2 := len(lst2)

	lcsTable := make([][]int, len1+1)
	for i := range lcsTable {
		lcsTable[i] = make([]int, len2+1)
	}

	for i := 0; i <= len1; i++ {
		for j := 0; j <= len2; j++ {
			if i == 0 || j == 0 {
				lcsTable[i][j] = 0
			} else if lst1[i-1] == lst2[j-1] {
				lcsTable[i][j] = lcsTable[i-1][j-1] + 1
			} else {
				lcsTable[i][j] = Max(lcsTable[i-1][j], lcsTable[i][j-1])
			}
		}
	}

	lcs := make([]string, 0, lcsTable[len1][len2])
	for i, j := len1, len2; i > 0 && j > 0; {
		if lst1[i-1] == lst2[j-1] {
			lcs = append([]string{lst1[i-1]}, lcs...)
			i--
			j--
		} else if lcsTable[i-1][j] > lcsTable[i][j-1] {
			i--
		} else {
			j--
		}
	}

	return lcs
}

func Max(i, j int) int {
	if i > j {
		return i
	}
	return j
}
