package StateMachine

type StateManager struct {
	allStatus     []*stateLogic
	stateMgr      IStateManager
	currentStatus *stateLogic
}

//----------------public interface--------------------------
func (stateManager *StateManager) Init(stateMgr IStateManager, allStatus []IState) {
	for _, state := range allStatus {
		if state == nil {
			log.Errorln("state can't not be null")
			continue
		}
		if stateManager.getStatusByType(state.GetStatusType()) != nil {
			log.Errorln("state is already in list ", state.GetStatusType())
			continue
		}
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
	if stateManager.currentStatus != nil && newStatus.logic.GetStatusType() == stateManager.currentStatus.logic.GetStatusType() {
		return
	}
	if nil == stateManager.currentStatus {
		stateManager.stateMgr.OnChangeStatus_Before(nil, newStatus.logic)
	} else {
		stateManager.stateMgr.OnChangeStatus_Before(stateManager.currentStatus.logic, newStatus.logic)
	}

	if nil != stateManager.currentStatus {
		stateManager.currentStatus.exit()
	}
	stateManager.currentStatus = newStatus
	stateManager.currentStatus.enter()
	stateManager.stateMgr.OnChangeStatus_After(stateManager.currentStatus.logic, newStatus.logic)
}
func (stateManager *StateManager) Tick() {
	stateManager.stateMgr.OnTick_Before()
	if nil != stateManager.currentStatus {
		stateManager.currentStatus.tick()
	}
	stateManager.stateMgr.OnTick_After()
}
func (stateManager *StateManager) GetCurrentState() IState {
	return stateManager.currentStatus.logic
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
