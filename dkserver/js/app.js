$(document).ready(function(){
	init();
});


function die(){
	this.value = 1;
	this.saved = false;
	this.committed = false;
	this.scored = false;
}

var dice = [];
var saved = 0;
var score = 0;
var tempScore = 0;
var scored = false;

function init() {
	console.log("Dice King initialized");

	for(x = 0; x < 6; x++){
		dice.push(new die())
	}

	//roll(dice);
	rollAjax(dice);
}

function log(msg){
	$('#feedback').append(msg + "<br/>");
}

function getRandomInt(max) {
  return Math.floor(Math.random() * Math.floor(max));
}

function userRoll() {
	console.log("rolling dice")
	$('#feedback').html('');

	if(scored) {
		$.each(dice, function(i,die){
			dice[i].committed = false;
			dice[i].saved = false;
			dice[i].scored = false;
		});
		roll(dice);
		scored = false;
		tempScore = 0;
		return;
	}

	savedCount = 0;
	$.each(dice, function(i,die){
		if(die.saved && !die.committed){
			savedCount++;
		}
	});

	if(!scored && savedCount === 0) {
		log("User must save at least 1 die to roll again");
		return
	}

	savedDice = [];
	$.each(dice, function(i,die){
		if(die.saved && !die.committed){
			savedDice.push(die);
		}
	});

	tempScore += evaluateScore(savedDice);

	roll(dice);
}

function userCommitScore(){
	scored = true;
	score += tempScore;
	tempScore = 0;

	savedDice = [];
	$.each(dice, function(i,die){
		if(!die.committed){
			savedDice.push(die);
		}
	});
	score += evaluateScore(savedDice);
	console.log("score increased to " + score);
	$('#score').html(score);
	//userRoll();
}

// Roll
// Save scoring die
// Click score to commit currently accumulated score
// OR
// Click roll to continue building score with another roll, potentially getting a farkle

function evaluateScore(dice){ // return score, but do not change game state

	var cscore = 0;

	var straight = [1,2,3,4,5,6]
	$.each(dice, function(i,die){
		$.each(straight, function(j, value){
			if(value == die.value){
				straight.splice(j, 1);
			}
		});
	});

	if(straight.length == 0){
		console.log("Got a straight!");
		cscore += 1500;
		$("#score").html(score);
		return cscore;
	}

	// evaluate for potential 2x triples, 3x pairs
	var matches = [0,0,0,0,0,0];
	$.each(dice, function(i,dieA){
		if(matches[dieA.value-1] > 0){
			return;
		}

		$.each(dice.slice(i+1), function(j,dieB){
			j = j+i+1

			if(dieA.value === dieB.value){
				//console.log("Found match ["+i+","+j+"]: dieA " + dieA.value + " === dieB " + dieB.value)
				if(matches[dieA.value-1] === 0){
					matches[dieA.value-1] = 2;
				} else {
					matches[dieA.value-1]++
				}
			}
		});
	});

	tripleCandidates = [];
	pairCandidates = [];

	$.each(matches, function(idx,count){
		if(count < 2){
			return
		}

		switch(count){
			case 6:
				console.log("Matched 6 " + (idx+1) + "'s! BOOM! 3000 points!")
				cscore += 3000;
				$.each(dice, function(i,die){
					if(die.value === idx+1){
						dice[i].scored = true;
					}
				});
				return;
			break;
			case 5:
				console.log("Matched 5 " + (idx+1) + "'s! BOOM! 2000 points!")
				cscore += 2000;
				$.each(dice, function(i,die){
					if(die.value === idx+1){
						dice[i].scored = true;
					}
				});
			break;
			case 4:
				console.log("Matched 4 " + (idx+1) + "'s! BOOM! 1000 points!")
				cscore += 1000;
				$.each(dice, function(i,die){
					if(die.value === idx+1){
						dice[i].scored = true;
					}
				});
			break;
			case 3:
				tripleCandidates.push(idx+1)
			break;
			case 2:
				pairCandidates.push(idx+1)
			break;
		}
	});

	//console.log("Found " + tripleCandidates.length + " triple candidates")
	//console.log("Found " + pairCandidates.length + " pair candidates")

	if(tripleCandidates.length === 1){
		var points = 0;
		if(tripleCandidates[0] === 1){
			points = 1000;
		} else {
			points = tripleCandidates[0] * 100
		}

		cscore += points;
		
		console.log("Matched 3 " + tripleCandidates[0] + "'s! " + points + " points!")
		$.each(dice, function(i,die){
			if (die.value === tripleCandidates[0]) {
				dice[i].scored = true;
			}
		});
	}

	if(tripleCandidates.length === 2){
		console.log("Triples! 2500 points!")
		cscore += 2500;
		return cscore;
	}

	if(pairCandidates.length === 3){
		console.log("3 pairs! 1500 points!")
		cscore += 1500;
		return cscore;
	}

	$.each(dice, function(i,die){
		if(die.scored) {
			return;
		}

		//console.log("Checking idx "+i+" for score on value "+die.value)

		if(die.value === 1) {
			console.log("Scored 100 at idx "+i)
			cscore += 100;
		}

		if(die.value === 5) {
			console.log("Scored 50 at idx "+i)
			cscore += 50;
		}
	});

	return cscore;
}

