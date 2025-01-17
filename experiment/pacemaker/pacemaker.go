package pacemaker

import (
	"sync"
	"time"

	"github.com/gitferry/bamboo/config"

	"github.com/gitferry/bamboo/types"
)

type Pacemaker struct {
	curView              types.View
	newViewChan          chan types.View
	timeoutController    *TimeoutController
	viewChangeController *ViewChangeController
	mu                   sync.Mutex
}

func NewPacemaker(n int) *Pacemaker {
	pm := new(Pacemaker)
	pm.newViewChan = make(chan types.View, 100)
	pm.timeoutController = NewTimeoutController(n)
	pm.viewChangeController = NewViewChangeController(n)
	return pm
}

func (p *Pacemaker) ProcessRemoteTmo(tmo *TMO) (bool, *TC) {
	if tmo.View < p.curView {
		return false, nil
	}
	return p.timeoutController.AddTmo(tmo)
}

func (p *Pacemaker) ProcessRemoteVmo(vmo *VMO) (bool, *VC) {
	if vmo.View < p.curView {
		return false, nil
	}
	return p.viewChangeController.AddVmo(vmo)
}

func (p *Pacemaker) AdvanceView(view types.View) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if view < p.curView {
		return
	}
	p.curView = view + 1
	p.newViewChan <- view + 1 // reset timer for the next view
}

func (p *Pacemaker) EnteringViewEvent() chan types.View {
	return p.newViewChan
}

func (p *Pacemaker) GetCurView() types.View {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.curView
}

func (p *Pacemaker) GetTimerForView() time.Duration {
	return time.Duration(config.GetConfig().Timeout) * time.Millisecond
}
