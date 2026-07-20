package core

import "github.com/runtime-terror404/pio-scaffold/internal/ioc"

// Registry holds fields genuinely shared by both platforms.
type Registry interface {
	Key() PlatformKey
	DebugProbes() map[string]DebugProbe // empty for pico2, populated for stm32 — never nil
	Frameworks() []string
}

// StaticBoardRegistry is implemented by platforms with a real, enumerable
// board catalog (Pico2).
type StaticBoardRegistry interface {
	Registry
	Boards() map[string]Board
}

// IOCDerivedRegistry is implemented by platforms that synthesize exactly one
// board from a parsed .ioc file, and can fail doing so (STM32).
type IOCDerivedRegistry interface {
	Registry
	BoardFromIOC(parsed ioc.Parsed) (Board, error)
}
