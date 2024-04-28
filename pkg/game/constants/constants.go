package constants

const (

	// PlayerSpeed is the speed at which players move
	PlayerSpeed float64 = 350.0
	// PlayerJumpSpeed is the speed at which players jump
	PlayerJumpSpeed float64 = 750.0
	// Player Height
	PlayerHeight float64 = 64.0
	// Player Width
	PlayerWidth float64 = 64.0
	// Player Starting X
	PlayerStartingX float64 = 320.0
	// Player Starting Y
	PlayerStartingY float64 = 240.0
	// PlayerGravityMultiplier
	PlayerGravityMultiplier float64 = 300.0

	// PlayerAttack1Duration is the duration of the attack (channel time + cooldown time)
	PlayerAttack1Duration float64 = 0.6 // seconds
	// PlayerAttack1ChannelTime is the time it takes for the attack to register
	PlayerAttack1ChannelTime float64 = 0.2 // seconds
	// PlayerAttack1HitboxWidth is the width of the attack hitbox
	PlayerAttack1HitboxWidth float64 = PlayerWidth
	// PlayerAttack1HitboxOffset is the offset from the player's position to the attack hitbox
	PlayerAttack1HitboxOffset float64 = PlayerWidth / 2
	// PlayerAttack1Damage is the amount of damage a player does
	PlayerAttack1Damage int16 = 30

	// PlayerAttack2Duration is the duration of the attack (channel time + cooldown time)
	PlayerAttack2Duration float64 = 0.3 // seconds
	// PlayerAttack2ChannelTime is the time it takes for the attack to register
	PlayerAttack2ChannelTime float64 = 0.0 // seconds
	// PlayerAttack2HitboxWidth is the width of the attack hitbox
	PlayerAttack2HitboxWidth float64 = PlayerWidth
	// PlayerAttack2HitboxOffset is the offset from the player's position to the attack hitbox
	PlayerAttack2HitboxOffset float64 = PlayerWidth / 2
	// PlayerAttack2Damage is the amount of damage a player does
	PlayerAttack2Damage int16 = 15

	// PlayerAttack3Duration is the duration of the attack (channel time + cooldown time)
	PlayerAttack3Duration float64 = 0.4 // seconds
	// PlayerAttack3ChannelTime is the time it takes for the attack to register
	PlayerAttack3ChannelTime float64 = 0.1 // seconds
	// PlayerAttack3HitboxWidth is the width of the attack hitbox
	PlayerAttack3HitboxWidth float64 = PlayerWidth
	// PlayerAttack3HitboxOffset is the offset from the player's position to the attack hitbox
	PlayerAttack3HitboxOffset float64 = PlayerWidth / 2
	// PlayerAttack3Damage is the amount of damage a player does
	PlayerAttack3Damage int16 = 20

	// NPCSpeed is the speed at which NPCs move
	NPCSpeed float64 = 100.0
	// NPC Height
	NPCHeight float64 = 64.0
	// NPC Width
	NPCWidth float64 = 64.0
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
