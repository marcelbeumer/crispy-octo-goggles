package game

type World interface{}

type Message struct {
	Sender string
	Data   any
}

type WorldState struct {
	Players map[string]PlayerState
}

type PlayerState struct {
	Speed int
	Loc   int
}

type Player interface {
	Uuid() string
	Chan(in *<-chan Message, out *chan<- Message)
}
