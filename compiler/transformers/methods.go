package transformers

import (
	"biscuit/compiler/ast"
	"biscuit/compiler/lexer"
)

// createInitMethod creates the __init__ method for the view class
func (vm *ViewTransformer) createInitMethod(viewStmt *ast.ViewStmt) (*ast.Function, error) {
	// Create __init__ method name
	initName := &ast.Name{
		Token: lexer.Token{
			Lexeme: "__init__",
			Type:   lexer.Identifier,
		},
		Span: viewStmt.Span,
	}

	// Create 'self' parameter
	selfParam := &ast.Parameter{
		Name: &ast.Name{
			Token: lexer.Token{
				Lexeme: "self",
				Type:   lexer.Identifier,
			},
			Span: viewStmt.Span,
		},
		Default:    nil,
		Annotation: nil,
		IsStar:     false,
		Span:       viewStmt.Span,
	}

	// Start with self parameter
	initParams := []*ast.Parameter{selfParam}

	// Add view parameters to __init__
	if viewStmt.Params != nil && len(viewStmt.Params.Parameters) > 0 {
		for _, param := range viewStmt.Params.Parameters {
			// Skip parameters with nil names
			if param == nil || param.Name == nil {
				continue
			}
			initParams = append(initParams, param)
		}
	}

	// Add slot parameters if we have slots
	if len(vm.slots) > 0 {
		// Add star parameter to make slot parameters keyword-only
		starParam := &ast.Parameter{
			Name:       nil, // Unnamed star parameter
			Default:    nil,
			Annotation: nil,
			IsStar:     true,
			Span:       viewStmt.Span,
		}
		initParams = append(initParams, starParam)

		// Add default slot parameter first (children)
		if _, hasDefaultSlot := vm.slots[""]; hasDefaultSlot {
			childrenParam := &ast.Parameter{
				Name: &ast.Name{
					Token: lexer.Token{
						Lexeme: "children",
						Type:   lexer.Identifier,
					},
					Span: viewStmt.Span,
				},
				Default: &ast.Literal{
					Type:  ast.LiteralTypeNone,
					Value: nil,
					Span:  viewStmt.Span,
				},
				Annotation: vm.createSlotTypeAnnotation(),
				IsStar:     false,
				Span:       viewStmt.Span,
			}
			initParams = append(initParams, childrenParam)
		}

		// Add named slot parameters in order
		for _, slotName := range vm.slotOrder {
			if slotName != "" { // Skip default slot (already added as children)
				slotParam := &ast.Parameter{
					Name: &ast.Name{
						Token: lexer.Token{
							Lexeme: slotName,
							Type:   lexer.Identifier,
						},
						Span: viewStmt.Span,
					},
					Default: &ast.Literal{
						Type:  ast.LiteralTypeNone,
						Value: nil,
						Span:  viewStmt.Span,
					},
					Annotation: vm.createSlotTypeAnnotation(),
					IsStar:     false,
					Span:       viewStmt.Span,
				}
				initParams = append(initParams, slotParam)
			}
		}
	}

	// Create parameter list
	paramList := &ast.ParameterList{
		Parameters:  initParams,
		SlashIndex:  -1,
		VarArgIndex: -1,
		KwArgIndex:  -1,
		Span:        viewStmt.Span,
	}

	// Create assignment statements for each view parameter
	var initBody []ast.Stmt

	// Add super().__init__() call as the first statement
	superCall := &ast.ExprStmt{
		Expr: &ast.Call{
			Callee: &ast.Attribute{
				Object: &ast.Call{
					Callee: &ast.Name{
						Token: lexer.Token{Lexeme: "super", Type: lexer.Identifier},
						Span:  viewStmt.Span,
					},
					Arguments: []*ast.Argument{},
					Span:      viewStmt.Span,
				},
				Name: lexer.Token{Lexeme: "__init__", Type: lexer.Identifier},
				Span: viewStmt.Span,
			},
			Arguments: []*ast.Argument{},
			Span:      viewStmt.Span,
		},
		Span: viewStmt.Span,
	}
	initBody = append(initBody, superCall)

	// Add view parameter assignments
	if viewStmt.Params != nil && len(viewStmt.Params.Parameters) > 0 {
		for _, param := range viewStmt.Params.Parameters {
			// Skip parameters with nil names
			if param == nil || param.Name == nil {
				continue
			}

			// Create self.param_name = param_name
			selfAttr := &ast.Attribute{
				Object: &ast.Name{
					Token: lexer.Token{Lexeme: "self", Type: lexer.Identifier},
					Span:  viewStmt.Span,
				},
				Name: param.Name.Token,
				Span: viewStmt.Span,
			}

			assignment := &ast.AssignStmt{
				Targets: []ast.Expr{selfAttr},
				Value:   param.Name,
				Span:    viewStmt.Span,
			}

			initBody = append(initBody, assignment)
		}
	}

	// Add slot parameter assignments
	if len(vm.slots) > 0 {
		// Add default slot assignment (children)
		if _, hasDefaultSlot := vm.slots[""]; hasDefaultSlot {
			selfChildren := &ast.Attribute{
				Object: &ast.Name{
					Token: lexer.Token{Lexeme: "self", Type: lexer.Identifier},
					Span:  viewStmt.Span,
				},
				Name: lexer.Token{Lexeme: "children", Type: lexer.Identifier},
				Span: viewStmt.Span,
			}
			assignment := &ast.AssignStmt{
				Targets: []ast.Expr{selfChildren},
				Value: &ast.Name{
					Token: lexer.Token{Lexeme: "children", Type: lexer.Identifier},
					Span:  viewStmt.Span,
				},
				Span: viewStmt.Span,
			}
			initBody = append(initBody, assignment)
		}

		// Add named slot assignments in order
		for _, slotName := range vm.slotOrder {
			if slotName != "" { // Skip default slot
				selfSlot := &ast.Attribute{
					Object: &ast.Name{
						Token: lexer.Token{Lexeme: "self", Type: lexer.Identifier},
						Span:  viewStmt.Span,
					},
					Name: lexer.Token{Lexeme: slotName, Type: lexer.Identifier},
					Span: viewStmt.Span,
				}
				assignment := &ast.AssignStmt{
					Targets: []ast.Expr{selfSlot},
					Value: &ast.Name{
						Token: lexer.Token{Lexeme: slotName, Type: lexer.Identifier},
						Span:  viewStmt.Span,
					},
					Span: viewStmt.Span,
				}
				initBody = append(initBody, assignment)
			}
		}
	}

	// If no body, add pass statement
	if len(initBody) == 0 {
		passStmt := &ast.PassStmt{
			Span: viewStmt.Span,
		}
		initBody = append(initBody, passStmt)
	}

	return &ast.Function{
		Name:           initName,
		TypeParameters: nil,
		Parameters:     paramList,
		ReturnType:     nil,
		Body:           initBody,
		IsAsync:        false,
		Span:           viewStmt.Span,
	}, nil
}

