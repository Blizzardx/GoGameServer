package StateMachine

type IState interface {
	OnInit()
	OnEnter()
	OnTick()
	OnExit()
	GetStatusType() int32
}
