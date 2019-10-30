package switcher

import "time"

type AutoSwitcher struct {
	sw      *Switcher
	timeout uint
	cancel  chan interface{}
}

func NewAutoSwitcher(sw *Switcher, minutes uint) *AutoSwitcher {
	return &AutoSwitcher{
		sw:      sw,
		timeout: minutes,
		cancel:  make(chan interface{}),
	}
}

func (as *AutoSwitcher) SetTimeout(minutes uint) {
	as.timeout = minutes
	as.Stop()
	as.Start()
}

func (as *AutoSwitcher) Start() {
	if as.timeout == 0 {
		return
	}
	go func() {
		ticker := time.NewTicker(time.Duration(as.timeout) * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-as.cancel:

				return
			case <-ticker.C:
				as.sw.Switch()
			}
		}
	}()
}

func (as *AutoSwitcher) Stop() {
	close(as.cancel)
	as.cancel = make(chan interface{})
}
