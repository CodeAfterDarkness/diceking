package main

import "fmt"

func evaluateScore(dice []die, output *[]string) int {
	var score int

	straight := []int{1, 2, 3, 4, 5, 6}
	for _, die := range dice {
		for j := len(straight) - 1; j >= 0; j-- {
			if straight[j] == die.Value {
				straight = append(straight[:j], straight[j+1:]...)
			}
		}
	}

	fmt.Printf("Straight: %v\n", straight)

	if len(straight) == 0 {
		score = 1500
		return score
	}

	if len(dice) == 6 {
		matches := []int{0, 0, 0, 0, 0, 0}
		for i, dieA := range dice {
			if matches[dieA.Value-1] > 0 {
				continue
			}

			// if i == 6 {
			// 	continue
			// }
			for j, dieB := range dice[i+1:] {
				j = j + i + 1

				if dieA.Value == dieB.Value {
					if matches[dieA.Value-1] == 0 {
						matches[dieA.Value-1] = 2
					} else {
						matches[dieA.Value-1]++
					}
				}
			}
		}
	}

	fmt.Printf("Matches: %v\n", matches)

	tripleCandidates := []int{}
	pairCandidates := []int{}

	for idx, count := range matches {
		if count < 2 {
			continue
		}

		switch count {
		case 6:
			score += 3000
			for i, _ := range dice {
				dice[i].Scored = true
			}
			return score
		case 5:
			score += 2000
			for i, _ := range dice {
				dice[i].Scored = true
			}
			return score
		case 4:
			score += 1000
			for i, _ := range dice {
				dice[i].Scored = true
			}
			return score
		case 3:
			tripleCandidates = append(tripleCandidates, idx+1)
		case 2:
			pairCandidates = append(pairCandidates, idx+1)
		}
	}

	if len(tripleCandidates) == 1 {
		var points int
		if tripleCandidates[0] == 1 {
			points = 1000
		} else {
			points = tripleCandidates[0] * 100
		}

		score += points
		for i, _ := range dice {
			if dice[i].Value == tripleCandidates[0] {
				dice[i].Scored = true
			}
		}
	}

	if len(tripleCandidates) == 2 {
		score += 2500
		return score
	}

	if len(pairCandidates) == 3 {
		score += 1500
		return score
	}

	for _, die := range dice {
		if die.Scored {
			continue
		}

		if die.Value == 1 {
			score += 100
		}

		if die.Value == 5 {
			score += 50
		}
	}

	*output = append(*output, fmt.Sprintf("Scored %d", score))

	return score
}
