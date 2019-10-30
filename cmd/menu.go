package main

import "github.com/lxn/walk"

// addNewAction adds new action to ActionList
func addNewAction(name string, actions *walk.ActionList, handler walk.EventHandler) (*walk.Action, error) {
	action := walk.NewAction()
	if err := action.SetText(name); err != nil {
		return nil, err
	}
	action.Triggered().Attach(func() { handler() })
	if err := actions.Add(action); err != nil {
		return nil, err
	}

	return action, nil
}

// addNewRadioAction adds new enabled radio-logic action to ActionList
func addNewRadioAction(name string, actions *walk.ActionList, handler walk.EventHandler) (*walk.Action, error) {
	action, err := addNewAction(name, actions, handler)
	if err != nil {
		return nil, err
	}
	if err = action.SetCheckable(true); err != nil {
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
