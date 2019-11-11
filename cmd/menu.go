package main

import "github.com/lxn/walk"

type handlerProvider func(*walk.Action) walk.EventHandler

// addNewAction adds new action to ActionList
func addNewAction(name string, actions *walk.ActionList, hp handlerProvider) (*walk.Action, error) {
	action := walk.NewAction()
	if err := action.SetText(name); err != nil {
		return nil, err
	}
	action.Triggered().Attach(hp(action))
	if err := actions.Add(action); err != nil {
		return nil, err
	}

	return action, nil
}

// addNewCheckableAction adds new action with checkbox
func addNewCheckableAction(name string, actions *walk.ActionList, init bool,
	check walk.EventHandler, uncheck walk.EventHandler) (*walk.Action, error) {
	hp := func(action *walk.Action) walk.EventHandler {
		return func() {
			if action.Checked() {
				check()
				return
			}
			uncheck()
		}
	}
	action, err := addNewAction(name, actions, hp)
	if err != nil {
		return nil, err
	}
	if err = action.SetCheckable(true); err != nil {
		return nil, err
	}
	if err = action.SetChecked(init); err != nil {
		return nil, err
	}
	return action, nil
}

// addNewRadioAction adds new enabled radio-like action to ActionList
func addNewRadioAction(name string, actions *walk.ActionList,
	check walk.EventHandler, uncheck walk.EventHandler) (*walk.Action, error) {
	action, err := addNewCheckableAction(name, actions, false, check, uncheck)
	if err != nil {
		return nil, err
	}
	action.Triggered().Attach(func() {
		switchAsRadio(actions, action)
	})

	return action, nil
}

func switchAsRadio(actions *walk.ActionList, action *walk.Action) {
	l := actions.Len()
	for i := 0; i < l; i++ {
		if actions.At(i) == action {
			action.SetChecked(true)
			continue
		}
		actions.At(i).SetChecked(false)
	}
}
