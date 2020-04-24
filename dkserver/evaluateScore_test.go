package main

import "testing"

func d(v int) die {
	return die{Value: v}
}

func TestEvaluateScore(t *testing.T) {

	type scoreTest struct {
		Name          string
		Dice          []die
		ExpectedScore int
	}

	tests := []scoreTest{
		{
			Name:          "straight",
			Dice:          []die{d(1), d(2), d(3), d(4), d(5), d(6)},
			ExpectedScore: 1500,
		},
		{
			Name:          "triples",
			Dice:          []die{d(1), d(1), d(1), d(2), d(2), d(2)},
			ExpectedScore: 2500,
		},
		{
			Name:          "pairs",
			Dice:          []die{d(1), d(1), d(5), d(5), d(2), d(2)},
			ExpectedScore: 1500,
		},
		{
			Name:          "tripleones",
			Dice:          []die{d(1), d(1), d(1), d(2), d(3), d(4)},
			ExpectedScore: 1000,
		},
		{
			Name:          "triplesixes",
			Dice:          []die{d(6), d(6), d(6), d(2), d(3), d(4)},
			ExpectedScore: 600,
		},
	}

	for _, test := range tests {
		if test.ExpectedScore != evaluateScore(test.Dice) {
			t.Fatalf("Test '%s' failed: %d != %d", test.Name, test.ExpectedScore, evaluateScore(test.Dice))
		}
	}

}
