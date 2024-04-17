package constants

const (

	// PlayerSpeed is the speed at which players move
	PlayerSpeed float64 = 350.0
	// PlayerJumpSpeed is the speed at which players jump
	PlayerJumpSpeed float64 = 750.0
	// Player Height
	PlayerHeight float64 = 32.0
	// Player Width
	PlayerWidth float64 = 32.0
	// Player Starting X
	PlayerStartingX float64 = 320.0
	// Player Starting Y
	PlayerStartingY float64 = 240.0
	// PlayerGravityMultiplier
	PlayerGravityMultiplier float64 = 300.0

	// Attack channel time is attack duration

	// PlayerAttackDuration is the duration of the attack (channel time + cooldown time)
	PlayerAttackDuration float64 = 0.4 // seconds
	// PlayerAttackChannelTime is the time it takes for the attack to register
	PlayerAttackChannelTime float64 = 0.1 // seconds
	// PlayerAttackHitboxWidth is the width of the attack hitbox
	PlayerAttackHitboxWidth float64 = PlayerWidth
	// PlayerAttackHitboxOffset is the offset from the player's position to the attack hitbox
	PlayerAttackHitboxOffset float64 = PlayerWidth / 2
	// PlayerAttackDamage is the amount of damage a player does
	PlayerAttackDamage int16 = 25

	// NPCSpeed is the speed at which NPCs move
	NPCSpeed float64 = 100.0
	// NPC Height
	NPCHeight float64 = 32.0
	// NPC Width
	NPCWidth float64 = 32.0
	// NPC Starting X
	NPCStartingX float64 = 100.0
	// NPC Starting Y
	NPCStartingY float64 = 16.0
	// NPCGravityMultiplier
	NPCGravityMultiplier float64 = 300.0
	// NPCRespawnTime is the time it takes for an NPC to respawn
	NPCRespawnTime float64 = 10.0
	// NPC Hitpoints
	NPCHitpoints int16 = 100
)
