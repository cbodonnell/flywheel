package constants

const (
	// PlayerSpeed is the speed at which players move
	PlayerSpeed float64 = 350.0
	// PlayerJumpSpeed is the speed at which players jump
	PlayerJumpSpeed float64 = 850.0
	// Player Height
	PlayerHeight float64 = 64.0
	// Player Width
	PlayerWidth float64 = 64.0
	// Player Starting X
	PlayerStartingX float64 = 640.0 - PlayerWidth/2
	// Player Starting Y
	PlayerStartingY float64 = 240.0 - PlayerHeight/2
	// PlayerGravityMultiplier
	PlayerGravityMultiplier float64 = 300.0
	// PlayerHitpoints is the amount of hitpoints a player has
	PlayerHitpoints int16 = 100

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
	// NPCGravityMultiplier
	NPCGravityMultiplier float64 = 300.0
	// NPCRespawnTime is the time it takes for an NPC to respawn
	NPCRespawnTime float64 = 10.0
	// NPC Hitpoints
	NPCHitpoints int16 = 100
	// NPC Line of Sight
	NPCLineOfSight float64 = 320.0
	// NPC Attack Range
	NPCAttackRange float64 = 48.0

	// NPCAttack1Duration is the duration of the attack (channel time + cooldown time)
	NPCAttack1Duration float64 = 1.4 // seconds
	// NPCAttack1ChannelTime is the time it takes for the attack to register
	NPCAttack1ChannelTime float64 = 0.7 // seconds
	// NPCAttack1HitboxWidth is the width of the attack hitbox
	NPCAttack1HitboxWidth float64 = NPCWidth
	// NPCAttack1HitboxOffset is the offset from the npc's position to the attack hitbox
	NPCAttack1HitboxOffset float64 = NPCWidth / 2
	// NPCAttack1Damage is the amount of damage a npc does
	NPCAttack1Damage int16 = 30
)
