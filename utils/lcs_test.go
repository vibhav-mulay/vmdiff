package utils

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDetermineLCS(t *testing.T) {
	testcases := []struct {
		s1  string
		s2  string
		lcs string
	}{
		{
			s1:  "BCDAACD",
			s2:  "ACDBAC",
			lcs: "CDAC",
		},
		{
			s1:  "ABCDGH",
			s2:  "AEDFHR",
			lcs: "ADH",
		},
		{
			s1:  "AGGTAB",
			s2:  "GXTXAYB",
			lcs: "GTAB",
		},
		{
			s1:  "ABCDEF",
			s2:  "GHIJ",
			lcs: "",
		},
		{
			s1:  "ABCDEF",
			s2:  "ABCDEF",
			lcs: "ABCDEF",
		},
	}

	for i, tc := range testcases {
		t.Run(fmt.Sprintf("TestCase %d", i+1), func(t *testing.T) {
			l1 := strings.Split(tc.s1, "")
			l2 := strings.Split(tc.s2, "")

			lcsStr := DetermineLCS(l1, l2)
			lcs := strings.Join(lcsStr, "")

			assert.Equal(t, lcs, tc.lcs)
		})
	}
}

func TestMax(t *testing.T) {
	testcases := []struct {
		n1  int
		n2  int
		max int
	}{
		{
			n1:  10,
			n2:  20,
			max: 20,
		},
		{
			n1:  22,
			n2:  20,
			max: 22,
		},
		{
			n1:  10,
			n2:  10,
			max: 10,
		},
	}
	for i, tc := range testcases {
		t.Run(fmt.Sprintf("TestCase %d", i+1), func(t *testing.T) {
			max := Max(tc.n1, tc.n2)
			assert.Equal(t, max, tc.max)
		})
	}
}
