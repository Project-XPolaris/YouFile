package api

import (
	"errors"
	"github.com/project-xpolaris/youplustoolkit/youlink"
	"github.com/sirupsen/logrus"
	"youfile/service"
)

var youLinkMoveFileFunction = youlink.Function{
	Endpoint: "movefile",
	Name:     "YouFileMoveFile",
	Desc:     "Move file",
	Template: youlink.TemplateTypeHttpCall,
	InputDefinitions: []*youlink.VariableDefinition{
		{
			Name: "source",
			Type: youlink.VariableTypeString,
			Desc: "move source",
		},
		{
			Name: "target",
			Type: youlink.VariableTypeString,
			Desc: "move to",
		},
	},
	OutputDefinitions: []*youlink.VariableDefinition{},
	HandlerFunc: func(function *youlink.Function) error {
		source := function.GetInputString("source")
		if len(source) == 0 {
			return errors.New("source not found")
		}
		target := function.GetInputString("target")
		if len(source) == 0 {
			return errors.New("target not found")
		}
		go func() {
			err := service.Move(source, target, nil, "overwrite")
			if err != nil {
				logrus.Error(err)
				function.CallbackFunc(nil, err)
			}
			function.CallbackFunc(nil, nil)
		}()
		return nil
	},
}
