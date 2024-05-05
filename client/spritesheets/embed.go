package spritesheets

import _ "embed"

//go:embed player/swordsman/Idle.png
var PlayerSwordsmanIdle []byte

//go:embed player/swordsman/Run.png
var PlayerSwordsmanRun []byte

//go:embed player/swordsman/Jump.png
var PlayerSwordsmanJump []byte

//go:embed player/swordsman/Attack_1.png
var PlayerSwordsmanAttack1 []byte

//go:embed player/swordsman/Attack_2.png
var PlayerSwordsmanAttack2 []byte

//go:embed player/swordsman/Attack_3.png
var PlayerSwordsmanAttack3 []byte

//go:embed skeleton/Idle.png
var SkeletonIdle []byte

//go:embed skeleton/Walk.png
var SkeletonWalk []byte

//go:embed skeleton/Dead.png
var SkeletonDead []byte

//go:embed skeleton/Attack_1.png
var SkeletonAttack1 []byte
