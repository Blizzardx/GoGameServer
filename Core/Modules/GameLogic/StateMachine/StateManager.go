package StateMachine

type StateManager struct {
	allStatus     []*stateLogic
	stateMgr      IStateManager
	currentStatus *stateLogic
}

//----------------public interface--------------------------
func (stateManager *StateManager) Init(stateMgr IStateManager, allStatus []IState) {
	for _, state := range allStatus {
		stateManager.allStatus = append(stateManager.allStatus, createState(state))
	}
	stateManager.stateMgr = stateMgr
	stateManager.currentStatus = nil
	stateManager.stateMgr.OnInit()
}
func (stateManager *StateManager) ChangeStatus(newStatusType int32) {
	// check state type
	newStatus := stateManager.getStatusByType(newStatusType)
	if newStatus == nil {
		log.Errorln("cant't change status ,not found state by type ", newStatusType)
		return
	}
	stateManager.stateMgr.OnChangeStatus_Before(stateManager.currentStatus.logic, newStatus.logic)

	if nil != stateManager.currentStatus {
		stateManager.currentStatus.exit()
	}
	stateManager.currentStatus = newStatus
	stateManager.currentStatus.enter()
	stateManager.stateMgr.OnChangeStatus_Before(stateManager.currentStatus.logic, newStatus.logic)
}
func (stateManager *StateManager) Tick() {
	stateManager.stateMgr.OnTick_Before()
	if nil != stateManager.currentStatus {
		stateManager.currentStatus.tick()
	}
	stateManager.stateMgr.OnTick_After()
}

//----------------system function--------------------------
func (stateManager *StateManager) getStatusByType(stateType int32) *stateLogic {
	for _, state := range stateManager.allStatus {
		if state.logic.GetStatusType() == stateType {
			return state
		}
	}
	return nil
}
