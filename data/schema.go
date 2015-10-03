package data

import (
	"github.com/chris-ramon/graphql-go/types"
	"github.com/sogko/graphql-relay-go"
)

var todoType *types.GraphQLObjectType
var userType *types.GraphQLObjectType

var nodeDefinitions *gqlrelay.NodeDefinitions
var todosConnection *gqlrelay.GraphQLConnectionDefinitions

var Schema types.GraphQLSchema

func init() {

	nodeDefinitions = gqlrelay.NewNodeDefinitions(gqlrelay.NodeDefinitionsConfig{
		IdFetcher: func(id string, info types.GraphQLResolveInfo) interface{} {
			resolvedId := gqlrelay.FromGlobalId(id)
			if resolvedId.Type == "Todo" {
				return GetTodo(resolvedId.Id)
			}
			if resolvedId.Type == "User" {
				return GetUser(resolvedId.Id)
			}
			return nil
		},
		TypeResolve: func(value interface{}, info types.GraphQLResolveInfo) *types.GraphQLObjectType {
			switch value.(type) {
			case *Todo:
				return todoType
			case *User:
				return userType
			default:
				return userType
			}
		},
	})

	todoType = types.NewGraphQLObjectType(types.GraphQLObjectTypeConfig{
		Name: "Todo",
		Fields: types.GraphQLFieldConfigMap{
			"id": gqlrelay.GlobalIdField("Todo", nil),
			"text": &types.GraphQLFieldConfig{
				Type: types.GraphQLString,
			},
			"complete": &types.GraphQLFieldConfig{
				Type: types.GraphQLBoolean,
			},
		},
		Interfaces: []*types.GraphQLInterfaceType{nodeDefinitions.NodeInterface},
	})

	todosConnection = gqlrelay.ConnectionDefinitions(gqlrelay.ConnectionConfig{
		Name:     "Todo",
		NodeType: todoType,
	})

	userType = types.NewGraphQLObjectType(types.GraphQLObjectTypeConfig{
		Name: "User",
		Fields: types.GraphQLFieldConfigMap{
			"id": gqlrelay.GlobalIdField("User", nil),
			"todos": &types.GraphQLFieldConfig{
				Type: todosConnection.ConnectionType,
				Args: gqlrelay.NewConnectionArgs(types.GraphQLFieldConfigArgumentMap{
					"status": &types.GraphQLArgumentConfig{
						Type:         types.GraphQLString,
						DefaultValue: "any",
					},
				}),
				Resolve: func(p types.GQLFRParams) interface{} {
					status, _ := p.Args["status"].(string)
					args := gqlrelay.NewConnectionArguments(p.Args)
					todos := TodosToSliceInterface(GetTodos(status))
					return gqlrelay.ConnectionFromArray(todos, args)
				},
			},
			"totalCount": &types.GraphQLFieldConfig{
				Type: types.GraphQLInt,
				Resolve: func(p types.GQLFRParams) interface{} {
					return len(GetTodos("any"))
				},
			},
			"completedCount": &types.GraphQLFieldConfig{
				Type: types.GraphQLInt,
				Resolve: func(p types.GQLFRParams) interface{} {
					return len(GetTodos("completed"))
				},
			},
		},
		Interfaces: []*types.GraphQLInterfaceType{nodeDefinitions.NodeInterface},
	})

	rootType := types.NewGraphQLObjectType(types.GraphQLObjectTypeConfig{
		Name: "Root",
		Fields: types.GraphQLFieldConfigMap{
			"viewer": &types.GraphQLFieldConfig{
				Type: userType,
				Resolve: func(p types.GQLFRParams) interface{} {
					return GetViewer()
				},
			},
			"node": nodeDefinitions.NodeField,
		},
	})

	addTodoMutation := gqlrelay.MutationWithClientMutationId(gqlrelay.MutationConfig{
		Name: "AddTodo",
		InputFields: types.InputObjectConfigFieldMap{
			"text": &types.InputObjectFieldConfig{
				Type: types.NewGraphQLNonNull(types.GraphQLString),
			},
		},
		OutputFields: types.GraphQLFieldConfigMap{
			"todoEdge": &types.GraphQLFieldConfig{
				Type: todosConnection.EdgeType,
				Resolve: func(p types.GQLFRParams) interface{} {
					payload, _ := p.Source.(map[string]interface{})
					todoId, _ := payload["todoId"].(string)
					todo := GetTodo(todoId)
					return gqlrelay.EdgeType{
						Node:   todo,
						Cursor: gqlrelay.CursorForObjectInConnection(TodosToSliceInterface(GetTodos("any")), todo),
					}
				},
			},
			"viewer": &types.GraphQLFieldConfig{
				Type: userType,
				Resolve: func(p types.GQLFRParams) interface{} {
					return GetViewer()
				},
			},
		},
		MutateAndGetPayload: func(inputMap map[string]interface{}, info types.GraphQLResolveInfo) map[string]interface{} {
			text, _ := inputMap["text"].(string)
			todoId := AddTodo(text, false)
			return map[string]interface{}{
				"todoId": todoId,
			}
		},
	})

	changeTodoStatusMutation := gqlrelay.MutationWithClientMutationId(gqlrelay.MutationConfig{
		Name: "ChangeTodoStatus",
		InputFields: types.InputObjectConfigFieldMap{
			"id": &types.InputObjectFieldConfig{
				Type: types.NewGraphQLNonNull(types.GraphQLID),
			},
			"complete": &types.InputObjectFieldConfig{
				Type: types.NewGraphQLNonNull(types.GraphQLBoolean),
			},
		},
		OutputFields: types.GraphQLFieldConfigMap{
			"todo": &types.GraphQLFieldConfig{
				Type: todoType,
				Resolve: func(p types.GQLFRParams) interface{} {
					payload, _ := p.Source.(map[string]interface{})
					todoId, _ := payload["todoId"].(string)
					todo := GetTodo(todoId)
					return todo
				},
			},
			"viewer": &types.GraphQLFieldConfig{
				Type: userType,
				Resolve: func(p types.GQLFRParams) interface{} {
					return GetViewer()
				},
			},
		},
		MutateAndGetPayload: func(inputMap map[string]interface{}, info types.GraphQLResolveInfo) map[string]interface{} {
			id, _ := inputMap["id"].(string)
			complete, _ := inputMap["complete"].(bool)
			resolvedId := gqlrelay.FromGlobalId(id)
			ChangeTodoStatus(resolvedId.Id, complete)
			return map[string]interface{}{
				"todoId": resolvedId.Id,
			}
		},
	})

	markAllTodosMutation := gqlrelay.MutationWithClientMutationId(gqlrelay.MutationConfig{
		Name: "MarkAllTodos",
		InputFields: types.InputObjectConfigFieldMap{
			"complete": &types.InputObjectFieldConfig{
				Type: types.NewGraphQLNonNull(types.GraphQLBoolean),
			},
		},
		OutputFields: types.GraphQLFieldConfigMap{
			"changedTodos": &types.GraphQLFieldConfig{
				Type: types.NewGraphQLList(todoType),
				Resolve: func(p types.GQLFRParams) interface{} {
					payload, _ := p.Source.(map[string]interface{})
					todoIds, _ := payload["todoIds"].([]string)
					todos := []*Todo{}
					for _, todoId := range todoIds {
						todo := GetTodo(todoId)
						if todo != nil {
							todos = append(todos, todo)
						}
					}
					return todos
				},
			},
			"viewer": &types.GraphQLFieldConfig{
				Type: userType,
				Resolve: func(p types.GQLFRParams) interface{} {
					return GetViewer()
				},
			},
		},
		MutateAndGetPayload: func(inputMap map[string]interface{}, info types.GraphQLResolveInfo) map[string]interface{} {
			complete, _ := inputMap["complete"].(bool)
			todoIds := MarkAllTodos(complete)
			return map[string]interface{}{
				"todoIds": todoIds,
			}
		},
	})

	removeCompletedTodosMutation := gqlrelay.MutationWithClientMutationId(gqlrelay.MutationConfig{
		Name: "RemoveCompletedTodos",
		OutputFields: types.GraphQLFieldConfigMap{
			"deletedTodoIds": &types.GraphQLFieldConfig{
				Type: types.NewGraphQLList(types.GraphQLString),
				Resolve: func(p types.GQLFRParams) interface{} {
					payload, _ := p.Source.(map[string]interface{})
					return payload["todoIds"]
				},
			},
			"viewer": &types.GraphQLFieldConfig{
				Type: userType,
				Resolve: func(p types.GQLFRParams) interface{} {
					return GetViewer()
				},
			},
		},
		MutateAndGetPayload: func(inputMap map[string]interface{}, info types.GraphQLResolveInfo) map[string]interface{} {
			todoIds := RemoveCompletedTodos()
			return map[string]interface{}{
				"todoIds": todoIds,
			}
		},
	})

	removeTodoMutation := gqlrelay.MutationWithClientMutationId(gqlrelay.MutationConfig{
		Name: "RemoveTodo",
		InputFields: types.InputObjectConfigFieldMap{
			"id": &types.InputObjectFieldConfig{
				Type: types.NewGraphQLNonNull(types.GraphQLID),
			},
		},
		OutputFields: types.GraphQLFieldConfigMap{
			"deletedTodoId": &types.GraphQLFieldConfig{
				Type: types.GraphQLID,
				Resolve: func(p types.GQLFRParams) interface{} {
					payload, _ := p.Source.(map[string]interface{})
					return payload["todoId"]
				},
			},
			"viewer": &types.GraphQLFieldConfig{
				Type: userType,
				Resolve: func(p types.GQLFRParams) interface{} {
					return GetViewer()
				},
			},
		},
		MutateAndGetPayload: func(inputMap map[string]interface{}, info types.GraphQLResolveInfo) map[string]interface{} {
			id, _ := inputMap["id"].(string)
			resolvedId := gqlrelay.FromGlobalId(id)
			RemoveTodo(resolvedId.Id)
			return map[string]interface{}{
				"todoId": resolvedId.Id,
			}
		},
	})
	renameTodoMutation := gqlrelay.MutationWithClientMutationId(gqlrelay.MutationConfig{
		Name: "RenameTodo",
		InputFields: types.InputObjectConfigFieldMap{
			"id": &types.InputObjectFieldConfig{
				Type: types.NewGraphQLNonNull(types.GraphQLID),
			},
			"text": &types.InputObjectFieldConfig{
				Type: types.NewGraphQLNonNull(types.GraphQLString),
			},
		},
		OutputFields: types.GraphQLFieldConfigMap{
			"todo": &types.GraphQLFieldConfig{
				Type: todoType,
				Resolve: func(p types.GQLFRParams) interface{} {
					payload, _ := p.Source.(map[string]interface{})
					todoId, _ := payload["todoId"].(string)
					return GetTodo(todoId)
				},
			},
			"viewer": &types.GraphQLFieldConfig{
				Type: userType,
				Resolve: func(p types.GQLFRParams) interface{} {
					return GetViewer()
				},
			},
		},
		MutateAndGetPayload: func(inputMap map[string]interface{}, info types.GraphQLResolveInfo) map[string]interface{} {
			id, _ := inputMap["id"].(string)
			resolvedId := gqlrelay.FromGlobalId(id)
			text, _ := inputMap["text"].(string)
			RenameTodo(resolvedId.Id, text)
			return map[string]interface{}{
				"todoId": resolvedId.Id,
			}
		},
	})
	mutationType := types.NewGraphQLObjectType(types.GraphQLObjectTypeConfig{
		Name: "Mutation",
		Fields: types.GraphQLFieldConfigMap{
			"addTodo":              addTodoMutation,
			"changeTodoStatus":     changeTodoStatusMutation,
			"markAllTodos":         markAllTodosMutation,
			"removeCompletedTodos": removeCompletedTodosMutation,
			"removeTodo":           removeTodoMutation,
			"renameTodo":           renameTodoMutation,
		},
	})

	var err error
	Schema, err = types.NewGraphQLSchema(types.GraphQLSchemaConfig{
		Query:    rootType,
		Mutation: mutationType,
	})
	if err != nil {
		panic(err)
	}
}
