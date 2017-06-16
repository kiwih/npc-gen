package npcgen

import (
	"encoding/json"
	"fmt"
)

//NPC is the overall container for a non-player character generated by npcgen package
type NPC struct {
	Name          string
	HitPoints     DiceFunction `json:"-"`
	BaseStatBlock StatBlock
	Race          RaceTraits

	ConstantProficiencyModifier int //This is used to make an NPC more or less difficult, and does not exist on PCs

	Items []Item
}

//ACMod is for modifying a character's base AC
type ACMod struct {
	Set      int //Sets base AC of character to this value
	Addition int //Adds this to character's AC no questions asked

	AddMaxAbilityScores AbilityScores //Allows up to this much of each abilityScore to be added to the character's AC, -1 or AbilityScoreUnlimited for infinite
}

//AC returns the final AC of the character
func (n NPC) AC() int {
	/*
		Things that change base AC (10 + dex):
			armor (light, med, heavy)
			spells (like mage armor or barkskin)
			class features (like unarmored defense or the dragon sorcerer one)
		Things that modify AC:
			shields
			spells (like shield of faith)
			class features (like protection fighting style)
			magic items (like ioun stone of protection)
	*/
	// AC algorithm:
	// Start at top of modifier stack (furthest from inherent properties) then go down
	// Keep track of modifiers along the way, but update AC if a more inherent method is available
	// e.g. spells -> items -> class features -> racial traits -> base AC
	/*
		baseAC = 10 + Dex.Modifier()
		for _, item in range items
			if item.isArmor
				if calculateAC(item) > baseAC
					baseAC = calculateAC(item)
			if item.modifiesAC
				acMod += item.acMod
		finalAC = baseAC + acMod
	*/
	// TODO: account for other acMethods
	var acMod int
	acMethod := BaseAC
	return n.calculateAC(acMethod) + acMod
}

// calculateAC finds the AC of a character given a specific method of calculating AC
func (n NPC) calculateAC(a ACMod) int {
	stats := n.StatBlock()
	var asMod int

	// It's weird that this is necessary
	min := func(x, y int) int {
		if x < y {
			return x
		}
		return y
	}

	clampAbilityScore := func(a AbilityScore, b int) int {
		if a != 0 {
			if a == AbilityScoreUnlimited {
				return b
			}
			return min(int(a), b)
		}
		return 0
	}

	asMod += clampAbilityScore(a.AddMaxAbilityScores.Str, stats.Str.Modifier())
	asMod += clampAbilityScore(a.AddMaxAbilityScores.Dex, stats.Dex.Modifier())
	asMod += clampAbilityScore(a.AddMaxAbilityScores.Con, stats.Con.Modifier())
	asMod += clampAbilityScore(a.AddMaxAbilityScores.Int, stats.Int.Modifier())
	asMod += clampAbilityScore(a.AddMaxAbilityScores.Wis, stats.Wis.Modifier())
	asMod += clampAbilityScore(a.AddMaxAbilityScores.Cha, stats.Cha.Modifier())

	return a.Set + a.Addition + asMod
}

//ProficiencyBonus returns the appropriate proficiency bonus for an NPC based on guesstimation of character level by assuming 1HD = 1 level
func (n NPC) ProficiencyBonus() int {
	return 2 + (len(n.HP().Dice)-1)/4 + n.ConstantProficiencyModifier
}

//HP returns the final max hitpoints of the character
func (n NPC) HP() DiceFunction {
	t := n.HitPoints
	t.Constant += len(n.HitPoints.Dice) * n.StatBlock().Con.Modifier() //TODO: make more efficient by caching n.StatBlock()?
	return t
}

//StatBlock returns the final statblock of the character (combining all other statblocks)
func (n NPC) StatBlock() StatBlock {
	s := n.BaseStatBlock
	s = CombineStatBlocks(s, n.Race.StatBlockMods)
	//TODO: iterate through items
	return s
}

//SpellSaveDC returns the spell save DC for a caster. If the NPC cannot cast, the minimum value is returned
func (n NPC) SpellSaveDC() int {
	return 0 //TODO
}

//SpellAttackModifier returns the spell attack modiifer for a caster. If the NPC cannot cast, the minimum value is returned.
func (n NPC) SpellAttackModifier(addProf bool) int {
	// prof := 0
	// if addProf {
	// 	prof = n.ProficiencyBonus
	// }
	return 0 //TODO
}

//StrAttackModifier returns the strength-based attack modifier for an npc.
func (n NPC) StrAttackModifier(addProf bool) int {
	prof := 0
	if addProf {
		prof = n.ProficiencyBonus()
	}
	return n.StatBlock().Str.Modifier() + prof
}

//DexAttackModifier returns the dexterity-based attack modifier for an npc.
func (n NPC) DexAttackModifier(addProf bool) int {
	prof := 0
	if addProf {
		prof = n.ProficiencyBonus()
	}
	return n.StatBlock().Dex.Modifier() + prof
}

//GetAllFeatures returns all features that currently are applying/available to an NPC
func (n NPC) GetAllFeatures() []Feature {
	var features []Feature
	features = append(features, n.Race.RacialFeatures...)
	for _, item := range n.Items {
		features = append(features, item.Features...)
	}
	return features
}

//GetAllActions returns all Actions that a NPC can do, based on their available features
func (n NPC) GetAllActions() []Action {
	var baseActions []Action
	features := n.GetAllFeatures()
	for _, feature := range features {
		baseActions = append(baseActions, feature.Actions...)
	}
	return baseActions
}

//GetAllReactions returns all Reactions that a NPC can do, based on their available features
func (n NPC) GetAllReactions() []Action {
	var baseReactions []Action
	features := n.GetAllFeatures()
	for _, feature := range features {
		baseReactions = append(baseReactions, feature.Reactions...)
	}
	return baseReactions
}

func (n NPC) String() string {
	h := n.HP()

	j, _ := json.MarshalIndent(n, "", "\t")
	s := string(j)

	m := fmt.Sprintf("\n\nName: %s\nRace: %s\nHP: %s\nAC: %d\n", n.Name, "Human", h.String(), n.AC())
	s += m

	s += "\nActions:\n"

	features := n.GetAllFeatures()
	for _, feature := range features {
		for _, action := range feature.Actions {
			m := fmt.Sprintf(
				"\n%s (%s) (%s) %s, %s, Hit: %s %s damage",
				feature.Name,
				action.Name,
				action.ActionType,
				action.AttackString(n),
				action.RangeString(),
				action.DamageString(n),
				action.DamageTypeString(),
			)
			s += m
		}
	}
	return s
}
