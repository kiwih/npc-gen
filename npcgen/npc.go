package npcgen

import (
	"encoding/json"
	"fmt"
)

// min finds the minimum value of two ints. Why is this even necessary
func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

//NPC is the overall container for a non-player character generated by npcgen package
type NPC struct {
	HitPoints        DiceFunction `json:"-"`
	BaseStatBlock    StatBlock
	ProficiencyBonus int
	Race             RaceTraits

	Items []Item
}

//ACMod is for modifying a character's base AC
type ACMod struct {
	Set      int //Sets base AC of character to this value
	Addition int //Adds this to character's AC no questions asked

	AddMaxAbilityScores AbilityScores //Allows up to this much of each abilityScore to be added to the character's AC, -1 for infinite
}

// baseAC is the way we calculate AC for someone not wearing any clothes
var baseAC = ACMod{
	Set:      10,
	Addition: 0,
	AddMaxAbilityScores: AbilityScores{
		Str: 0,
		Dex: -1,
		Con: 0,
		Int: 0,
		Wis: 0,
		Cha: 0,
	},
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
	var acMod int
	acMethod := baseAC
	return n.calculateAC(acMethod) + acMod //TODO
}

// calculateAC finds the AC of a character given a specific method of calculating AC
func (n NPC) calculateAC(a ACMod) int {
	stats := n.StatBlock()
	var asMod int

	clampAbilityScore := func(a AbilityScore, b int, mod *int) {
		if a != 0 {
			if a < 0 {
				*mod += b
			} else {
				*mod += min(int(a), b)
			}
		}
	}

	clampAbilityScore(a.AddMaxAbilityScores.Str, stats.Str.Modifier(), &asMod)
	clampAbilityScore(a.AddMaxAbilityScores.Dex, stats.Dex.Modifier(), &asMod)
	clampAbilityScore(a.AddMaxAbilityScores.Con, stats.Con.Modifier(), &asMod)
	clampAbilityScore(a.AddMaxAbilityScores.Int, stats.Int.Modifier(), &asMod)
	clampAbilityScore(a.AddMaxAbilityScores.Wis, stats.Wis.Modifier(), &asMod)
	clampAbilityScore(a.AddMaxAbilityScores.Cha, stats.Cha.Modifier(), &asMod)

	return a.Set + a.Addition + asMod
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
		prof = n.ProficiencyBonus
	}
	return n.StatBlock().Str.Modifier() + prof
}

//DexAttackModifier returns the dexterity-based attack modifier for an npc.
func (n NPC) DexAttackModifier(addProf bool) int {
	prof := 0
	if addProf {
		prof = n.ProficiencyBonus
	}
	return n.StatBlock().Dex.Modifier() + prof
}

//GetAllActions returns all Actions that a NPC can do
func (n NPC) GetAllActions() []Action {
	return nil //TODO
}

//GetAllReactions returns all Actions that a NPC can do
func (n NPC) GetAllReactions() []Action {
	return nil //TODO
}

//GetAllFeatures returns all Features that a NPC can do
func (n NPC) GetAllFeatures() []Feature {
	return nil //TODO
}

func (n NPC) String() string {
	h := n.HP()
	s := fmt.Sprintf("Name: %s\nRace: %s\nHP: %s\nAC: %d\n\n\n", "Bandit Captain", "Human", h.String(), n.AC())
	j, _ := json.MarshalIndent(n, "", "\t")
	s += string(j)
	return s
}