// createRenderMethod creates the _render method for the view class
func (vm *ViewTransformer) createRenderMethod(viewStmt *ast.ViewStmt) (*ast.Function, error) {
	// Create _render method name
	renderName := &ast.Name{
		Token: lexer.Token{
			Lexeme: "_render",
			Type:   lexer.Identifier,
		},
		Span: viewStmt.Span,
	}

	// Create 'self' parameter
	selfParam := &ast.Parameter{
		Name: &ast.Name{
			Token: lexer.Token{
				Lexeme: "self",
				Type:   lexer.Identifier,
			},
			Span: viewStmt.Span,
		},
		Default:    nil,
		Annotation: nil,
		IsStar:     false,
		Span:       viewStmt.Span,
	}

	// Create parameter list with just self
	paramList := &ast.ParameterList{
		Parameters:  []*ast.Parameter{selfParam},
		SlashIndex:  -1,
		VarArgIndex: -1,
		KwArgIndex:  -1,
		Span:        viewStmt.Span,
	}

	// Create return type annotation: Element
	elementType := &ast.Name{
		Token: lexer.Token{
			Lexeme: "Element",
			Type:   lexer.Identifier,
		},
		Span: viewStmt.Span,
	}

	// Transform view body into _render method body
	renderBody, err := vm.transformViewBody(viewStmt.Body)
	if err != nil {
		return nil, err
	}

	return &ast.Function{
		Name:           renderName,
		TypeParameters: nil,
		Parameters:     paramList,
		ReturnType:     elementType,
		Body:           renderBody,
		IsAsync:        false,
		Span:           viewStmt.Span,
	}, nil
}