function restartAjax() {
	$.ajax({
		url: "/gracefullyRestart",
		method: "GET",
		dataType: "json",
		success: function(json){
			console.log("server restarted");
			window.location.reload(false); 
		},
	});
}

function rollAjax(dice){
	$.each(dice, function(i,die){
		if(die.saved) {
			dice[i].committed = true;
			return;
		}
	});

	$.ajax({
		url: "/roll",
		data: JSON.stringify(dice),
		method: "POST",
		dataType: "json",
		success: function(json){
			$.each(json.Player.Dice, function(i,die){
				dice[i].value = die.Value;
			});
			displayRoll(dice);
			$.each(json.Messages, function(i,msg){
				console.log(msg);
			})
			
		},
	});
}

function scoreAjax(dice){
	$.each(dice, function(i,die){
		if(die.saved) {
			dice[i].committed = true;
			return;
		}
	});
	
	$.ajax({
		url: "/score",
		data: JSON.stringify(dice),
		method: "POST",
		dataType: "json",
		success: function(json){
			$.each(json.Player.Dice, function(i,die){
				dice[i].value = die.Value;
				dice[i].saved = die.saved;
				dice[i].committed = die.committed;
				dice[i].scored = die.scored;
			});
			displayRoll(dice);
			$('#score').html(json.Player.Score);
			$.each(json.Messages, function(i,msg){
				console.log(msg);
			})
		},
	});
}

function displayRoll(dice){
	
	diceHTML = "";
	$.each(dice, function(i,die){
		//console.log("Die " + i + " Value: " + die.value);

		if(die.committed) {
			diceHTML += '<div class="dice dice-' + die.value + '-committed" id="die-idx-' + i + '" data-die-idx=' + i + '></div>';
		} else {
			diceHTML += '<div class="dice dice-' + die.value + '" id="die-idx-' + i + '" data-die-idx=' + i + '></div>';
		}
		
	});

	$('#dicetray').html(diceHTML);

	//console.log("pre-roll: "+s1+" and post-roll: " +s2)
	var potentialScore = evaluateScore(dice)
	console.log("roll has " + potentialScore + " potential points, bringing hand up to " + (tempScore + potentialScore))
	if(potentialScore === 0) {
		console.log("farkle");
		log("Farkle! All points this hand lost. Reroll.")
		scored = true;
		return
	} else {
		$('.dice').on('click', function(el){

			id = $(el.target).data("die-idx");

			if(dice[id].committed) {
				// not allowed to change saved die after rolling
				log("not allowed to change saved die after rolling")
				return
			}

			if(dice[id].saved) {
				//console.log("unsaving die " + id);
				dice[id].saved = false;
				$(el.target).addClass("dice-" + dice[id].value);
				$(el.target).removeClass("dice-" + dice[id].value + "-saved");
			} else {
				//console.log("saving die " + id);
				dice[id].saved = true;
				$(el.target).addClass("dice-" + dice[id].value + "-saved");
				$(el.target).removeClass("dice-" + dice[id].value);
			}
			
		});
	}

	
}

function roll(dice){
	// generate 6x random numbers 1-6

	// determine score before rolling
	// determine score after rolling
	// no change in score == farkle

	var newDice = [];
	$.each(dice, function(i,die){
		if(die.saved) {
			dice[i].committed = true;
			return;
		}
		dice[i].value = getRandomInt(6) + 1;
		newDice.push(dice[i])
	});

	diceHTML = "";
	$.each(dice, function(i,die){
		//console.log("Die " + i + " Value: " + die.value);

		if(die.committed) {
			diceHTML += '<div class="dice dice-' + die.value + '-committed" id="die-idx-' + i + '" data-die-idx=' + i + '></div>';
		} else {
			diceHTML += '<div class="dice dice-' + die.value + '" id="die-idx-' + i + '" data-die-idx=' + i + '></div>';
		}
		
	});

	$('#dicetray').html(diceHTML);

	//console.log("pre-roll: "+s1+" and post-roll: " +s2)
	var potentialScore = evaluateScore(newDice)
	console.log("roll has " + potentialScore + " potential points, bringing hand up to " + (tempScore + potentialScore))
	if(potentialScore === 0) {
		console.log("farkle");
		log("Farkle! All points this hand lost. Reroll.")
		scored = true;
		return
	} else {
		$('.dice').on('click', function(el){

			id = $(el.target).data("die-idx");

			if(dice[id].committed) {
				// not allowed to change saved die after rolling
				log("not allowed to change saved die after rolling")
				return
			}

			if(dice[id].saved) {
				//console.log("unsaving die " + id);
				dice[id].saved = false;
				$(el.target).addClass("dice-" + dice[id].value);
				$(el.target).removeClass("dice-" + dice[id].value + "-saved");
			} else {
				//console.log("saving die " + id);
				dice[id].saved = true;
				$(el.target).addClass("dice-" + dice[id].value + "-saved");
				$(el.target).removeClass("dice-" + dice[id].value);
			}
			
		});
	}

	
}